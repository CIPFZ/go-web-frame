import type { RequestOptions } from '@@/plugin-request/request';
import type { RequestConfig } from '@umijs/max';

import { isBackendRequest, renewedToken } from './utils/authHeaders';

const TOKEN_KEY = 'token';
const HEADER_TOKEN_KEY = 'x-token';

export const errorConfig: RequestConfig = {
  timeout: 10000,
  requestInterceptors: [
    (config: RequestOptions) => {
      if (!isBackendRequest(config.url, config.baseURL) || config.url?.includes('/user/login')) {
        return config;
      }

      const token = localStorage.getItem(TOKEN_KEY);
      if (!token) {
        return config;
      }

      if (config.headers instanceof Headers) {
        config.headers.set(HEADER_TOKEN_KEY, token);
      } else {
        config.headers = {
          ...(config.headers || {}),
          [HEADER_TOKEN_KEY]: token,
        };
      }

      return config;
    },
  ],
  responseInterceptors: [
    (response) => {
      const newToken = isBackendRequest(response.config?.url, response.config?.baseURL)
        ? renewedToken(response.headers) : undefined;
      if (newToken) {
        localStorage.setItem(TOKEN_KEY, newToken);
      }
      return response;
    },
  ],
};
