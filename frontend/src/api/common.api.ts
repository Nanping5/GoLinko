import apiClient from './client';
import type { ApiResponse } from './types';

export const commonApi = {
  ping: () => apiClient.get<any, ApiResponse<null>>('/v1/ping'),
};
