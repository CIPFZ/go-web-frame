import { useEffect, useState } from 'react';
import {
  getMyNotices,
  type MyNotice,
  type MyNoticePage,
} from '@/services/system/notice';

export const noticePageSize = 5;

export function useNotices() {
  const [page, setPage] = useState(1);
  const [revision, setRevision] = useState(0);
  const [data, setData] = useState<MyNoticePage & { list: MyNotice[] }>();
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    const refresh = () => setRevision((n) => n + 1);
    window.addEventListener('cms:notices-changed', refresh);
    return () => window.removeEventListener('cms:notices-changed', refresh);
  }, []);
  const [error, setError] = useState(false);
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(false);
    void getMyNotices({ page, pageSize: noticePageSize })
      .then((result) => {
        if (cancelled) return;
        if (result.code !== 0) throw new Error(result.msg);
        // A notice may expire or be deleted while the last page is open.
        const lastPage = Math.max(
          1,
          Math.ceil(result.data.total / noticePageSize),
        );
        if (page > lastPage) {
          setPage(lastPage);
          return;
        }
        // Go serializes an empty receiver slice as null.
        setData({ ...result.data, list: result.data.list ?? [] });
      })
      .catch(() => {
        if (!cancelled) setError(true);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [page, revision]);
  return {
    page,
    data,
    loading,
    error,
    changePage: (next: number) => {
      setLoading(true);
      setPage(next);
    },
    refresh: () => {
      setLoading(true);
      setRevision((value) => value + 1);
    },
  };
}
