// AxiosHeaders, Fetch Headers and plain response-header objects are all supported.
export function renewedToken(headers: unknown): string | undefined {
  if (!headers || typeof headers !== 'object') return undefined;
  const getter = (headers as { get?: (name: string) => unknown }).get;
  const value = typeof getter === 'function'
    ? getter.call(headers, 'new-token')
    : Object.entries(headers).find(([key]) => key.toLowerCase() === 'new-token')?.[1];
  return typeof value === 'string' && value.trim() ? value : undefined;
}

export function isBackendRequest(url?: string, baseURL?: string): boolean {
  if (!url) return false;
  try {
    const target = new URL(url, baseURL ? new URL(baseURL, window.location.origin) : window.location.origin);
    return target.origin === window.location.origin && target.pathname.startsWith('/api/');
  } catch {
    return false;
  }
}
