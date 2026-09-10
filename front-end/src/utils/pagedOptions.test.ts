import { loadPagedOptions } from './pagedOptions';

test('loads every page when the backend caps page size at 100', async () => {
  const items = Array.from({ length: 205 }, (_, id) => ({ id }));
  const load = jest.fn(async ({ page, pageSize }) => ({
    code: 0, data: { total: items.length, list: items.slice((page - 1) * pageSize, page * pageSize) },
  }));
  expect((await loadPagedOptions(load)).data?.list).toEqual(items);
  expect(load).toHaveBeenCalledTimes(3);
  expect(load).toHaveBeenLastCalledWith({ page: 3, pageSize: 100 });
});

test('does not return a partial option list after a failed page', async () => {
  const load = jest.fn()
    .mockResolvedValueOnce({ code: 0, data: { total: 2, list: [{ id: 1 }] } })
    .mockResolvedValueOnce({ code: 7, msg: 'unavailable' });
  await expect(loadPagedOptions(load)).rejects.toThrow('unavailable');
});
