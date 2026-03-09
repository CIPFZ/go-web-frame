import type { RequestOptions } from '@@/plugin-request/request';
import type { RequestConfig } from '@umijs/max';

const TOKEN_KEY = 'token';
const HEADER_TOKEN_KEY = 'x-token';
const HEADER_NEW_TOKEN_KEY = 'new-token';

export const errorConfig: RequestConfig = {
  timeout: 10000,
  requestInterceptors: [
    (config: RequestOptions) => {
      if (config.url?.includes('/user/login')) {
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
      const newToken = response.headers?.get?.(HEADER_NEW_TOKEN_KEY);
      if (newToken) {
        localStorage.setItem(TOKEN_KEY, newToken);
      }
      return response;
    },
  ],
};
