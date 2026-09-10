import { t } from '@/i18n';
type PageResponse<T> = {
  code: number;
  msg?: string;
  data?: { list?: T[]; total?: number };
};

// Selectors need all options, while each backend query remains bounded.
export async function loadPagedOptions<T>(
  load: (params: {
    page: number;
    pageSize: number;
  }) => Promise<PageResponse<T>>,
): Promise<PageResponse<T>> {
  const list: T[] = [];
  for (let page = 1; page <= 1000; page += 1) {
    const response = await load({ page, pageSize: 100 });
    if (response.code !== 0)
      throw new Error(response.msg || t('cms.loadOptionsFailed'));
    const items = response.data?.list || [];
    list.push(...items);
    const total = response.data?.total ?? list.length;
    if (list.length >= total)
      return { ...response, data: { ...response.data, list, total } };
    if (!items.length) throw new Error(t('cms.optionsChanged'));
  }
  throw new Error(t('cms.tooManyOptions'));
}
