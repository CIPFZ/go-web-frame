type PageResponse<T> = {
  code: number;
  msg?: string;
  data?: { list?: T[]; total?: number };
};

// Selectors need all options, while each backend query remains bounded.
export async function loadPagedOptions<T>(
  load: (params: { page: number; pageSize: number }) => Promise<PageResponse<T>>,
): Promise<PageResponse<T>> {
  const list: T[] = [];
  for (let page = 1; page <= 1000; page += 1) {
    const response = await load({ page, pageSize: 100 });
    if (response.code !== 0) throw new Error(response.msg || '加载选项失败');
    const items = response.data?.list || [];
    list.push(...items);
    const total = response.data?.total ?? list.length;
    if (list.length >= total) return { ...response, data: { ...response.data, list, total } };
    if (!items.length) throw new Error('选项数据发生变化，请重新加载');
  }
  throw new Error('选项数量过多，请缩小查询范围');
}
