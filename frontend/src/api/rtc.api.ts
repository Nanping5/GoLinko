import apiClient from './client';
import type { ApiResponse } from './types';

export interface RTCIceServerResp {
  urls: string[];
  username?: string;
  credential?: string;
}

export interface RTCIceServersData {
  iceServers: RTCIceServerResp[];
  ttl?: number;
}

export const rtcApi = {
  getIceServers: (ttl = 3600) =>
    apiClient.get<any, ApiResponse<RTCIceServersData>>(`/v1/rtc_ice_servers?ttl=${ttl}`),
};
