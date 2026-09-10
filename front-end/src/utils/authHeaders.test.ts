import { isBackendRequest, renewedToken } from './authHeaders';

test('accepts renewed tokens from Axios, Fetch and plain headers', () => {
  expect(renewedToken({ 'new-token': 'fresh' })).toBe('fresh');
  expect(renewedToken({ 'New-Token': 'fresh' })).toBe('fresh');
  const axiosHeaders = { value: 'fresh', get(name: string) { return name === 'new-token' ? this.value : undefined; } };
  expect(renewedToken(axiosHeaders)).toBe('fresh');
  expect(renewedToken(new Headers({ 'new-token': 'fresh' }))).toBe('fresh');
  expect(renewedToken({ 'new-token': ['untrusted'] })).toBeUndefined();
  expect(renewedToken({})).toBeUndefined();
});

test('credentials are only used for same-origin backend requests', () => {
  expect(isBackendRequest('/api/v1/sys/menu/getMenu')).toBe(true);
  expect(isBackendRequest('/api/v1/sys/menu/getMenu', window.location.origin)).toBe(true);
  expect(isBackendRequest('https://external.example/api/v1')).toBe(false);
  expect(isBackendRequest('//external.example/api/v1')).toBe(false);
  expect(isBackendRequest('/api/v1', 'https://external.example')).toBe(false);
  expect(isBackendRequest('/uploads/file/a.txt')).toBe(false);
});
