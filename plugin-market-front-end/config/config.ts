import { defineConfig } from '@umijs/max';

export default defineConfig({
  routes: [
    { path: '/', redirect: '/plugins' },
    { path: '/plugins', component: '@/pages/plugin-list' },
    { path: '/plugins/:id', component: '@/pages/plugin-detail' },
  ],
  history: { type: 'browser' },
  title: 'Plugin Market',
  hash: true,
  publicPath: '/',
  layout: false,
  antd: {
    configProvider: {
      theme: {
        cssVar: true,
      },
    },
  },
  proxy: {
    '/api/': {
      target: 'http://127.0.0.1:18081',
      changeOrigin: true,
    },
  },
  request: {},
  npmClient: 'npm',
});
