import { act, render } from '@testing-library/react';
import React from 'react';
import { useShared } from './shared';

jest.mock('@umijs/max', () => ({
  getLocale: () => 'zh-CN',
  request: jest.fn(),
}));

const h = React.createElement;

function Probe({
  k,
  load,
  onRender,
}: {
  k: string | null;
  load: () => Promise<number>;
  onRender: () => void;
}) {
  onRender();
  const { data } = useShared(k, load);
  return h('span', null, data ?? '-');
}

test('an idle widget renders once and two widgets share one request', async () => {
  let renders = 0;
  render(h(Probe, { k: null, load: async () => 1, onRender: () => renders++ }));
  expect(renders).toBe(1);

  const load = jest.fn(async () => 7);
  const { findAllByText } = render(
    h(
      React.Fragment,
      null,
      h(Probe, { k: 'same', load, onRender: () => undefined }),
      h(Probe, { k: 'same', load, onRender: () => undefined }),
    ),
  );
  await act(async () => undefined);
  expect(await findAllByText('7')).toHaveLength(2);
  expect(load).toHaveBeenCalledTimes(1);
});
