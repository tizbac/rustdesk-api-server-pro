import { request } from '@/service/request';

/**
 * Fetch server configuration
 */
export function fetchConfig() {
  return request({ url: '/config' });
}
