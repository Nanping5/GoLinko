const LOCAL_HOST_SET = new Set(['localhost', '127.0.0.1']);
const DEFAULT_ICE_SERVERS: RTCIceServer[] = [
  {
    urls: [
      'stun:stun.l.google.com:19302',
      'stun:stun1.l.google.com:19302',
    ],
  },
];

const getFallbackBaseUrl = () => {
  if (typeof window === 'undefined') return 'http://localhost:8080';
  const { protocol, hostname } = window.location;

  // 非本机访问时，默认回退到当前站点同源，便于 frpc/反向代理联调
  if (!LOCAL_HOST_SET.has(hostname)) {
    return `${protocol}//${window.location.host}`;
  }

  return 'http://localhost:8080';
};

export const getApiBaseUrl = () => {
  const fromEnv = import.meta.env.VITE_API_URL;
  const base = (fromEnv && fromEnv.trim()) || getFallbackBaseUrl();
  return base.replace(/\/$/, '');
};

export const getWsBaseUrl = () => {
  return getApiBaseUrl().replace(/^https/, 'wss').replace(/^http/, 'ws');
};

export const getWebRTCIceServers = (): RTCIceServer[] => {
  return [...DEFAULT_ICE_SERVERS];
};
