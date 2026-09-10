// Jest setup file required by the Umi test config.

// Unit tests have no Umi bootstrap. Keep the real locale catalogs, React Intl
// provider and language events; only omit unrelated application runtime plugins.
jest.mock('@@/core/plugin', () => ({
  getPluginManager: () => ({
    applyPlugins: ({ initialValue }) => initialValue,
  }),
}));

if (typeof window !== 'undefined' && !window.matchMedia) {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: jest.fn().mockImplementation((query) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: jest.fn(),
      removeListener: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
    })),
  });
}
