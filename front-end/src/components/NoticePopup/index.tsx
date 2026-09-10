import { useEffect, useState } from 'react';
import { Alert, Button, Modal, Space, message } from 'antd';
import { t, useI18n } from '@/i18n';
import {
  getMyNotices,
  markNoticeRead,
  type MyNotice,
} from '@/services/system/notice';

// Mounted at the authenticated layout so a custom home route still gets notices.
// Dismissal never acknowledges a notice; it reappears on the next login/reload.
export default function NoticePopup({ userId }: { userId?: number }) {
  useI18n();
  const [queue, setQueue] = useState<MyNotice[]>([]);
  const [failed, setFailed] = useState(false);
  const [retry, setRetry] = useState(0);
  const [saving, setSaving] = useState(false);
  useEffect(() => {
    let cancelled = false;
    setQueue([]);
    setFailed(false);
    if (!userId) return;
    void (async () => {
      const notices: MyNotice[] = [];
      for (let page = 1; ; page++) {
        const res = await getMyNotices({
          page,
          pageSize: 100,
          popupOnly: true,
        });
        if (cancelled) return;
        if (res.code !== 0) throw new Error(res.msg);
        notices.push(...(res.data.list || []));
        if (
          page * res.data.pageSize >= res.data.total ||
          !res.data.list?.length
        )
          break;
      }
      if (!cancelled) setQueue(notices);
    })().catch(() => {
      if (!cancelled) setFailed(true);
    });
    return () => {
      cancelled = true;
    };
  }, [userId, retry]);
  const selected = queue[0];
  const next = () => setQueue((items) => items.slice(1));
  const confirm = async () => {
    if (!selected) return;
    setSaving(true);
    try {
      const res = await markNoticeRead({ noticeId: selected.ID });
      if (res.code !== 0) throw new Error(res.msg);
      window.dispatchEvent(new Event('cms:notices-changed'));
      next();
    } catch (error) {
      message.error(
        error instanceof Error ? error.message : t('cms.operationFailed'),
      );
    } finally {
      setSaving(false);
    }
  };
  return (
    <>
      {failed && (
        <Alert
          type="warning"
          message={t('cms.unableToLoadNotices')}
          action={
            <Button onClick={() => setRetry((n) => n + 1)}>
              {t('cms.retry')}
            </Button>
          }
        />
      )}
      <Modal
        open={!!selected}
        title={selected?.title}
        onCancel={next}
        closable={!saving}
        maskClosable={!saving}
        keyboard={!saving}
        footer={
          <Space>
            <Button disabled={saving} onClick={next}>
              {t('cms.close')}
            </Button>
            <Button type="primary" loading={saving} onClick={confirm}>
              {selected?.needConfirm
                ? t('cms.confirmAsRead')
                : t('cms.markAsRead')}
            </Button>
          </Space>
        }
      >
        <p style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>
          {selected?.content}
        </p>
        {selected?.needConfirm && (
          <Alert
            type="info"
            showIcon
            message={t('cms.pleaseConfirmThatYouHaveRead')}
          />
        )}
      </Modal>
    </>
  );
}
