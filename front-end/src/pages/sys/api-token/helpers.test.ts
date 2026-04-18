import dayjs from 'dayjs';

import {
  buildPermissionSummary,
  buildTokenFormInitialValues,
  buildTokenSubmitPayload,
} from './helpers';

describe('api-token helpers', () => {
  it('summarizes authorized apis with overflow metadata', () => {
    const summary = buildPermissionSummary(
      [
        { ID: 1, method: 'GET', path: '/a' },
        { ID: 2, method: 'POST', path: '/b' },
        { ID: 3, method: 'DELETE', path: '/c' },
      ],
      2,
    );

    expect(summary.total).toBe(3);
    expect(summary.visible.map((item) => item.label)).toEqual(['[GET] /a', '[POST] /b']);
    expect(summary.hidden.map((item) => item.label)).toEqual(['[DELETE] /c']);
  });

  it('builds create form initial values with safe defaults', () => {
    expect(buildTokenFormInitialValues()).toEqual({
      maxConcurrency: 5,
      apiIds: [],
    });
  });

  it('builds update payload with normalized expiresAt and api ids', () => {
    const expiresAt = dayjs('2026-04-19 12:30:00');

    expect(
      buildTokenSubmitPayload(
        {
          name: 'server-token',
          description: 'used by scripts',
          maxConcurrency: 8,
          expiresAt: expiresAt.toISOString(),
          apiIds: [2, 5],
        },
        42,
      ),
    ).toEqual({
      id: 42,
      name: 'server-token',
      description: 'used by scripts',
      maxConcurrency: 8,
      expiresAt: expiresAt.toISOString(),
      apiIds: [2, 5],
    });
  });
});
