import { formatDate, t, useI18n } from '@/i18n';
import { markNoticeRead, type MyNotice } from '@/services/system/notice';
import { getServerState } from '@/services/system/state';
import { EditOutlined, PlusOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useModel, useRequest } from '@umijs/max';
import { Alert, Button, Modal, Space, Tag, message } from 'antd';
import React, { useMemo, useRef, useState } from 'react';
import Desk from './Desk';
import Gallery from './Gallery';
import {
  type Desk as DeskLayout,
  type DeskItem,
  defaultDesk,
  resolveDesk,
  roleDesk,
  saveDesk,
} from './layout';
import './setup';
import { useDeskPages } from './useDeskPages';
import { useStyles } from './styles';
import { useNotices } from './useNotices';
import { DeskContext, type DeskContextValue } from './widgets/context';

// 工作台：本人的桌面（小组件与应用图标，见 Desk.tsx）。
// “编辑工作台”进入编辑模式：拖动、移除、切换尺寸、添加；“完成”保存到本人账号，“取消”放弃。
export default function Workplace() {
  useI18n();
  const levels = {
    info: {
      label: t('cms.notice'),
      color: 'blue',
    },
    warning: {
      label: t('cms.reminder'),
      color: 'orange',
    },
    error: {
      label: t('cms.important'),
      color: 'red',
    },
  };
  const { styles } = useStyles();
  const { initialState, setInitialState } = useModel('@@initialState');
  const user = initialState?.currentUser;
  const {
    menus,
    pages,
    allPages,
    links,
    pageByPath,
    canWidget,
    pathOf,
    resolveItems,
  } = useDeskPages();
  const statePage = pages.find((menu) => menu.component === 'state');
  const accountPage = allPages.find((menu) => menu.component === 'user/info');
  // 默认：本人角色的默认工作台（管理员在角色管理里设计，没设计过用预置的）；都没有时按本人能看到的内容起步
  // 预置里按组件标识指向的页面换成实际地址
  const fallback = () => {
    const desk =
      roleDesk(user?.authorities || []) ??
      defaultDesk({ widget: canWidget, page: () => true });
    return { ...desk, items: resolveItems(desk.items) };
  };
  const saved = useMemo(
    () => {
      const stored = resolveDesk(user?.workplace);
      return stored
        ? { ...stored, items: resolveItems(stored.items) }
        : fallback();
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [user?.workplace, user?.authorities, pageByPath],
  );
  // 编辑时的草稿：改动立即生效，“完成”才保存
  const [draft, setDraft] = useState<DeskItem[] | null>(null);
  const [adding, setAdding] = useState<'widgets' | 'apps' | null>(null);
  const [saving, setSaving] = useState(false);
  const items = draft ?? saved.items;
  const editing = draft !== null;
  const serverQuery = useRequest(
    async () => {
      const result = await getServerState();
      if (result.code !== 0)
        throw new Error(result.msg || t('cms.unableToLoadStatus'));
      return {
        data: {
          ...result.data,
          checkedAt: new Date(),
        },
      };
    },
    {
      ready: !!statePage,
      pollingInterval: 15000,
      pollingWhenHidden: false,
    },
  );
  const checks = serverQuery.data?.server?.checks as
    | Record<string, string>
    | undefined;
  const notices = useNotices();
  const [selected, setSelected] = useState<MyNotice>();
  const [marking, setMarking] = useState(false);
  const markingRef = useRef(false);
  const [messageApi, messageContext] = message.useMessage();
  const role =
    user?.authorities
      ?.map((a) => a.authorityName)
      .filter(Boolean)
      .join(' / ') ||
    user?.authority?.authorityName ||
    t('cms.noRoleAssigned');
  const confirmRead = async () => {
    if (!selected || markingRef.current) return;
    markingRef.current = true;
    setMarking(true);
    try {
      const result = await markNoticeRead({
        noticeId: selected.ID,
      });
      if (result.code !== 0)
        throw new Error(result.msg || t('cms.operationFailed'));
      messageApi.success(t('cms.markedAsRead'));
      setSelected(undefined);
      notices.refresh();
    } catch {
      messageApi.error(t('cms.unableToMarkAsReadPlease'));
    } finally {
      markingRef.current = false;
      setMarking(false);
    }
  };
  const finish = async () => {
    if (!draft) return;
    const next: DeskLayout = { v: 1, items: draft };
    setSaving(true);
    try {
      await saveDesk(next);
      setInitialState((s) =>
        s?.currentUser
          ? { ...s, currentUser: { ...s.currentUser, workplace: next } }
          : s,
      );
      setDraft(null);
      setAdding(null);
    } catch {
      messageApi.error(t('cms.desk.saveFailed'));
    } finally {
      setSaving(false);
    }
  };
  const desk: DeskContextValue = {
    notices,
    openNotice: setSelected,
    server: {
      checks,
      checkedAt: serverQuery.data?.checkedAt,
      error: !!serverQuery.error,
      loading: serverQuery.loading,
      refresh: serverQuery.refresh,
    },
    user,
    role,
    pathOf,
    accountPath: accountPage?.path,
    statePath: statePage?.path,
    pagesCount: pages.length,
    openGallery: (tab) => {
      if (!draft) setDraft(saved.items);
      setAdding(tab);
    },
  };
  return (
    <PageContainer
      title={t('cms.workplace')}
      subTitle={t('cms.startYourDayHere')}
      extra={
        editing ? (
          <Space>
            <Button
              icon={<PlusOutlined />}
              onClick={() => setAdding('widgets')}
            >
              {t('cms.desk.addTitle')}
            </Button>
            <Button onClick={() => setDraft(fallback().items)}>
              {t('cms.desk.reset')}
            </Button>
            <Button
              onClick={() => {
                setDraft(null);
                setAdding(null);
              }}
            >
              {t('cms.cancel')}
            </Button>
            <Button type="primary" loading={saving} onClick={finish}>
              {t('cms.desk.done')}
            </Button>
          </Space>
        ) : (
          <Button icon={<EditOutlined />} onClick={() => setDraft(saved.items)}>
            {t('cms.desk.edit')}
          </Button>
        )
      }
    >
      {messageContext}
      <DeskContext.Provider value={desk}>
        <div className={styles.stack}>
          {editing ? (
            <Alert type="info" showIcon message={t('cms.desk.editHint')} />
          ) : null}
          <Desk
            items={items}
            editing={editing}
            pages={pageByPath}
            canWidget={(i) => canWidget(i.widget)}
            onChange={setDraft}
          />
        </div>
        <Gallery
          tab={adding}
          onTab={setAdding}
          menus={menus}
          links={links}
          items={items}
          canWidget={canWidget}
          onAdd={(item) => {
            setDraft([...items, item]);
            messageApi.success(t('cms.desk.added'));
          }}
          onClose={() => setAdding(null)}
        />
      </DeskContext.Provider>
      <Modal
        open={!!selected}
        title={selected?.title}
        onCancel={() => {
          if (!marking) setSelected(undefined);
        }}
        maskClosable={!marking}
        closable={!marking}
        keyboard={!marking}
        footer={
          <Space>
            <Button disabled={marking} onClick={() => setSelected(undefined)}>
              {t('cms.close')}
            </Button>
            {selected && !selected.readAt && (
              <Button
                type="primary"
                loading={marking}
                onClick={() => void confirmRead()}
              >
                {selected.needConfirm
                  ? t('cms.confirmAsRead')
                  : t('cms.markAsRead')}
              </Button>
            )}
          </Space>
        }
      >
        {selected && (
          <>
            <Space wrap>
              <Tag color={(levels[selected.level] || levels.info).color}>
                {(levels[selected.level] || levels.info).label}
              </Tag>
              <span className={styles.muted}>
                {formatDate(selected.createdAt)}
              </span>
              {selected.readAt && <Tag>{t('cms.read')}</Tag>}
            </Space>
            <div className={styles.noticeBody}>{selected.content}</div>
            {selected.needConfirm && !selected.readAt && (
              <Alert
                type="info"
                showIcon
                message={t('cms.pleaseConfirmThatYouHaveRead')}
              />
            )}
          </>
        )}
      </Modal>
    </PageContainer>
  );
}
