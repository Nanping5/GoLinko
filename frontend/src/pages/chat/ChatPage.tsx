import React, { useState, useEffect, useMemo, useRef } from 'react';
import { useAuthStore } from '../../store/useAuthStore';
import { MessageSquare, Users, Settings, Search, Send, Plus, Paperclip, Loader2, Inbox, ArrowLeft, X, ChevronRight, User, CircleUser, Copy, Check, UserX, Info, Phone, Video, Shield, UserCog, Ban, Mic, Play, Pause } from 'lucide-react';
import { groupApi, type GroupInfo } from '../../api/group.api';
import { contactApi, type ContactApply, type ContactUserInfo } from '../../api/contact.api';
import { messageApi } from '../../api/message.api';
import { adminApi, type AdminUserItem } from '../../api/admin.api';
import { userApi } from '../../api/user.api';
import { useWebSocket } from '../../hooks/useWebSocket';
import { CreateGroupModal } from './CreateGroupModal';
import { SearchGroupModal } from './SearchGroupModal';
import { GroupDetailsModal } from './GroupDetailsModal';
import { ProfileModal } from './ProfileModal';
import { ConfirmDialog } from '../../components/common/ConfirmDialog';
import type { ChatMessage, SessionItem } from '../../types';
import { getApiBaseUrl, getWsBaseUrl, getWebRTCIceServers } from '../../utils/runtime';

// 根据字符串生成确定性头像背景色
const AVATAR_COLORS = [
  'bg-emerald-500','bg-sky-500','bg-violet-500','bg-amber-500',
  'bg-rose-500','bg-teal-500','bg-indigo-500','bg-orange-500',
];
const avatarBg = (s: string) => {
  let h = 0;
  for (let i = 0; i < s.length; i++) h = s.charCodeAt(i) + ((h << 5) - h);
  return AVATAR_COLORS[Math.abs(h) % AVATAR_COLORS.length];
};

const BASE_NAV_ITEMS = [
  { key: 'sessions', label: '消息',  Icon: MessageSquare },
  { key: 'groups',   label: '群组',  Icon: Users },
  { key: 'contacts', label: '联系人', Icon: User },
  { key: 'profile',  label: '我的',  Icon: CircleUser },
] as const;

const ADMIN_NAV_ITEM = { key: 'admin', label: '管理', Icon: Shield } as const;

const API_BASE_URL = getApiBaseUrl();
const WS_BASE_URL = getWsBaseUrl();
const WEBRTC_ICE_SERVERS = getWebRTCIceServers();

const turnUrlPriority = (url: string) => {
  const u = url.toLowerCase();
  if (u.startsWith('turns:') && u.includes(':443')) return 100;
  if (u.startsWith('turns:') && u.includes(':5349')) return 90;
  if (u.startsWith('turn:') && u.includes('transport=tcp') && u.includes(':3478')) return 80;
  if (u.startsWith('turn:') && u.includes('transport=tcp') && u.includes(':80')) return 70;
  if (u.startsWith('turn:') && u.includes('transport=udp')) return 60;
  if (u.startsWith('stun:')) return 10;
  return 1;
};

const normalizeIceServers = (servers: RTCIceServer[]) => {
  return servers.map((s) => {
    const urls = Array.isArray(s.urls) ? [...s.urls] : [s.urls as string];
    urls.sort((a, b) => turnUrlPriority(b) - turnUrlPriority(a));
    return { ...s, urls };
  });
};

/** 头像组件：有 URL 用 img，无则用字母占位符 */
const Av: React.FC<{ url?: string; name: string; cls: string }> = ({ url, name, cls }) => {
  const [err, setErr] = React.useState(false);
  const fullUrl = React.useMemo(() => {
    if (!url) return '';
    if (url.startsWith('data:')) return url;
    if (url.startsWith('http')) return url;
    const baseUrl = API_BASE_URL;
    return `${baseUrl}${url.startsWith('/') ? '' : '/'}${url}`;
  }, [url]);

  if (url && !err) {
    return <img src={fullUrl} alt={name} className={`${cls} object-cover`} onError={() => setErr(true)} />;
  }
  return (
    <div className={`${cls} flex items-center justify-center text-white font-bold ${avatarBg(name)}`}>
      {(name || '?')[0].toUpperCase()}
    </div>
  );
};

/** 全局图片预览组件 */
const ImagePreview: React.FC<{ url: string; onClose: () => void }> = ({ url, onClose }) => {
  return (
    <div className="fixed inset-0 z-[300] bg-black/90 flex flex-col items-center justify-center p-4" onClick={onClose}>
      <button 
        className="absolute top-6 right-6 p-2 bg-white/10 hover:bg-white/20 rounded-full text-white transition-colors"
        onClick={onClose}
      >
        <X className="w-6 h-6" />
      </button>
      <img 
        src={url} 
        alt="预览图片" 
        className="max-w-full max-h-full object-contain shadow-2xl rounded-lg animate-in zoom-in-95 duration-200"
        onClick={(e) => e.stopPropagation()} 
      />
      <div className="absolute bottom-10 flex gap-4">
        <a 
          href={url} 
          download 
          target="_blank" 
          rel="noreferrer"
          className="px-6 py-2 bg-white text-black rounded-full font-medium hover:bg-gray-100 transition-colors"
          onClick={(e) => e.stopPropagation()}
        >
          查看原图 / 下载
        </a>
      </div>
    </div>
  );
};

export const ChatPage: React.FC = () => {
  const { user, logout, updateUser } = useAuthStore();
  const [message, setMessage] = useState('');
  const [groups, setGroups] = useState<GroupInfo[]>([]);
  const [sessions, setSessions] = useState<SessionItem[]>([]);
  const [contacts, setContacts] = useState<ContactUserInfo[]>([]);
  const [applies, setApplies] = useState<ContactApply[]>([]);
  const [adminUsers, setAdminUsers] = useState<AdminUserItem[]>([]);
  const [loadingAdminUsers, setLoadingAdminUsers] = useState(false);
  const [loadingGroups, setLoadingGroups] = useState(true);
  const [loadingMessages, setLoadingMessages] = useState(false);
  const [activeTab, setActiveTab] = useState<'groups' | 'sessions' | 'contacts' | 'profile' | 'admin'>('sessions');
  const [activeChat, setActiveChat] = useState<SessionItem | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isSearchModalOpen, setIsSearchModalOpen] = useState(false);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [fileUploading, setFileUploading] = useState(false);
  const [recording, setRecording] = useState(false);
  const [recordSeconds, setRecordSeconds] = useState(0);
  const [recordWillCancel, setRecordWillCancel] = useState(false);
  const [playingVoiceUrl, setPlayingVoiceUrl] = useState<string | null>(null);
  const [voiceDurations, setVoiceDurations] = useState<Record<string, number>>({});
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const recordingStreamRef = useRef<MediaStream | null>(null);
  const recordingChunksRef = useRef<Blob[]>([]);
  const voiceAudioRefs = useRef<Record<string, HTMLAudioElement | null>>({});
  const recordTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const recordStartAtRef = useRef<number>(0);
  const recordPointerStartYRef = useRef<number>(0);
  const shouldSendRecordingRef = useRef<boolean>(true);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [toast, setToast] = useState<{ msg: string; type: 'success' | 'error' | 'info' } | null>(null);
  const toastTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const showToast = (msg: string, type: 'success' | 'error' | 'info' = 'success') => {
    if (toastTimer.current) clearTimeout(toastTimer.current);
    setToast({ msg, type });
    toastTimer.current = setTimeout(() => setToast(null), 2500);
  };
  const [mobileView, setMobileView] = useState<'list' | 'chat'>('list');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const [confirmDialog, setConfirmDialog] = useState<{
    title: string; message: string; confirmLabel?: string; danger?: boolean; onConfirm: () => void;
  } | null>(null);
  const [viewingContact, setViewingContact] = useState<ContactUserInfo | null>(null);
  const [viewingContactDetail, setViewingContactDetail] = useState<any>(null);
  const [contactsSubView, setContactsSubView] = useState<'list' | 'new-friends'>('list');
  const applyInputRef = useRef<HTMLInputElement>(null);
  const [previewImage, setPreviewImage] = useState<string | null>(null);

  const navItems = useMemo(
    () => (user?.isAdmin ? [...BASE_NAV_ITEMS, ADMIN_NAV_ITEM] : BASE_NAV_ITEMS),
    [user?.isAdmin]
  );

  type CallMode = 'audio' | 'video';
  type CallStatus = 'idle' | 'calling' | 'ringing' | 'in-call';
  type SignalPayload = {
    action: 'offer' | 'answer' | 'candidate' | 'hangup' | 'reject';
    callType?: CallMode;
    sdp?: RTCSessionDescriptionInit;
    candidate?: RTCIceCandidateInit;
  };

  const [callStatus, setCallStatus] = useState<CallStatus>('idle');
  const [callMode, setCallMode] = useState<CallMode | null>(null);
  const [callPeerId, setCallPeerId] = useState<string | null>(null);
  const [incomingOffer, setIncomingOffer] = useState<{ from: string; payload: SignalPayload } | null>(null);
  const localVideoRef = useRef<HTMLVideoElement | null>(null);
  const remoteVideoRef = useRef<HTMLVideoElement | null>(null);
  const remoteAudioRef = useRef<HTMLAudioElement | null>(null);
  const localStreamRef = useRef<MediaStream | null>(null);
  const remoteStreamRef = useRef<MediaStream | null>(null);
  const pcRef = useRef<RTCPeerConnection | null>(null);
  const pendingRemoteCandidatesRef = useRef<RTCIceCandidateInit[]>([]);
  const [callSeconds, setCallSeconds] = useState(0);
  const callTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const callMetaRef = useRef<{
    initiator: boolean;
    peerId: string | null;
    mode: CallMode | null;
    connectedAt: number | null;
    summarySent: boolean;
  }>({ initiator: false, peerId: null, mode: null, connectedAt: null, summarySent: false });
  const rtcIceServersRef = useRef<RTCIceServer[]>(WEBRTC_ICE_SERVERS);

  const askConfirm = (opts: typeof confirmDialog) => setConfirmDialog(opts);

  const formatCallDuration = (seconds: number) => {
    const m = Math.floor(seconds / 60).toString().padStart(2, '0');
    const s = (seconds % 60).toString().padStart(2, '0');
    return `${m}:${s}`;
  };

  const formatRecordDuration = (seconds: number) => {
    const m = Math.floor(seconds / 60).toString().padStart(2, '0');
    const s = (seconds % 60).toString().padStart(2, '0');
    return `${m}:${s}`;
  };

  const formatVoiceSeconds = (seconds?: number) => {
    const safe = Math.max(1, Math.round(seconds || 0));
    return `${safe}''`;
  };

  const toggleVoicePlayback = async (url: string) => {
    const target = voiceAudioRefs.current[url];
    if (!target) return;

    if (playingVoiceUrl && playingVoiceUrl !== url) {
      const current = voiceAudioRefs.current[playingVoiceUrl];
      if (current) {
        current.pause();
        current.currentTime = 0;
      }
    }

    if (playingVoiceUrl === url && !target.paused) {
      target.pause();
      setPlayingVoiceUrl(null);
      return;
    }

    try {
      await target.play();
      setPlayingVoiceUrl(url);
    } catch (e) {
      console.error('play voice failed', e);
    }
  };

  const sendSignal = (targetUserId: string, payload: SignalPayload) => {
    const privateSession = sessions.find((s) => s.type === 'private' && s.receiveId === targetUserId);
    sendMessage({
      session_id: privateSession?.sessionId || activeChat?.sessionId || '',
      type: 3,
      receive_id: targetUserId,
      av_data: JSON.stringify(payload),
    } as any);
  };

  const sendCallSummaryIfNeeded = () => {
    const meta = callMetaRef.current;
    if (!meta.initiator || meta.summarySent || !meta.connectedAt || !meta.peerId || !meta.mode || !user) return;

    const durationSec = Math.max(1, Math.floor((Date.now() - meta.connectedAt) / 1000));
    const privateSession = sessions.find((s) => s.type === 'private' && s.receiveId === meta.peerId);
    const sessionId = privateSession?.sessionId || (activeChat?.type === 'private' && activeChat.receiveId === meta.peerId ? activeChat.sessionId : '');
    if (!sessionId) return;
    const summaryText = `${meta.mode === 'video' ? '视频' : '语音'}通话 ${formatCallDuration(durationSec)}`;

    sendMessage({
      session_id: sessionId,
      type: 0,
      content: summaryText,
      send_name: user.nickName,
      send_avatar: user.avatar,
      receive_id: meta.peerId,
    } as any);

    callMetaRef.current = { ...meta, summarySent: true };
  };

  const cleanupCall = () => {
    if (callTimerRef.current) {
      clearInterval(callTimerRef.current);
      callTimerRef.current = null;
    }
    if (pcRef.current) {
      pcRef.current.ontrack = null;
      pcRef.current.onicecandidate = null;
      pcRef.current.close();
      pcRef.current = null;
    }
    if (localStreamRef.current) {
      localStreamRef.current.getTracks().forEach((t) => t.stop());
      localStreamRef.current = null;
    }
    pendingRemoteCandidatesRef.current = [];
    remoteStreamRef.current = null;
    if (localVideoRef.current) localVideoRef.current.srcObject = null;
    if (remoteVideoRef.current) remoteVideoRef.current.srcObject = null;
    if (remoteAudioRef.current) remoteAudioRef.current.srcObject = null;
    setCallSeconds(0);
    setCallStatus('idle');
    setCallMode(null);
    setCallPeerId(null);
    setIncomingOffer(null);
    callMetaRef.current = { initiator: false, peerId: null, mode: null, connectedAt: null, summarySent: false };
  };

  const finishCall = (needSummary: boolean) => {
    if (needSummary) {
      sendCallSummaryIfNeeded();
    }
    cleanupCall();
  };

  const ensureRtcIceServers = async () => {
    rtcIceServersRef.current = normalizeIceServers(WEBRTC_ICE_SERVERS);
    return rtcIceServersRef.current;
  };

  const flushPendingRemoteCandidates = async () => {
    const pc = pcRef.current;
    if (!pc || !pc.remoteDescription) return;
    if (!pendingRemoteCandidatesRef.current.length) return;

    const candidates = [...pendingRemoteCandidatesRef.current];
    pendingRemoteCandidatesRef.current = [];
    for (const candidate of candidates) {
      try {
        await pc.addIceCandidate(new RTCIceCandidate(candidate));
      } catch (error) {
        console.warn('flush remote candidate failed', error);
      }
    }
  };

  const createPeerConnection = (targetUserId: string, iceServers: RTCIceServer[]) => {
    const pc = new RTCPeerConnection({
      iceServers,
      iceTransportPolicy: 'all',
    });

    pc.oniceconnectionstatechange = () => {
      const state = pc.iceConnectionState;
      if (state === 'failed') {
        showToast('通话连接失败：跨网络通话通常需要配置 TURN 中继服务', 'error');
      }
    };

    pc.onicecandidate = (event) => {
      if (!event.candidate) return;
      sendSignal(targetUserId, { action: 'candidate', candidate: event.candidate.toJSON() });
    };

    pc.ontrack = (event) => {
      const [stream] = event.streams;
      if (!stream) return;
      remoteStreamRef.current = stream;
      if (remoteVideoRef.current) remoteVideoRef.current.srcObject = stream;
      if (remoteAudioRef.current) remoteAudioRef.current.srcObject = stream;
    };

    pcRef.current = pc;
    return pc;
  };

  const startCall = async (mode: CallMode) => {
    if (!activeChat || activeChat.type !== 'private') return;
    try {
      callMetaRef.current = {
        initiator: true,
        peerId: activeChat.receiveId,
        mode,
        connectedAt: null,
        summarySent: false,
      };
      const constraints: MediaStreamConstraints = mode === 'video'
        ? { audio: true, video: true }
        : { audio: true, video: false };
      const localStream = await navigator.mediaDevices.getUserMedia(constraints);
      localStreamRef.current = localStream;
      if (localVideoRef.current) localVideoRef.current.srcObject = localStream;

      const iceServers = await ensureRtcIceServers();
      const pc = createPeerConnection(activeChat.receiveId, iceServers);
      localStream.getTracks().forEach((track) => pc.addTrack(track, localStream));

      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
      sendSignal(activeChat.receiveId, { action: 'offer', callType: mode, sdp: offer });

      setCallStatus('calling');
      setCallMode(mode);
      setCallPeerId(activeChat.receiveId);
    } catch (e) {
      console.error('start call failed', e);
      showToast('无法获取麦克风/摄像头权限', 'error');
      cleanupCall();
    }
  };

  const acceptIncomingCall = async () => {
    if (!incomingOffer) return;
    try {
      const privateSession = sessions.find((s) => s.type === 'private' && s.receiveId === incomingOffer.from);
      if (privateSession) {
        setActiveTab('sessions');
        setActiveChat(privateSession);
        setMobileView('chat');
      }

      const mode = incomingOffer.payload.callType || 'audio';
      callMetaRef.current = {
        initiator: false,
        peerId: incomingOffer.from,
        mode,
        connectedAt: Date.now(),
        summarySent: false,
      };
      const constraints: MediaStreamConstraints = mode === 'video'
        ? { audio: true, video: true }
        : { audio: true, video: false };
      const localStream = await navigator.mediaDevices.getUserMedia(constraints);
      localStreamRef.current = localStream;
      if (localVideoRef.current) localVideoRef.current.srcObject = localStream;

      const iceServers = await ensureRtcIceServers();
      const pc = createPeerConnection(incomingOffer.from, iceServers);
      localStream.getTracks().forEach((track) => pc.addTrack(track, localStream));

      if (incomingOffer.payload.sdp) {
        await pc.setRemoteDescription(new RTCSessionDescription(incomingOffer.payload.sdp));
        await flushPendingRemoteCandidates();
      }
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);
      sendSignal(incomingOffer.from, { action: 'answer', sdp: answer });

      setCallStatus('in-call');
      setCallMode(mode);
      setCallPeerId(incomingOffer.from);
      setIncomingOffer(null);
    } catch (e) {
      console.error('accept call failed', e);
      showToast('接听失败', 'error');
      cleanupCall();
    }
  };

  const rejectIncomingCall = () => {
    if (incomingOffer) {
      sendSignal(incomingOffer.from, { action: 'reject' });
    }
    finishCall(false);
  };

  const hangupCall = () => {
    const target = callPeerId || incomingOffer?.from || null;
    if (target) {
      sendSignal(target, { action: 'hangup' });
    }
    finishCall(true);
  };

  // 消息时间格式化
  const formatMsgTime = (iso?: string) => {
    if (!iso) return '';
    const d = new Date(iso);
    const now = new Date();
    const diffDays = Math.floor((now.getTime() - d.getTime()) / 86400000);
    if (diffDays === 0) return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
    if (diffDays === 1) return '昨天';
    if (diffDays < 7) return ['日','一','二','三','四','五','六'][d.getDay()] ? `周${'日一二三四五六'[d.getDay()]}` : '';
    return `${d.getMonth() + 1}/${d.getDate()}`;
  };

  // 小工具
  const formatBytes = (bytes?: number | string) => {
    if (bytes === undefined || bytes === null) return '';
    const num = typeof bytes === 'string' ? Number(bytes) : bytes;
    if (Number.isNaN(num)) return '';
    if (num < 1024) return `${num} B`;
    const units = ['KB', 'MB', 'GB'];
    const i = Math.min(Math.floor(Math.log(num) / Math.log(1024)), units.length);
    const value = num / Math.pow(1024, i + 1);
    return `${value.toFixed(value >= 10 ? 0 : 1)} ${units[i]}`;
  };

  const isImage = (mime?: string, url?: string) => {
    if (!mime && !url) return false;
    const lower = (mime || '').toLowerCase();
    if (lower.startsWith('image/')) return true;
    return (url || '').match(/\.(jpg|jpeg|png|gif|webp)$/i) !== null;
  };

  const alertMessage = (msg: string) => {
    showToast(msg, 'error');
  };

  // 过滤后的数据
  const filteredGroups = useMemo(
    () =>
      groups.filter(
        (g) =>
          g.group_name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
          g.group_id?.toString().includes(searchTerm)
      ),
    [groups, searchTerm]
  );
  const filteredSessions = useMemo(
    () =>
      sessions.filter(
        (s) =>
          s.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
          s.receiveId?.toString().includes(searchTerm)
      ),
    [sessions, searchTerm]
  );
  const filteredContacts = useMemo(
    () =>
      contacts.filter(
        (c) =>
          c.nickname?.toLowerCase().includes(searchTerm.toLowerCase()) ||
          c.user_id?.toString().includes(searchTerm)
      ),
    [contacts, searchTerm]
  );
  const filteredAdminUsers = useMemo(
    () =>
      adminUsers.filter(
        (u) =>
          u.nickname?.toLowerCase().includes(searchTerm.toLowerCase()) ||
          u.email?.toLowerCase().includes(searchTerm.toLowerCase()) ||
          u.uuid?.toLowerCase().includes(searchTerm.toLowerCase())
      ),
    [adminUsers, searchTerm]
  );

  const callPeerSession = useMemo(() => {
    if (!callPeerId) return null;
    return sessions.find((s) => s.type === 'private' && s.receiveId === callPeerId) || null;
  }, [sessions, callPeerId]);

  // 初始化加载群组列表
  const fetchGroups = async () => {
    try {
      setLoadingGroups(true);
      const resp = await groupApi.loadMyGroups();
      if (resp.code === 200 || resp.code === 0) {
        setGroups(resp.data || []);
      }
    } catch (e) {
      console.error('Failed to load groups:', e);
    } finally {
      setLoadingGroups(false);
    }
  };

  const fetchSessions = async () => {
    try {
      const [priv, group] = await Promise.all([messageApi.getSessionList(), messageApi.getGroupSessionList()]);
      const privList: SessionItem[] =
        (priv.data || []).map((s) => ({
          sessionId: s.session_id,
          receiveId: s.receive_id,
          name: s.receive_name,
          avatar: s.avatar,
          type: 'private',
        })) || [];
      const groupList: SessionItem[] =
        (group.data || []).map((s) => ({
          sessionId: s.session_id,
          receiveId: s.group_id || s.receive_id || '',
          name: s.group_name || s.receive_name || '',
          avatar: s.avatar,
          type: 'group',
        })) || [];
      const merged = [...groupList, ...privList];
      // 去重：按 type+receiveId 保留第一条，防止后端历史数据重复
      const seen = new Set<string>();
      const deduped = merged.filter((s) => {
        const key = `${s.type}:${s.receiveId}`;
        if (seen.has(key)) return false;
        seen.add(key);
        return true;
      });
      setSessions(deduped);
    } catch (e) {
      console.error('Failed to load sessions:', e);
    }
  };

  const fetchContacts = async () => {
    try {
      const resp = await contactApi.getContactUserList();
      if (resp.code === 200 || resp.code === 0) {
        setContacts(resp.data || []);
      }
    } catch (e) {
      console.error(e);
    }
  };

  const fetchApplies = async () => {
    try {
      const resp = await contactApi.getContactApplyList();
      if (resp.code === 200 || resp.code === 0) {
        setApplies(resp.data || []);
      }
    } catch (e) {
      console.error(e);
    }
  };

  const fetchAdminUsers = async () => {
    if (!user?.isAdmin) return;
    try {
      setLoadingAdminUsers(true);
      const resp = await adminApi.getUserList();
      if (resp.code === 200 || resp.code === 0) {
        setAdminUsers(resp.data || []);
      }
    } catch (e) {
      console.error('Failed to load admin users:', e);
      showToast('加载用户列表失败', 'error');
    } finally {
      setLoadingAdminUsers(false);
    }
  };

  const syncMyProfile = async () => {
    if (!user) return;
    try {
      const resp = await userApi.getMyInfo();
      if (resp.code !== 200 && resp.code !== 0) return;
      const profile = resp.data || {};
      const isAdminRaw = (profile as any).is_admin ?? (profile as any).isAdmin ?? 0;
      const nextUser = {
        ...user,
        userId: (profile as any).user_id || user.userId,
        email: (profile as any).email || user.email,
        nickName: (profile as any).nickname || user.nickName,
        avatar: (profile as any).avatar || user.avatar,
        telephone: (profile as any).telephone || user.telephone,
        gender: Number((profile as any).gender ?? user.gender ?? 0),
        signature: (profile as any).signature || user.signature,
        birthday: (profile as any).birthday || user.birthday,
        isAdmin: Number(isAdminRaw) === 1,
        status: Number((profile as any).status ?? user.status ?? 0),
      };
      updateUser(nextUser);
    } catch (e) {
      console.error('sync my profile failed', e);
    }
  };

  const handleAdminAction = (action: 'disable' | 'enable' | 'set-admin' | 'delete', item: AdminUserItem) => {
    const actionMap = {
      disable: { title: '禁用用户', message: `确定禁用「${item.nickname}」吗？`, confirmLabel: '禁用', danger: true },
      enable: { title: '启用用户', message: `确定启用「${item.nickname}」吗？`, confirmLabel: '启用', danger: false },
      'set-admin': { title: '设置管理员', message: `确定将「${item.nickname}」设为管理员吗？`, confirmLabel: '设为管理员', danger: false },
      delete: { title: '删除用户', message: `确定删除「${item.nickname}」吗？该操作不可恢复。`, confirmLabel: '删除', danger: true },
    } as const;

    const current = actionMap[action];
    askConfirm({
      title: current.title,
      message: current.message,
      confirmLabel: current.confirmLabel,
      danger: current.danger,
      onConfirm: async () => {
        setConfirmDialog(null);
        try {
          let resp;
          if (action === 'disable') resp = await adminApi.disableUser(item.uuid);
          else if (action === 'enable') resp = await adminApi.ableUser(item.uuid);
          else if (action === 'set-admin') resp = await adminApi.setAdmin(item.uuid);
          else resp = await adminApi.deleteUser(item.uuid);

          if (resp.code === 200 || resp.code === 0) {
            showToast(resp.message || '操作成功', 'success');
            fetchAdminUsers();
          } else {
            showToast(resp.message || '操作失败', 'error');
          }
        } catch (e) {
          console.error('admin action failed', e);
          showToast('操作失败', 'error');
        }
      },
    });
  };

  useEffect(() => {
    syncMyProfile();
    fetchGroups();
    fetchSessions();
    fetchContacts();
    fetchApplies();
  }, []);

  useEffect(() => {
    if (!user?.isAdmin && activeTab === 'admin') {
      setActiveTab('sessions');
    }
  }, [user?.isAdmin, activeTab]);

  useEffect(() => {
    if (activeTab === 'admin' && user?.isAdmin) {
      fetchAdminUsers();
    }
  }, [activeTab, user?.isAdmin]);

  // 根据 API 地址自动推导 WebSocket 地址（http→ws，https→wss）
  const wsUrl = `${WS_BASE_URL}/v1/ws`;

  const { status, sendMessage } = useWebSocket(wsUrl, {
    onMessage: (msg: ChatMessage) => {
      if (msg.type === 3) {
        if (!msg.avData) return;
        try {
          const payload = JSON.parse(msg.avData) as SignalPayload;
          const fromUserId = msg.sendId;
          if (!fromUserId) return;

          if (payload.action === 'offer') {
            // 仅支持私聊通话；当前不在该会话时也提示
            setIncomingOffer({ from: fromUserId, payload });
            setCallStatus('ringing');
            setCallMode(payload.callType || 'audio');
            setCallPeerId(fromUserId);
            return;
          }

          if (payload.action === 'answer') {
            if (pcRef.current && payload.sdp) {
              const answerSdp = payload.sdp;
              void (async () => {
                try {
                  await pcRef.current!.setRemoteDescription(new RTCSessionDescription(answerSdp));
                  await flushPendingRemoteCandidates();
                  if (!callMetaRef.current.connectedAt) {
                    callMetaRef.current = { ...callMetaRef.current, connectedAt: Date.now() };
                  }
                  setCallStatus('in-call');
                } catch (error) {
                  console.error('set remote answer failed', error);
                }
              })();
            }
            return;
          }

          if (payload.action === 'candidate') {
            if (pcRef.current && payload.candidate) {
              const canApplyNow = !!pcRef.current.remoteDescription;
              if (canApplyNow) {
                void pcRef.current.addIceCandidate(new RTCIceCandidate(payload.candidate)).catch((error) => {
                  console.warn('add remote candidate failed', error);
                });
              } else {
                pendingRemoteCandidatesRef.current.push(payload.candidate);
              }
            }
            return;
          }

          if (payload.action === 'hangup') {
            showToast('对方已挂断', 'info');
            finishCall(true);
            return;
          }

          if (payload.action === 'reject') {
            showToast('对方已拒绝通话', 'info');
            finishCall(false);
            return;
          }
        } catch (e) {
          console.error('parse signaling failed', e);
        }
        return;
      }

      // 1. 更新或置顶会话列表
      setSessions((prev) => {
        const msgContent = msg.content || (msg.url ? '[文件]' : '');
        const myId = user?.userId;
        const existingIdx = prev.findIndex((s) => 
          s.type === 'group' 
            ? msg.receiveId === s.receiveId 
            : (msg.sendId === s.receiveId || msg.receiveId === s.receiveId)
        );

        if (existingIdx > -1) {
          const updated = [...prev];
          updated[existingIdx] = { 
            ...updated[existingIdx], 
            lastMessage: msgContent, 
            lastTime: msg.createdAt || new Date().toISOString() 
          };
          // 将有新消息的会话移到顶部
          const current = updated.splice(existingIdx, 1)[0];
          return [current, ...updated];
        }

        // 首次私聊消息：本地还没有会话项时，自动补一条，保证消息页实时出现
        if (myId && msg.receiveId[0] === 'U') {
          const peerId = msg.sendId === myId ? msg.receiveId : msg.sendId;
          if (peerId && peerId !== myId) {
            const sessionId = msg.sessionId || '';
            const newSession: SessionItem = {
              sessionId,
              receiveId: peerId,
              name: msg.sendId === myId ? (activeChat?.name || peerId) : (msg.sendName || peerId),
              avatar: msg.sendId === myId ? activeChat?.avatar : msg.sendAvatar,
              type: 'private',
              lastMessage: msgContent,
              lastTime: msg.createdAt || new Date().toISOString(),
            };
            return [newSession, ...prev];
          }
        }
        return prev;
      });

      // 2. 如果属于当前活跃聊天，追加到消息列表
      if (activeChat) {
        const peer = activeChat.receiveId;
        const myId = user?.userId;
        const isGroupMsg = activeChat.type === 'group' && msg.receiveId === peer;
        const isPrivateMsg = activeChat.type === 'private' && (
          (msg.sendId === peer && msg.receiveId === myId) || 
          (msg.sendId === myId && msg.receiveId === peer)
        );

        if (isGroupMsg || isPrivateMsg) {
          setMessages((prev) => {
            // 仅在有稳定时间戳时去重，避免 createdAt 缺失导致同发送者消息被误判重复
            const hasStableCreatedAt = !!msg.createdAt;
            const isDuplicate = hasStableCreatedAt
              ? prev.some(
                  (m) =>
                    m.createdAt === msg.createdAt &&
                    m.sendId === msg.sendId &&
                    m.receiveId === msg.receiveId,
                )
              : false;
            if (isDuplicate) return prev;
            return [...prev, msg];
          });
        }
      }
    },
  });

  useEffect(() => {
    if (callStatus !== 'in-call') {
      if (callTimerRef.current) {
        clearInterval(callTimerRef.current);
        callTimerRef.current = null;
      }
      return;
    }

    const getNowSeconds = () => {
      const connectedAt = callMetaRef.current.connectedAt || Date.now();
      return Math.max(0, Math.floor((Date.now() - connectedAt) / 1000));
    };

    setCallSeconds(getNowSeconds());
    callTimerRef.current = setInterval(() => {
      setCallSeconds(getNowSeconds());
    }, 1000);

    return () => {
      if (callTimerRef.current) {
        clearInterval(callTimerRef.current);
        callTimerRef.current = null;
      }
    };
  }, [callStatus]);

  // 修复时序问题：发起方先拿到本地流、后渲染小窗时，确保视频元素挂载后能回填 srcObject
  useEffect(() => {
    if (localVideoRef.current && localStreamRef.current && localVideoRef.current.srcObject !== localStreamRef.current) {
      localVideoRef.current.srcObject = localStreamRef.current;
      void localVideoRef.current.play().catch(() => {});
    }
    if (remoteVideoRef.current && remoteStreamRef.current && remoteVideoRef.current.srcObject !== remoteStreamRef.current) {
      remoteVideoRef.current.srcObject = remoteStreamRef.current;
      void remoteVideoRef.current.play().catch(() => {});
    }
    if (remoteAudioRef.current && remoteStreamRef.current && remoteAudioRef.current.srcObject !== remoteStreamRef.current) {
      remoteAudioRef.current.srcObject = remoteStreamRef.current;
      void remoteAudioRef.current.play().catch(() => {});
    }
  }, [callStatus, callMode, incomingOffer, callPeerId]);

  // 当前活跃群组（用于GroupDetailsModal）
  const activeGroup = activeChat?.type === 'group' ? groups.find((g) => g.group_id === activeChat.receiveId) : null;

  const updateSessionLastMsg = (chatReceiveId: string, list: ChatMessage[]) => {
    if (!list.length) return;
    const last = list[list.length - 1];
    const content = last.content || (last.url ? '[文件]' : '');
    setSessions((prev) => prev.map((s) =>
      s.receiveId === chatReceiveId ? { ...s, lastMessage: content, lastTime: last.createdAt } : s
    ));
  };

  const loadHistory = async (chat: SessionItem) => {
    setLoadingMessages(true);
    try {
      if (chat.type === 'group') {
      const resp = await messageApi.getGroupMessageList(chat.receiveId);
      if (resp.code === 200 || resp.code === 0) {
      const list = (resp.data || []).map((m) => ({
        sessionId: chat.sessionId,
        sendId: m.send_id,
        sendName: m.send_name,
        sendAvatar: m.send_avatar,
        receiveId: m.group_id || chat.receiveId,
            content: m.content,
            url: m.url,
        type: Number(m.type ?? 0),
        fileType: m.file_type,
        fileName: m.file_name,
        fileSize: m.file_size,
        avData: m.av_data,
        createdAt: m.created_at,
      })) as ChatMessage[];
        setMessages(list);
        updateSessionLastMsg(chat.receiveId, list);
      }
    } else {
      const resp = await messageApi.getMessageList(chat.sessionId);
        if (resp.code === 200 || resp.code === 0) {
          const list = (resp.data || []).map((m) => ({
            sessionId: chat.sessionId,
            sendId: m.send_id,
            sendName: m.send_name,
            sendAvatar: m.send_avatar,
            receiveId: m.receive_id,
            content: m.content,
            url: m.url,
          type: Number(m.type ?? 0),
          fileType: m.file_type,
          fileName: m.file_name,
          fileSize: m.file_size,
          avData: m.av_data,
          createdAt: m.created_at,
        })) as ChatMessage[];
        setMessages(list);
        updateSessionLastMsg(chat.receiveId, list);
      }
    }
    } catch (e) {
      console.error('load history failed', e);
    } finally {
      setLoadingMessages(false);
    }
  };

  useEffect(() => {
    if (activeChat) {
      loadHistory(activeChat);
    } else {
      setMessages([]);
    }
  }, [activeChat?.sessionId, activeChat?.receiveId, activeChat?.type]);

  // 切换会话或离开页面时自动清理通话资源
  useEffect(() => {
    return () => {
      cleanupCall();
    };
  }, []);

  // 有新消息时自动滚到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // 桌面端：标签页标题跟随当前会话
  useEffect(() => {
    document.title = activeChat ? `${activeChat.name} · GoLinko` : 'GoLinko';
    return () => { document.title = 'GoLinko'; };
  }, [activeChat?.name]);

  const ensureSession = async (chat: SessionItem): Promise<SessionItem> => {
    if (chat.sessionId) return chat;
    try {
      const resp = await messageApi.openSession(chat.receiveId);
      if (resp.code === 200 || resp.code === 0) {
        const sessionId = resp.data?.session_id || '';
        const updated: SessionItem = { ...chat, sessionId };
        setSessions((prev) => {
          if (prev.find((s) => s.sessionId === sessionId)) return prev;
          return [updated, ...prev];
        });
        setActiveChat(updated);
        return updated;
      }
    } catch (e) {
      console.error('open session failed', e);
    }
    return chat;
  };

  const handleHideSession = (session: SessionItem) => {
    if (!session.sessionId) {
      alertMessage('会话尚未建立，无法隐藏');
      return;
    }
    askConfirm({
      title: '隐藏会话',
      message: `确定隐藏与「${session.name}」的会话吗？隐藏后可再次发消息恢复。`,
      confirmLabel: '隐藏',
      danger: true,
      onConfirm: async () => {
        setConfirmDialog(null);
        try {
          const resp = await messageApi.hideSession(session.sessionId);
          if (resp.code === 200 || resp.code === 0) {
            setSessions((prev) => prev.filter((s) => s.sessionId !== session.sessionId));
            if (activeChat?.sessionId === session.sessionId) {
              setActiveChat(null);
              setMessages([]);
            }
            showToast('会话已隐藏', 'info');
          }
        } catch (e) {
          console.error('hide session failed', e);
          showToast('隐藏失败', 'error');
        }
      },
    });
  };

  const doSend = async () => {
    if (!message.trim() || !activeChat || !user || status !== 'open') return;
    const chat = await ensureSession(activeChat);
    if (!chat.sessionId) { alertMessage('无法建立会话，请稍后重试'); return; }
    const content = message;
    sendMessage({
      session_id: chat.sessionId,
      type: 0,
      content,
      send_name: user.nickName,
      send_avatar: user.avatar,
      receive_id: chat.receiveId,
    } as any);
    setSessions((prev) => prev.map((s) =>
      s.receiveId === chat.receiveId ? { ...s, lastMessage: content, lastTime: new Date().toISOString() } : s
    ));
    setMessage('');
  };

  const handleSelectGroup = (group: GroupInfo) => {
    setActiveTab('groups');
    setActiveChat({
      sessionId: '',
      receiveId: group.group_id,
      name: group.group_name,
      type: 'group',
    });
    setMobileView('chat');
  };

  const handleSelectSession = (session: SessionItem) => {
    setActiveTab('sessions');
    setActiveChat(session);
    setMobileView('chat');
  };

  const openContactChat = async (contact: ContactUserInfo) => {
    // 先查本地 sessions 列表，避免重复点击多次请求接口创建重复会话
    const existing = sessions.find((s) => s.receiveId === contact.user_id && s.type === 'private');
    if (existing) {
      setActiveChat(existing);
      setActiveTab('sessions');
      setMobileView('chat');
      return;
    }
    try {
      const allowed = await messageApi.checkOpenSessionAllowed(contact.user_id);
      if ((allowed.data?.allowed ?? true) === false) {
        return;
      }
      const resp = await messageApi.openSession(contact.user_id);
      if (resp.code === 200 || resp.code === 0) {
        const sessionId = resp.data?.session_id || '';
        const session: SessionItem = {
          sessionId,
          receiveId: contact.user_id,
          name: contact.nickname,
          avatar: contact.avatar,
          type: 'private',
        };
        setSessions((prev) => {
          if (prev.find((s) => s.sessionId === sessionId)) return prev;
          return [session, ...prev];
        });
        setActiveChat(session);
        setMobileView('chat');
      }
    } catch (e) {
      console.error(e);
    }
  };

  const handleFileUpload = async (file: File) => {
    if (!activeChat) return;

    // 前端校验：类型与大小
    const allowedExt = /(\.jpg|\.jpeg|\.png|\.gif|\.webp|\.pdf|\.doc|\.docx|\.xls|\.xlsx|\.txt|\.zip|\.mp4|\.mp3|\.wav|\.m4a|\.aac|\.ogg|\.webm)$/i;
    if (!allowedExt.test(file.name)) {
      alertMessage('不支持的文件类型');
      return;
    }
    if (file.size > 32 * 1024 * 1024) {
      alertMessage('文件大小超过 32MB 限制');
      return;
    }

    setFileUploading(true);
    try {
      const chat = await ensureSession(activeChat);
      if (!chat.sessionId) {
        alertMessage('无法建立会话，请稍后重试');
        return;
      }
      const resp = await messageApi.uploadFile(file);
      if (resp.code === 200 || resp.code === 0) {
        const { url, filename } = resp.data || { url: '', filename: file.name };
        const payload = {
          session_id: chat.sessionId,
          type: 2, // 2 = File (message_type_enum.File)
          url,
          content: filename,
          file_name: filename,
          file_type: file.type,
          file_size: String(file.size), // 后端是 string
          send_id: user?.userId,
          send_name: user?.nickName,
          send_avatar: user?.avatar,
          receive_id: chat.receiveId,
        };
        sendMessage(payload);
      }
    } catch (e) {
      console.error(e);
    } finally {
      setFileUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  };

  const sendVoiceFile = async (file: File) => {
    if (!activeChat || !user) return;
    try {
      const chat = await ensureSession(activeChat);
      if (!chat.sessionId) {
        alertMessage('无法建立会话，请稍后重试');
        return;
      }
      const resp = await messageApi.uploadFile(file);
      if (resp.code === 200 || resp.code === 0) {
        const { url, filename } = resp.data || { url: '', filename: file.name };
        sendMessage({
          session_id: chat.sessionId,
          type: 1,
          url,
          content: '语音消息',
          file_name: filename,
          file_type: file.type || 'audio/webm',
          file_size: String(file.size),
          send_id: user.userId,
          send_name: user.nickName,
          send_avatar: user.avatar,
          receive_id: chat.receiveId,
        } as any);
      }
    } catch (e) {
      console.error(e);
      alertMessage('语音发送失败');
    }
  };

  const stopRecording = () => {
    if (!mediaRecorderRef.current || mediaRecorderRef.current.state === 'inactive') return;
    shouldSendRecordingRef.current = true;
    mediaRecorderRef.current.stop();
  };

  const stopRecordingWithoutSend = () => {
    if (!mediaRecorderRef.current || mediaRecorderRef.current.state === 'inactive') return;
    shouldSendRecordingRef.current = false;
    mediaRecorderRef.current.stop();
  };

  const startRecording = async () => {
    if (!activeChat || status !== 'open') return;
    if (!navigator.mediaDevices?.getUserMedia || typeof MediaRecorder === 'undefined') {
      alertMessage('当前浏览器不支持录音');
      return;
    }

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      recordingStreamRef.current = stream;

      const candidates = ['audio/webm;codecs=opus', 'audio/ogg;codecs=opus', 'audio/webm', 'audio/ogg'];
      const mimeType = candidates.find((t) => MediaRecorder.isTypeSupported(t)) || '';
      const recorder = mimeType ? new MediaRecorder(stream, { mimeType }) : new MediaRecorder(stream);

      recordingChunksRef.current = [];
      recorder.ondataavailable = (event) => {
        if (event.data && event.data.size > 0) recordingChunksRef.current.push(event.data);
      };

      recorder.onstop = async () => {
        try {
          if (!shouldSendRecordingRef.current) return;
          const blob = new Blob(recordingChunksRef.current, { type: recorder.mimeType || 'audio/webm' });
          if (blob.size <= 0) {
            alertMessage('未录到有效语音');
            return;
          }
          const ext = recorder.mimeType.includes('ogg') ? 'ogg' : 'webm';
          const voiceFile = new File([blob], `voice_${Date.now()}.${ext}`, { type: blob.type || 'audio/webm' });
          await sendVoiceFile(voiceFile);
        } finally {
          if (recordTimerRef.current) {
            clearInterval(recordTimerRef.current);
            recordTimerRef.current = null;
          }
          recordingStreamRef.current?.getTracks().forEach((t) => t.stop());
          recordingStreamRef.current = null;
          mediaRecorderRef.current = null;
          recordingChunksRef.current = [];
          setRecording(false);
          setRecordWillCancel(false);
          setRecordSeconds(0);
        }
      };

      recorder.start();
      mediaRecorderRef.current = recorder;
      recordStartAtRef.current = Date.now();
      shouldSendRecordingRef.current = true;
      setRecordWillCancel(false);
      setRecordSeconds(0);
      if (recordTimerRef.current) clearInterval(recordTimerRef.current);
      recordTimerRef.current = setInterval(() => {
        const sec = Math.max(0, Math.floor((Date.now() - recordStartAtRef.current) / 1000));
        setRecordSeconds(sec);
        if (sec >= 60) {
          stopRecording();
        }
      }, 200);
      setRecording(true);
    } catch (e) {
      console.error(e);
      alertMessage('无法使用麦克风，请检查权限');
    }
  };

  useEffect(() => {
    return () => {
      Object.values(voiceAudioRefs.current).forEach((audio) => {
        if (!audio) return;
        audio.pause();
        audio.src = '';
      });
      if (mediaRecorderRef.current && mediaRecorderRef.current.state !== 'inactive') {
        stopRecordingWithoutSend();
      }
      if (recordTimerRef.current) clearInterval(recordTimerRef.current);
      recordingStreamRef.current?.getTracks().forEach((t) => t.stop());
    };
  }, []);

  useEffect(() => {
    if (!recording) return;
    const CANCEL_THRESHOLD = 80;
    const handlePointerMove = (e: PointerEvent) => {
      const upDistance = recordPointerStartYRef.current - e.clientY;
      const willCancel = upDistance > CANCEL_THRESHOLD;
      setRecordWillCancel(willCancel);
      shouldSendRecordingRef.current = !willCancel;
    };
    const handlePointerUp = () => {
      if (recordWillCancel) {
        stopRecordingWithoutSend();
        showToast('已取消发送语音', 'info');
      } else {
        stopRecording();
      }
    };
    window.addEventListener('pointerup', handlePointerUp);
    window.addEventListener('pointercancel', handlePointerUp);
    window.addEventListener('pointermove', handlePointerMove);
    return () => {
      window.removeEventListener('pointerup', handlePointerUp);
      window.removeEventListener('pointercancel', handlePointerUp);
      window.removeEventListener('pointermove', handlePointerMove);
    };
  }, [recording, recordWillCancel]);
  
  return (
    <div className="h-dvh overflow-hidden flex bg-[#f0f2f5] relative max-w-[1440px] mx-auto">
      {/* Toast 反馈 */}
      {toast && (
        <div className={`fixed top-5 left-1/2 -translate-x-1/2 z-[200] px-5 py-2.5 rounded-2xl text-sm font-medium shadow-xl text-white pointer-events-none
          ${
            toast.type === 'error' ? 'bg-red-500' :
            toast.type === 'info'  ? 'bg-gray-700' :
            'bg-[#07c160]'
          }`}
        >
          {toast.msg}
        </div>
      )}
      {/* ▌ 桌面版左侧导航条 ▌ */}
      <div className="hidden md:flex flex-col w-[64px] bg-[#1e2435] items-center pt-4 pb-5 shrink-0">
        {/* Logo */}
        <div className="w-10 h-10 bg-[#07c160] rounded-[14px] flex items-center justify-center mb-5 shadow-lg shadow-emerald-900/30">
          <span className="text-white font-black text-[15px] tracking-tight select-none">GL</span>
        </div>
        {/* 导航项 */}
        <div className="flex flex-col gap-0.5 w-full px-2">
          {navItems.map(({ key, Icon, label }) => (
            <button
              key={key}
              title={label}
              onClick={() => { setActiveTab(key as any); setMobileView('list'); }}
              className={`relative w-full h-10 rounded-xl flex items-center justify-center transition-all group
                ${activeTab === key
                  ? 'bg-white/15 text-white'
                  : 'text-slate-500 hover:bg-white/8 hover:text-slate-300'}`}
            >
              {activeTab === key && (
                <span className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-5 bg-[#07c160] rounded-full" />
              )}
              <Icon className="w-[19px] h-[19px]" />
              {key === 'contacts' && applies.length > 0 && (
                <span className="absolute top-1.5 right-1.5 w-1.5 h-1.5 bg-red-500 rounded-full ring-1 ring-[#1e2435]" />
              )}
            </button>
          ))}
        </div>
        <div className="flex-1" />
        {/* 头像 — 仅保留一处，去掉退出按钮 */}
        <button
          onClick={() => { setActiveTab('profile'); setMobileView('list'); }}
          title={user?.nickName || '个人资料'}
          className="w-10 h-10 rounded-full overflow-hidden ring-2 ring-white/10 hover:ring-[#07c160]/60 transition-all"
        >
          <Av url={user?.avatar} name={user?.nickName || 'U'} cls="w-10 h-10 rounded-full" />
        </button>
      </div>

      {/* ▌ 移动版底部标签栏 ▌ */}
      <nav className="md:hidden fixed bottom-0 inset-x-0 bg-white/95 backdrop-blur-sm border-t border-gray-200 z-50 flex flex-col">
        <div className="h-14 flex items-stretch">
        {navItems.map(({ key, Icon, label }) => (
          <button
            key={key}
            onClick={() => { setActiveTab(key as any); setMobileView('list'); }}
            className={`flex-1 flex flex-col items-center justify-center gap-0.5 relative transition-colors
              ${activeTab === key ? 'text-[#07c160]' : 'text-gray-400'}`}
          >
            <Icon className="w-5 h-5" />
            <span className="text-[10px] font-medium">{label}</span>
            {key === 'contacts' && applies.length > 0 && (
              <span className="absolute top-2 right-[calc(50%-14px)] w-4 h-4 bg-red-500 text-white text-[9px] rounded-full flex items-center justify-center font-bold">
                {applies.length > 9 ? '9+' : applies.length}
              </span>
            )}
          </button>
        ))}
        </div>
        {/* iPhone home indicator 安全区 */}
        <div className="shrink-0" style={{ height: 'env(safe-area-inset-bottom, 0px)' }} />
      </nav>

      {/* ▌ 列表面板 ▌ */}
      <div className={`flex flex-col bg-white border-r border-gray-100/80
        w-full md:w-[280px] xl:w-[300px] shrink-0
        absolute inset-0 md:static md:inset-auto
        pb-[var(--nav-total)] md:pb-0
        overflow-hidden z-10
        ${mobileView === 'list' ? 'flex' : 'hidden md:flex'}
      `}>
        {/* 面板头部 */}
        <div className="px-4 pt-[18px] pb-3 space-y-2.5">
          <div className="flex items-center justify-between h-8">
            {activeTab === 'contacts' && contactsSubView === 'new-friends' ? (
              <div className="flex items-center gap-1.5">
                <button
                  onClick={() => setContactsSubView('list')}
                  className="p-1 -ml-0.5 rounded-lg hover:bg-gray-100 text-gray-400 transition-colors"
                >
                  <ArrowLeft className="w-4 h-4" />
                </button>
                <h2 className="text-[17px] font-bold text-gray-900 tracking-tight">新朋友</h2>
              </div>
            ) : (
              <h2 className="text-[17px] font-bold text-gray-900 tracking-tight">
                {activeTab === 'sessions' ? '消息' : activeTab === 'groups' ? '群组' : activeTab === 'contacts' ? '联系人' : activeTab === 'admin' ? '管理' : '我的'}
              </h2>
            )}
            <div className="flex gap-1">
              {activeTab === 'groups' && (
                <button onClick={() => setIsSearchModalOpen(true)}
                  className="w-7 h-7 flex items-center justify-center rounded-lg hover:bg-gray-100 text-gray-400 transition-colors"
                  title="搜索群组">
                  <Search className="w-3.5 h-3.5" />
                </button>
              )}
              {(activeTab === 'groups' || activeTab === 'sessions') && (
                <button onClick={() => setIsModalOpen(true)}
                  className="w-7 h-7 flex items-center justify-center rounded-lg bg-[#07c160]/10 text-[#07c160] hover:bg-[#07c160]/20 transition-colors"
                  title="创建群组">
                  <Plus className="w-3.5 h-3.5" />
                </button>
              )}
            </div>
          </div>
          {/* 搜索栏：新朋友子视图不显示搜索 */}
          {!(activeTab === 'contacts' && contactsSubView === 'new-friends') && (
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-300 w-3.5 h-3.5 pointer-events-none" />
              <input
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="搜索"
                className="w-full pl-8 pr-3 py-1.5 bg-gray-100/80 rounded-lg text-[13px] outline-none focus:ring-2 ring-[#07c160]/20 focus:bg-gray-100 placeholder:text-gray-400 transition-all"
              />
            </div>
          )}
        </div>
        <div className="h-px bg-gray-100 mx-0" />

        {/* 列表内容区 */}
        <div className="flex-1 overflow-y-auto overscroll-contain">

          {/* 会话列表 */}
          {activeTab === 'sessions' && (
            filteredSessions.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-16 gap-3 text-gray-300">
                <Inbox className="w-14 h-14" />
                <span className="text-sm font-medium">暂无会话</span>
              </div>
            ) : (
              <div>
                {filteredSessions.map((s) => (
                  <div
                    key={s.sessionId + s.receiveId}
                    onClick={() => handleSelectSession(s)}
                    className={`relative flex items-center gap-3 px-3 py-2.5 mx-2 my-0.5 rounded-xl cursor-pointer transition-all group
                      ${activeChat?.sessionId === s.sessionId
                        ? 'bg-[#07c160]/10'
                        : 'hover:bg-gray-50'}`}
                  >
                    {activeChat?.sessionId === s.sessionId && (
                      <span className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-6 bg-[#07c160] rounded-full" />
                    )}
                    <Av url={s.avatar} name={s.name || 'S'} cls="w-11 h-11 rounded-full shrink-0" />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-1">
                        <span className={`text-[13.5px] font-semibold truncate leading-tight ${
                          activeChat?.sessionId === s.sessionId ? 'text-[#07c160]' : 'text-gray-800'
                        }`}>{s.name}</span>
                        <span className="text-[11px] text-gray-400 shrink-0 tabular-nums">{formatMsgTime(s.lastTime)}</span>
                      </div>
                      <p className="text-[12px] text-gray-400 truncate mt-0.5 leading-tight">{s.lastMessage || (s.type === 'group' ? '群聊' : '点击开始聊天')}</p>
                    </div>
                    <button
                      onClick={(e) => { e.stopPropagation(); handleHideSession(s); }}
                      className="p-1 rounded-md text-gray-300 hover:text-gray-500 hover:bg-gray-100 opacity-0 group-hover:opacity-100 transition-all shrink-0"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </div>
                ))}
              </div>
            )
          )}

          {/* 群组列表 */}
          {activeTab === 'groups' && (
            loadingGroups && groups.length === 0 ? (
              <div className="flex justify-center py-16">
                <Loader2 className="w-6 h-6 text-gray-300 animate-spin" />
              </div>
            ) : filteredGroups.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-16 gap-3 text-gray-300">
                <Users className="w-14 h-14" />
                <span className="text-sm font-medium">{searchTerm ? '未找到群组' : '暂无群组'}</span>
              </div>
            ) : (
              <div>
                {filteredGroups.map((g) => (
                  <div
                    key={g.group_id}
                    onClick={() => handleSelectGroup(g)}
                    className={`relative flex items-center gap-3 px-3 py-2.5 mx-2 my-0.5 rounded-xl cursor-pointer transition-all
                      ${activeChat?.receiveId === g.group_id && activeChat?.type === 'group'
                        ? 'bg-[#07c160]/10'
                        : 'hover:bg-gray-50'}`}
                  >
                    {activeChat?.receiveId === g.group_id && activeChat?.type === 'group' && (
                      <span className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-6 bg-[#07c160] rounded-full" />
                    )}
                    <div className={`w-11 h-11 rounded-xl flex items-center justify-center text-white font-bold text-base shrink-0 ${avatarBg(g.group_name || 'G')}`}>
                      {(g.group_name || 'G')[0].toUpperCase()}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className={`text-[13.5px] font-semibold truncate leading-tight ${
                        activeChat?.receiveId === g.group_id && activeChat?.type === 'group' ? 'text-[#07c160]' : 'text-gray-800'
                      }`}>{g.group_name}</p>
                      <p className="text-[12px] text-gray-400 truncate mt-0.5">{g.notice || '暂无公告'}</p>
                    </div>
                  </div>
                ))}
              </div>
            )
          )}

          {/* 联系人列表 */}
          {activeTab === 'contacts' && contactsSubView === 'list' && (
            <div className="p-3 space-y-2">
              {/* 功能入口区 */}
              <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden shadow-sm">
                {/* 新朋友 */}
                <button
                  onClick={() => { setContactsSubView('new-friends'); fetchApplies(); }}
                  className="w-full flex items-center gap-3 px-4 py-3.5 hover:bg-gray-50 transition-colors"
                >
                  <div className="w-10 h-10 rounded-2xl bg-red-50 flex items-center justify-center shrink-0">
                    <User className="w-5 h-5 text-red-400" />
                  </div>
                  <span className="flex-1 text-sm font-medium text-gray-900 text-left">新朋友</span>
                  <div className="flex items-center gap-2">
                    {applies.length > 0 && (
                      <span className="min-w-[20px] h-5 px-1.5 bg-red-500 text-white text-[11px] rounded-full flex items-center justify-center font-bold">
                        {applies.length > 99 ? '99+' : applies.length}
                      </span>
                    )}
                    <ChevronRight className="w-4 h-4 text-gray-300" />
                  </div>
                </button>
                <div className="border-t border-gray-50" />
                {/* 添加朋友 */}
                <button
                  onClick={() => applyInputRef.current?.focus()}
                  className="w-full flex items-center gap-3 px-4 py-3.5 hover:bg-gray-50 transition-colors"
                >
                  <div className="w-10 h-10 rounded-2xl bg-[#07c160]/10 flex items-center justify-center shrink-0">
                    <Plus className="w-5 h-5 text-[#07c160]" />
                  </div>
                  <span className="flex-1 text-sm font-medium text-gray-900 text-left">添加朋友 / 入群</span>
                  <ChevronRight className="w-4 h-4 text-gray-300" />
                </button>
              </div>
              {/* 添加联系人输入框 */}
              <div className="bg-gray-50 rounded-2xl p-3">
                <p className="text-xs font-semibold text-gray-500 mb-2">输入用户ID或群ID发起申请</p>
                <div className="flex gap-2">
                  <input
                    ref={applyInputRef}
                    placeholder="用户ID 或 群ID"
                    className="flex-1 px-3 py-2.5 bg-white border border-gray-200 rounded-xl text-sm outline-none focus:ring-2 ring-[#07c160]/20"
                  />
                  <button
                    onClick={async () => {
                      const val = applyInputRef.current?.value?.trim();
                      if (!val) return;
                      try {
                        await contactApi.applyAddContact(val);
                        showToast('申请已发送 👋');
                        fetchApplies();
                        if (applyInputRef.current) applyInputRef.current.value = '';
                      } catch { showToast('发送失败，请确认ID是否正确', 'error'); }
                    }}
                    className="px-4 py-2.5 bg-[#07c160] text-white rounded-xl text-sm font-medium hover:bg-[#06ad55] active:scale-95 transition-all"
                  >申请</button>
                </div>
              </div>
              {/* 联系人列表 */}
              {filteredContacts.length === 0 ? (
                <div className="flex flex-col items-center py-10 gap-3 text-gray-300">
                  <User className="w-14 h-14" />
                  <span className="text-sm">暂无联系人</span>
                </div>
              ) : (
                filteredContacts.map((c) => (
                  <div key={c.user_id} className="bg-white rounded-2xl border border-gray-100 overflow-hidden shadow-sm">
                    <div className="flex items-center gap-3 p-3">
                      <button
                        onClick={() => {
                          setViewingContact(c);
                          setViewingContactDetail(null);
                          userApi.getUserById(c.user_id).then(r => {
                            if (r.code === 200 || r.code === 0) setViewingContactDetail(r.data);
                          }).catch(() => {});
                        }}
                        className="shrink-0"
                      >
                        <Av url={c.avatar} name={c.nickname || 'U'} cls="w-11 h-11 rounded-full" />
                      </button>
                      <div className="flex-1 min-w-0">
                        <p className="font-semibold text-gray-900 truncate text-sm">{c.nickname}</p>
                        <p className="text-xs text-gray-400 truncate">{c.user_id}</p>
                      </div>
                      <button
                        onClick={() => {
                          setViewingContact(c);
                          setViewingContactDetail(null);
                          userApi.getUserById(c.user_id).then(r => {
                            if (r.code === 200 || r.code === 0) setViewingContactDetail(r.data);
                          }).catch(() => {});
                        }}
                        className="p-2 text-gray-400 hover:text-[#07c160] hover:bg-[#07c160]/10 rounded-xl transition-colors"
                        title="查看资料"
                      >
                        <Info className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => openContactChat(c)}
                        className="p-2 bg-[#07c160]/10 text-[#07c160] rounded-xl hover:bg-[#07c160]/20 transition-colors"
                        title="发消息"
                      >
                        <MessageSquare className="w-4 h-4" />
                      </button>
                    </div>
                    <div className="flex border-t border-gray-50 divide-x divide-gray-50">
                      <button
                        onClick={() => {
                          askConfirm({
                            title: '拉黑联系人',
                            message: `确定拉黑 ${c.nickname}？拉黑后对方无法向你发消息。`,
                            confirmLabel: '拉黑',
                            danger: true,
                            onConfirm: async () => {
                              setConfirmDialog(null);
                              try {
                                await contactApi.blackContact(c.user_id);
                                showToast(`已拉黑 ${c.nickname}`);
                                fetchContacts();
                              } catch { showToast('操作失败', 'error'); }
                            },
                          });
                        }}
                        className="flex-1 py-2 text-xs text-gray-500 hover:bg-gray-50 transition-colors flex items-center justify-center gap-1"
                      ><UserX className="w-3 h-3" />拉黑</button>
                      <button
                        onClick={async () => {
                          try {
                            await contactApi.unblackContact(c.user_id);
                            showToast('已解除拉黑');
                            fetchContacts();
                          } catch { showToast('操作失败', 'error'); }
                        }}
                        className="flex-1 py-2 text-xs text-gray-500 hover:bg-gray-50 transition-colors"
                      >解黑</button>
                      <button
                        onClick={() => {
                          askConfirm({
                            title: '删除联系人',
                            message: `确定删除联系人 ${c.nickname}？删除后需重新添加才能聊天。`,
                            confirmLabel: '删除',
                            danger: true,
                            onConfirm: async () => {
                              setConfirmDialog(null);
                              try {
                                await contactApi.deleteContact(c.user_id);
                                showToast(`已删除 ${c.nickname}`, 'info');
                                fetchContacts();
                              } catch { showToast('操作失败', 'error'); }
                            },
                          });
                        }}
                        className="flex-1 py-2 text-xs text-red-400 hover:bg-red-50 transition-colors"
                      >删除</button>
                    </div>
                  </div>
                ))
              )}
            </div>
          )}

          {/* 新朋友子视图 */}
          {activeTab === 'contacts' && contactsSubView === 'new-friends' && (
            <div className="p-3 space-y-2">
              {applies.length === 0 ? (
                <div className="flex flex-col items-center py-20 gap-3 text-gray-300">
                  <div className="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center">
                    <User className="w-8 h-8" />
                  </div>
                  <span className="text-sm font-medium">暂无新的申请</span>
                </div>
              ) : (
                applies.map((a) => (
                  <div key={a.apply_id} className="bg-white rounded-2xl border border-gray-100 overflow-hidden shadow-sm">
                    <div className="flex items-center gap-3 p-4">
                      <Av
                        url={a.avatar}
                        name={a.applicant || 'A'}
                        cls="w-12 h-12 rounded-full shrink-0"
                      />
                      <div className="flex-1 min-w-0">
                        <p className="font-semibold text-gray-900 truncate text-sm">{a.applicant}</p>
                        <p className="text-xs text-gray-400 mt-0.5">
                          {a.contact_type === 0 ? '申请添加为好友' : `申请加入「${a.target_name}」`}
                        </p>
                        {a.message && (
                          <p className="text-xs text-gray-500 mt-1 bg-gray-50 rounded-lg px-2 py-1 italic">"{a.message}"</p>
                        )}
                      </div>
                    </div>
                    <div className="flex gap-2 px-4 pb-4">
                      <button
                        onClick={async () => {
                          try {
                            const resp = await contactApi.acceptContactApply(a.apply_id);
                            if (resp.code !== 200 && resp.code !== 0) {
                              showToast(resp.message || '操作失败', 'error');
                              return;
                            }
                            showToast(`已同意 ${a.applicant} 的申请`);
                            fetchApplies();
                            fetchContacts();
                            // 同意好友申请后自动打开会话
                            try {
                              const sessResp = await messageApi.openSession(a.applicant_id);
                              if (sessResp.code === 200 || sessResp.code === 0) {
                                const newSession: SessionItem = {
                                  sessionId: sessResp.data?.session_id || '',
                                  receiveId: a.applicant_id,
                                  name: a.applicant,
                                  avatar: a.avatar,
                                  type: 'private',
                                };
                                setSessions((prev) => {
                                  if (prev.find((s) => s.receiveId === a.applicant_id && s.type === 'private')) return prev;
                                  return [newSession, ...prev];
                                });
                                setActiveChat(newSession);
                                setActiveTab('sessions');
                                setContactsSubView('list');
                                setMobileView('chat');
                              }
                            } catch { /* 会话创建失败不影响主流程 */ }
                          } catch { showToast('操作失败，请重试', 'error'); }
                        }}
                        className="flex-1 py-2 bg-[#07c160] text-white rounded-xl text-sm font-medium hover:bg-[#06ad55] active:scale-95 transition-all flex items-center justify-center gap-1.5"
                      >
                        <Check className="w-3.5 h-3.5" />同意
                      </button>
                      <button
                        onClick={async () => {
                          try {
                            const resp = await contactApi.rejectContactApply(a.apply_id);
                            if (resp.code !== 200 && resp.code !== 0) {
                              showToast(resp.message || '操作失败', 'error');
                              return;
                            }
                            showToast('已拒绝申请', 'info');
                            fetchApplies();
                          } catch { showToast('操作失败，请重试', 'error'); }
                        }}
                        className="flex-1 py-2 bg-gray-100 text-gray-600 rounded-xl text-sm font-medium hover:bg-gray-200 active:scale-95 transition-all"
                      >拒绝</button>
                    </div>
                  </div>
                ))
              )}
            </div>
          )}

          {/* 管理员面板 */}
          {activeTab === 'admin' && user?.isAdmin && (
            <div className="p-3 space-y-2">
              {loadingAdminUsers ? (
                <div className="flex justify-center py-16">
                  <Loader2 className="w-6 h-6 text-gray-300 animate-spin" />
                </div>
              ) : filteredAdminUsers.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-16 gap-3 text-gray-300">
                  <Shield className="w-14 h-14" />
                  <span className="text-sm font-medium">暂无可管理用户</span>
                </div>
              ) : (
                filteredAdminUsers.map((u) => (
                  <div key={u.uuid} className="bg-white rounded-2xl border border-gray-100 overflow-hidden shadow-sm">
                    <div className="flex items-center gap-3 p-4">
                      <Av url={u.avatar} name={u.nickname || 'U'} cls="w-11 h-11 rounded-full shrink-0" />
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <p className="font-semibold text-gray-900 truncate text-sm">{u.nickname || u.uuid}</p>
                          {u.is_admin && (
                            <span className="text-[10px] px-1.5 py-0.5 rounded bg-amber-100 text-amber-700">管理员</span>
                          )}
                          {u.status === 1 && (
                            <span className="text-[10px] px-1.5 py-0.5 rounded bg-red-100 text-red-600">已禁用</span>
                          )}
                        </div>
                        <p className="text-xs text-gray-400 truncate mt-0.5">{u.email || '无邮箱'} · {u.uuid}</p>
                      </div>
                    </div>
                    <div className="grid grid-cols-4 border-t border-gray-50 divide-x divide-gray-50">
                      {u.status === 1 ? (
                        <button
                          onClick={() => handleAdminAction('enable', u)}
                          className="py-2 text-xs text-[#07c160] hover:bg-green-50 transition-colors flex items-center justify-center gap-1"
                        >
                          <Check className="w-3 h-3" />启用
                        </button>
                      ) : (
                        <button
                          onClick={() => handleAdminAction('disable', u)}
                          className="py-2 text-xs text-orange-500 hover:bg-orange-50 transition-colors flex items-center justify-center gap-1"
                        >
                          <Ban className="w-3 h-3" />禁用
                        </button>
                      )}
                      <button
                        onClick={() => handleAdminAction('set-admin', u)}
                        disabled={u.is_admin}
                        className="py-2 text-xs text-blue-500 hover:bg-blue-50 transition-colors flex items-center justify-center gap-1 disabled:text-gray-300 disabled:hover:bg-transparent"
                      >
                        <UserCog className="w-3 h-3" />设管
                      </button>
                      <button
                        onClick={() => navigator.clipboard?.writeText(u.uuid || '')}
                        className="py-2 text-xs text-gray-500 hover:bg-gray-50 transition-colors flex items-center justify-center gap-1"
                      >
                        <Copy className="w-3 h-3" />复制ID
                      </button>
                      <button
                        onClick={() => handleAdminAction('delete', u)}
                        className="py-2 text-xs text-red-500 hover:bg-red-50 transition-colors"
                      >
                        删除
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>
          )}

          {/* 个人信息 */}
          {activeTab === 'profile' && (
            <div className="p-4 space-y-3 overflow-y-auto">
              {/* 头像 & 名称 */}
              <div className="bg-gradient-to-br from-[#07c160]/10 to-emerald-50/60 rounded-2xl p-5 flex flex-col items-center gap-3">
                <Av url={user?.avatar} name={user?.nickName || 'U'} cls="w-20 h-20 rounded-full shadow-lg ring-4 ring-white" />
                <div className="text-center">
                  <h3 className="font-bold text-gray-900 text-lg leading-tight">{user?.nickName}</h3>
                  <p className="text-xs text-gray-400 mt-1.5 flex items-center gap-1.5 justify-center">
                    <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${status === 'open' ? 'bg-[#07c160] animate-pulse' : 'bg-gray-300'}`} />
                    {status === 'open' ? '在线' : '离线'}
                  </p>
                </div>
              </div>
              {/* 信息字段 */}
              <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden shadow-sm">
                {[
                  { label: 'ID', value: user?.userId },
                  { label: '邮箱', value: user?.email },
                  { label: '手机', value: user?.telephone || '未设置' },
                ].map(({ label, value }, i, arr) => (
                  <div key={label} className={`flex items-center gap-3 px-4 py-3.5 ${i < arr.length - 1 ? 'border-b border-gray-50' : ''}`}>
                    <span className="text-xs text-gray-400 w-8 shrink-0 font-medium">{label}</span>
                    <span className="flex-1 text-sm text-gray-800 break-all leading-snug">{value}</span>
                    {value && value !== '未设置' && (
                      <button
                        onClick={() => { navigator.clipboard?.writeText(value!); showToast('已复制'); }}
                        className="text-gray-300 hover:text-[#07c160] transition-colors shrink-0"
                      >
                        <Copy className="w-3.5 h-3.5" />
                      </button>
                    )}
                  </div>
                ))}
              </div>
              {/* 操作按钮 */}
              <button
                onClick={() => setIsProfileOpen(true)}
                className="w-full py-3 bg-[#07c160]/10 text-[#07c160] rounded-2xl text-sm font-semibold hover:bg-[#07c160]/20 active:scale-[0.99] transition-all"
              >
                编辑个人资料
              </button>
              <button
                onClick={logout}
                className="w-full py-3 bg-red-50 text-red-500 rounded-2xl text-sm font-semibold hover:bg-red-100 active:scale-[0.99] transition-all"
              >
                退出登录
              </button>
            </div>
          )}
        </div>

      </div>

      {/* ▌ 聊天区域 ▌ */}
      <div className={`flex-1 flex flex-col overflow-hidden
        absolute inset-0 md:static md:inset-auto
        pb-[var(--nav-total)] md:pb-0
        ${mobileView === 'chat' ? 'flex' : 'hidden md:flex'}
        md:flex
      `}>
        {activeChat ? (
          <>
            {/* 聊天头部 */}
            <div className="flex items-center gap-3 px-4 h-[58px] bg-white/95 backdrop-blur-sm border-b border-gray-100 shrink-0">
              <button
                className="md:hidden p-1.5 -ml-1 rounded-xl hover:bg-gray-100 text-gray-600 transition-colors"
                onClick={() => setMobileView('list')}
              >
                <ArrowLeft className="w-5 h-5" />
              </button>
              <Av url={activeChat.avatar} name={activeChat.name || '?'} cls="w-9 h-9 rounded-full shrink-0 ring-2 ring-gray-100" />
              <div className="flex-1 min-w-0">
                <h2 className="text-[14px] font-bold text-gray-900 leading-tight truncate">{activeChat.name}</h2>
                <div className="flex items-center gap-1.5 mt-0.5">
                  <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${status === 'open' ? 'bg-[#07c160] animate-pulse' : 'bg-gray-300'}`} />
                  <span className="text-[11px] text-gray-400">{activeChat.type === 'group' ? '群聊' : '私聊'}</span>
                </div>
              </div>
              {activeChat.type === 'group' && (
                <button
                  onClick={() => setIsDetailsOpen(true)}
                  className="p-2 text-gray-400 hover:text-[#07c160] hover:bg-[#07c160]/8 rounded-xl transition-all"
                  title="群组设置"
                >
                  <Settings className="w-4.5 h-4.5" />
                </button>
              )}
              {activeChat.type === 'private' && (
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => startCall('audio')}
                    disabled={callStatus !== 'idle'}
                    className="p-2 text-gray-400 hover:text-[#07c160] hover:bg-[#07c160]/8 rounded-xl transition-all disabled:opacity-40"
                    title="语音通话"
                  >
                    <Phone className="w-4.5 h-4.5" />
                  </button>
                  <button
                    onClick={() => startCall('video')}
                    disabled={callStatus !== 'idle'}
                    className="p-2 text-gray-400 hover:text-[#07c160] hover:bg-[#07c160]/8 rounded-xl transition-all disabled:opacity-40"
                    title="视频通话"
                  >
                    <Video className="w-4.5 h-4.5" />
                  </button>
                </div>
              )}
            </div>

            {/* 消息列表 */}
            <div className="flex-1 overflow-y-auto overscroll-contain px-4 py-4 bg-[#f0f2f5]">
              {loadingMessages ? (
                <div className="flex justify-center py-10">
                  <Loader2 className="w-6 h-6 text-gray-400 animate-spin" />
                </div>
              ) : messages.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full gap-2 text-gray-400 select-none">
                  <MessageSquare className="w-12 h-12 opacity-20" />
                  <p className="text-sm">还没有消息，打个招呼吧 👋</p>
                </div>
              ) : (
              <div className="max-w-3xl mx-auto space-y-3">
                {messages.map((msg, idx) => {
                  const isMine = msg.sendId === user?.userId;
                  const isVoice = msg.type === 1;
                  const isFile = msg.type === 2;
                  const time = msg.createdAt
                    ? new Date(msg.createdAt).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
                    : '';
                  return (
                    <div key={idx} className={`flex gap-2.5 ${isMine ? 'flex-row-reverse' : 'flex-row'}`}>
                      <Av
                        url={isMine ? user?.avatar : msg.sendAvatar}
                        name={isMine ? (user?.nickName || 'me') : (msg.sendName || '?')}
                        cls="w-9 h-9 rounded-full shrink-0"
                      />
                      <div className={`flex flex-col gap-1 max-w-[72%] md:max-w-[520px] ${isMine ? 'items-end' : 'items-start'}`}>
                        {!isMine && (
                          <span className="text-[11px] text-gray-500 ml-1 font-medium">{msg.sendName}</span>
                        )}
                        <div className={`rounded-2xl px-4 py-2.5 text-sm shadow-sm ${
                          isMine
                            ? 'bg-[#07c160] text-white rounded-tr-[4px]'
                            : 'bg-white text-gray-800 rounded-tl-[4px] border border-gray-100'
                        }`}>
                          {isVoice && msg.url ? (
                            (() => {
                              const baseUrl = API_BASE_URL;
                              const voiceUrl = msg.url.startsWith('http') ? msg.url : `${baseUrl}${msg.url.startsWith('/') ? '' : '/'}${msg.url}`;
                              const duration = voiceDurations[voiceUrl] || 0;
                              const bubbleWidth = Math.min(220, Math.max(120, Math.round(120 + Math.min(duration, 60) * 1.6)));
                              const isPlaying = playingVoiceUrl === voiceUrl;
                              return (
                                <button
                                  type="button"
                                  onClick={() => void toggleVoicePlayback(voiceUrl)}
                                  className={`group flex items-center gap-2 rounded-2xl px-3 py-2 transition-colors ${isMine ? 'bg-white/15 hover:bg-white/25 text-white' : 'bg-gray-50 hover:bg-gray-100 text-gray-800'}`}
                                  style={{ width: `${bubbleWidth}px` }}
                                >
                                  <span className={`w-7 h-7 rounded-full flex items-center justify-center shrink-0 ${isMine ? 'bg-white/25' : 'bg-[#07c160]/15'}`}>
                                    {isPlaying ? <Pause className={`w-3.5 h-3.5 ${isMine ? 'text-white' : 'text-[#07c160]'}`} /> : <Play className={`w-3.5 h-3.5 ${isMine ? 'text-white' : 'text-[#07c160]'}`} />}
                                  </span>
                                  <span className="flex-1 h-4 flex items-center gap-1.5">
                                    {[0, 1, 2, 3].map((i) => (
                                      <span
                                        key={i}
                                        className={`w-1 rounded-full ${isMine ? 'bg-white/90' : 'bg-[#07c160]'} ${isPlaying ? 'animate-pulse' : ''}`}
                                        style={{ height: `${6 + i * 2}px`, animationDelay: `${i * 0.08}s` }}
                                      />
                                    ))}
                                  </span>
                                  <span className={`text-[11px] tabular-nums shrink-0 ${isMine ? 'text-white/90' : 'text-gray-500'}`}>
                                    {formatVoiceSeconds(duration)}
                                  </span>
                                  <audio
                                    preload="metadata"
                                    src={voiceUrl}
                                    className="hidden"
                                    onContextMenu={(e) => e.preventDefault()}
                                    ref={(el) => {
                                      voiceAudioRefs.current[voiceUrl] = el;
                                    }}
                                    onLoadedMetadata={(e) => {
                                      const seconds = e.currentTarget.duration || 0;
                                      if (!seconds) return;
                                      setVoiceDurations((prev) => (prev[voiceUrl] ? prev : { ...prev, [voiceUrl]: seconds }));
                                    }}
                                    onEnded={() => {
                                      if (playingVoiceUrl === voiceUrl) setPlayingVoiceUrl(null);
                                    }}
                                  />
                                </button>
                              );
                            })()
                          ) : isFile && msg.url ? (
                            <div className="space-y-2">
                              {(() => {
                                const baseUrl = API_BASE_URL;
                                const fileUrl = msg.url.startsWith('http') ? msg.url : `${baseUrl}${msg.url.startsWith('/') ? '' : '/'}${msg.url}`;
                                
                                return (
                                  <>
                                    {isImage(msg.fileType, msg.url) && (
                                      <div 
                                        className="cursor-zoom-in group relative overflow-hidden rounded-xl bg-black/5"
                                        onClick={() => setPreviewImage(fileUrl)}
                                      >
                                        <img 
                                          src={fileUrl} 
                                          alt={msg.fileName} 
                                          className="max-w-[240px] max-h-[320px] object-cover transition-transform group-hover:scale-105 duration-300" 
                                        />
                                        <div className="absolute inset-0 bg-black/5 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-white text-[10px]">
                                          点击预览
                                        </div>
                                      </div>
                                    )}
                                    <div className="flex items-center gap-3">
                                      <div className={`p-2 rounded-lg ${isMine ? 'bg-white/20' : 'bg-[#07c160]/5'}`}>
                                        <Paperclip className={`w-4 h-4 ${isMine ? 'text-white' : 'text-[#07c160]'}`} />
                                      </div>
                                      <div className="flex flex-col min-w-0">
                                        <a 
                                          href={fileUrl} 
                                          download={msg.fileName}
                                          target="_blank" 
                                          rel="noreferrer"
                                          className={`font-medium text-sm truncate max-w-[160px] underline decoration-transparent hover:decoration-current transition-all ${isMine ? 'text-white' : 'text-[#07c160]'}`}
                                        >
                                          {msg.fileName || '下载文件'}
                                        </a>
                                        {msg.fileSize !== undefined && (
                                          <span className={`text-[10px] opacity-60 ${isMine ? 'text-white/80' : 'text-gray-400'}`}>
                                            {formatBytes(msg.fileSize)}
                                          </span>
                                        )}
                                      </div>
                                    </div>
                                  </>
                                );
                              })()}
                            </div>
                          ) : (
                            <span className="whitespace-pre-wrap break-words leading-relaxed">{msg.content}</span>
                          )}
                        </div>
                        <span className="text-[10px] text-gray-400 px-1">{time}</span>
                      </div>
                    </div>
                  );
                })}
              <div ref={messagesEndRef} />
              </div>
            )}
            </div>

            {/* 输入区 */}
            <div className="bg-white border-t border-gray-100 px-3 py-2.5 shrink-0">
              <div className="flex items-end gap-2 max-w-3xl mx-auto">
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={fileUploading || status !== 'open'}
                  className="p-2.5 text-gray-400 hover:text-[#07c160] disabled:opacity-40 transition-colors shrink-0"
                >
                  {fileUploading ? <Loader2 className="w-5 h-5 animate-spin" /> : <Paperclip className="w-5 h-5" />}
                </button>
                <button
                  type="button"
                  onPointerDown={(e) => {
                    e.preventDefault();
                    recordPointerStartYRef.current = e.clientY;
                    if (!recording) void startRecording();
                  }}
                  disabled={status !== 'open'}
                  className={`p-2.5 transition-colors shrink-0 touch-none ${recording ? 'text-red-500' : 'text-gray-400 hover:text-[#07c160]'} disabled:opacity-40`}
                  title={recording ? '上滑取消，松手发送' : '按住录音，上滑取消，松手发送'}
                >
                  <Mic className={`w-5 h-5 ${recording ? 'animate-pulse' : ''}`} />
                </button>
                <input ref={fileInputRef} type="file" hidden onChange={(e) => { const f = e.target.files?.[0]; if (f) handleFileUpload(f); }} />
                <div className="flex-1 bg-gray-50 border border-gray-200 rounded-2xl px-4 py-2.5 focus-within:ring-2 ring-[#07c160]/20 focus-within:border-[#07c160]/30 transition-all">
                  <textarea
                    rows={1}
                    value={message}
                    onChange={(e) => {
                      setMessage(e.target.value);
                      e.target.style.height = 'auto';
                      e.target.style.height = Math.min(e.target.scrollHeight, 120) + 'px';
                    }}
                    onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); doSend(); } }}
                    placeholder={recording ? '录音中，上滑取消，松手发送…' : '输入消息… (Enter 发送，Shift+Enter 换行)'}
                    className="w-full bg-transparent text-sm text-gray-800 outline-none resize-none placeholder:text-gray-400 max-h-[120px] leading-relaxed"
                  />
                </div>
                <button
                  type="button"
                  onClick={doSend}
                  disabled={!message.trim() || status !== 'open'}
                  className="w-10 h-10 bg-[#07c160] text-white rounded-full flex items-center justify-center hover:bg-[#06ad55] disabled:opacity-40 disabled:cursor-not-allowed active:scale-90 transition-all shadow-md shadow-emerald-200 shrink-0"
                >
                  <Send className="w-4 h-4" />
                </button>
              </div>
            </div>
          </>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center bg-[#f5f6f8] select-none">
            <div className="text-center space-y-5">
              <div className="w-20 h-20 bg-[#07c160] rounded-[24px] flex items-center justify-center mx-auto shadow-xl shadow-emerald-200">
                <span className="text-white font-black text-3xl tracking-tighter">CC</span>
              </div>
              <div className="space-y-1.5">
                <h3 className="text-[22px] font-bold text-gray-800 tracking-tight">GoLinko</h3>
                <p className="text-[13px] text-gray-400">选择一个会话，开始聊天</p>
              </div>
              <button
                onClick={() => setIsModalOpen(true)}
                className="px-5 py-2 bg-[#07c160] text-white rounded-full text-[13px] font-semibold hover:bg-[#06ad55] transition-all shadow-md shadow-emerald-200 active:scale-95"
              >
                + 创建群组
              </button>
            </div>
          </div>
        )}
      </div>

      {/* 弹窗 */}
      {isModalOpen && (
        <CreateGroupModal
          onClose={() => setIsModalOpen(false)}
          onSuccess={() => { fetchGroups(); setIsModalOpen(false); }}
        />
      )}
      {isSearchModalOpen && (
        <SearchGroupModal
          onClose={() => setIsSearchModalOpen(false)}
          onJoinSuccess={(groupId) => {
            fetchGroups();
            setActiveChat({ sessionId: '', receiveId: groupId, name: groups.find(g => g.group_id === groupId)?.group_name || '群聊', type: 'group' });
            setIsSearchModalOpen(false);
          }}
        />
      )}
      {isDetailsOpen && activeChat && activeGroup && (
        <GroupDetailsModal
          groupId={activeGroup.group_id}
          groupName={activeGroup.group_name}
          ownerId={activeGroup.owner_id}
          currentUserId={user?.userId || ''}
          onClose={() => setIsDetailsOpen(false)}
          onChanged={() => { fetchGroups(); fetchSessions(); }}
        />
      )}
      {isProfileOpen && (
        <ProfileModal onClose={() => setIsProfileOpen(false)} />
      )}
      {/* 好友资料弹窗 */}
      {viewingContact && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[300] flex items-center justify-center p-4"
          onClick={() => setViewingContact(null)}>
          <div className="bg-white rounded-2xl w-full max-w-xs shadow-2xl animate-in zoom-in duration-150"
            onClick={(e) => e.stopPropagation()}>
            <div className="flex justify-end p-3 pb-0">
              <button onClick={() => setViewingContact(null)}
                className="p-1.5 hover:bg-gray-100 rounded-full text-gray-400">
                <X className="w-4 h-4" />
              </button>
            </div>
            <div className="px-5 pb-5 text-center">
              <Av
                url={viewingContactDetail?.avatar || viewingContact.avatar}
                name={viewingContact.nickname || 'U'}
                cls="w-20 h-20 rounded-full mx-auto shadow-lg ring-4 ring-white"
              />
              <h3 className="font-bold text-gray-900 text-lg mt-3">{viewingContact.nickname}</h3>
              {viewingContactDetail?.signature && (
                <p className="text-sm text-gray-400 mt-1 italic">"{viewingContactDetail.signature}"</p>
              )}
              <div className="mt-4 space-y-2 text-left">
                {[
                  { label: 'ID', value: viewingContact.user_id },
                  ...(viewingContactDetail?.email ? [{ label: '邮箱', value: viewingContactDetail.email }] : []),
                  ...(viewingContactDetail?.telephone ? [{ label: '手机', value: viewingContactDetail.telephone }] : []),
                ].map(({ label, value }) => (
                  <div key={label} className="flex items-center gap-2 bg-gray-50 rounded-xl px-3 py-2.5">
                    <span className="text-xs text-gray-400 w-8 shrink-0 font-medium">{label}</span>
                    <span className="text-sm text-gray-700 flex-1 break-all">{value}</span>
                    <button
                      onClick={() => { navigator.clipboard?.writeText(value); showToast('已复制'); }}
                      className="text-gray-300 hover:text-[#07c160] transition-colors"
                    ><Copy className="w-3.5 h-3.5" /></button>
                  </div>
                ))}
              </div>
              <button
                onClick={() => { openContactChat(viewingContact); setViewingContact(null); }}
                className="mt-4 w-full py-2.5 bg-[#07c160] text-white rounded-xl text-sm font-medium hover:bg-[#06ad55] active:scale-95 transition-all"
              >
                发消息
              </button>
            </div>
          </div>
        </div>
      )}
      {/* 确认对话框 */}
      {confirmDialog && (
        <ConfirmDialog
          title={confirmDialog.title}
          message={confirmDialog.message}
          confirmLabel={confirmDialog.confirmLabel}
          danger={confirmDialog.danger}
          onConfirm={confirmDialog.onConfirm}
          onCancel={() => setConfirmDialog(null)}
        />
      )}

      {/* 图片预览 */}
      {previewImage && (
        <ImagePreview url={previewImage} onClose={() => setPreviewImage(null)} />
      )}

      {recording && (
        <div className="fixed inset-0 z-[280] pointer-events-none flex items-center justify-center">
          <div className={`text-white rounded-2xl px-6 py-5 shadow-2xl min-w-[220px] text-center transition-colors ${recordWillCancel ? 'bg-red-500/85' : 'bg-black/75'}`}>
            <div className="flex items-center justify-center gap-2 mb-2">
              <span className={`w-2.5 h-2.5 rounded-full ${recordWillCancel ? 'bg-white' : 'bg-red-500 animate-pulse'}`} />
              <span className="text-sm font-medium">{recordWillCancel ? '松手取消发送' : '正在录音'}</span>
            </div>
            <div className="text-3xl font-bold tracking-wider tabular-nums">{formatRecordDuration(recordSeconds)}</div>
            <div className="mt-2 text-xs text-white/80">上滑取消，松手发送，最长 60 秒</div>
          </div>
        </div>
      )}

      {(callStatus !== 'idle' || incomingOffer) && (
        <div className="fixed inset-0 z-[260] bg-black/75 backdrop-blur-sm p-3 md:p-5">
          <audio ref={remoteAudioRef} autoPlay />

          {callMode === 'video' ? (
            <div className="relative w-full h-full rounded-3xl overflow-hidden bg-black border border-white/10 shadow-2xl">
              <video ref={remoteVideoRef} autoPlay playsInline className="w-full h-full object-cover" />
              <div className="absolute inset-0 bg-gradient-to-t from-black/45 via-transparent to-black/20 pointer-events-none" />

              <div className="absolute top-4 left-4 text-white">
                <p className="text-sm font-semibold">{callPeerSession?.name || '对方'}</p>
                <p className="text-xs text-white/85">
                  {callStatus === 'in-call' ? formatCallDuration(callSeconds) : callStatus === 'calling' ? '呼叫中…' : '来电中…'}
                </p>
              </div>

              <div className="absolute right-4 top-4 w-28 h-40 md:w-36 md:h-52 rounded-2xl overflow-hidden border border-white/30 shadow-xl bg-black/40">
                <video ref={localVideoRef} autoPlay muted playsInline className="w-full h-full object-cover" />
              </div>

              <div className="absolute left-1/2 -translate-x-1/2 bottom-6 flex items-center gap-3">
                {incomingOffer && callStatus === 'ringing' ? (
                  <>
                    <button
                      onClick={rejectIncomingCall}
                      className="px-4 py-2 rounded-full bg-gray-700/90 hover:bg-gray-700 text-white text-sm"
                    >
                      拒绝
                    </button>
                    <button
                      onClick={acceptIncomingCall}
                      className="px-4 py-2 rounded-full bg-[#07c160] hover:bg-[#06ad55] text-white text-sm"
                    >
                      接听
                    </button>
                  </>
                ) : (
                  <button
                    onClick={hangupCall}
                    className="w-14 h-14 rounded-full bg-red-500 hover:bg-red-600 text-white flex items-center justify-center shadow-lg"
                  >
                    <Phone className="w-6 h-6" />
                  </button>
                )}
              </div>
            </div>
          ) : (
            <div className="relative w-full h-full rounded-3xl overflow-hidden border border-white/10 shadow-2xl bg-gradient-to-br from-[#0b1220] via-[#12243b] to-[#1a2f4d]">
              <div className="absolute inset-0 opacity-30 bg-[radial-gradient(circle_at_30%_20%,rgba(34,197,94,0.35),transparent_35%),radial-gradient(circle_at_70%_80%,rgba(14,165,233,0.35),transparent_35%)]" />
              <div className="relative z-10 h-full flex flex-col items-center justify-center px-6">
                <Av url={callPeerSession?.avatar} name={callPeerSession?.name || '对方'} cls="w-24 h-24 rounded-full ring-4 ring-white/20" />
                <p className="mt-4 text-white text-lg font-semibold">{callPeerSession?.name || '对方'}</p>
                <p className="mt-1 text-white/80 text-sm">
                  {callStatus === 'in-call' ? formatCallDuration(callSeconds) : callStatus === 'calling' ? '呼叫中…' : '来电中…'}
                </p>
              </div>

              <div className="absolute left-1/2 -translate-x-1/2 bottom-6 flex items-center gap-3 z-20">
                {incomingOffer && callStatus === 'ringing' ? (
                  <>
                    <button
                      onClick={rejectIncomingCall}
                      className="px-4 py-2 rounded-full bg-gray-700/90 hover:bg-gray-700 text-white text-sm"
                    >
                      拒绝
                    </button>
                    <button
                      onClick={acceptIncomingCall}
                      className="px-4 py-2 rounded-full bg-[#07c160] hover:bg-[#06ad55] text-white text-sm"
                    >
                      接听
                    </button>
                  </>
                ) : (
                  <button
                    onClick={hangupCall}
                    className="w-14 h-14 rounded-full bg-red-500 hover:bg-red-600 text-white flex items-center justify-center shadow-lg"
                  >
                    <Phone className="w-6 h-6" />
                  </button>
                )}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
