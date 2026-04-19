import { configUmiAlias, createConfig } from '@umijs/max/test.js';

export default async (): Promise<any> => {
  const config = await configUmiAlias(createConfig({ target: 'browser' }));
  return {
    ...config,
    setupFilesAfterEnv: ['<rootDir>/tests/setupTests.ts'],
    testEnvironmentOptions: {
      ...(config?.testEnvironmentOptions || {}),
      url: 'http://localhost:8000',
    },
  };
};
