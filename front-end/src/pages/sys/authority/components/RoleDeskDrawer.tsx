import { t, useI18n } from '@/i18n';
import Desk from '@/pages/dashboard/workplace/Desk';
import Gallery from '@/pages/dashboard/workplace/Gallery';
import '@/pages/dashboard/workplace/setup';
import {
  type DeskItem,
  defaultDesk,
  presetDesk,
  resolveDesk,
  saveRoleDesk,
} from '@/pages/dashboard/workplace/layout';
import { rolePriority } from '@/pages/dashboard/workplace/registry';
import type { WorkplaceMenu } from '@/pages/dashboard/workplace/menus';
import { useDeskPages } from '@/pages/dashboard/workplace/useDeskPages';
import { processMenuData } from '@/routing/menuHelpers';
import { getMenuAuthority, getMenuList } from '@/services/system/menu';
import { useNotices } from '@/pages/dashboard/workplace/useNotices';
import {
  DeskContext,
  type DeskContextValue,
} from '@/pages/dashboard/workplace/widgets/context';
import { PlusOutlined } from '@ant-design/icons';
import { useModel } from '@umijs/max';
import {
  Alert,
  Button,
  Drawer,
  Popconfirm,
  Space,
  Spin,
  Tag,
  message,
} from 'antd';
import React, { useEffect, useState } from 'react';
import type { AuthorityItem } from '../index';

/** 角色的默认工作台状态：设计过（存在角色上）、预置、或没有（用通用的起步桌面）。 */
export function roleDeskState(
  role: Pick<AuthorityItem, 'authorityId' | 'workplace'>,
): 'custom' | 'preset' | 'none' {
  if (resolveDesk(role.workplace)) return 'custom';
  return presetDesk(role.authorityId) ? 'preset' : 'none';
}

export function RoleDeskTag({
  role,
}: {
  role: Pick<AuthorityItem, 'authorityId' | 'workplace'>;
}) {
  const state = roleDeskState(role);
  const label = {
    custom: t('cms.roleDesk.custom'),
    preset: t('cms.roleDesk.preset'),
    none: t('cms.roleDesk.none'),
  }[state];
  return (
    <Tag
      bordered={false}
      color={state === 'custom' ? 'blue' : undefined}
      style={{ marginInlineEnd: 0 }}
    >
      {label}
    </Tag>
  );
}

type RawMenu = { ID: number; routes?: RawMenu[] };

/** 只留这些菜单（角色被授权的）：没授权的菜单连同它的子菜单去掉。 */
const keepMenus = (tree: RawMenu[], ids: Set<number>): RawMenu[] =>
  tree
    .filter((m) => ids.has(m.ID))
    .map((m) => ({
      ...m,
      routes: m.routes ? keepMenus(m.routes, ids) : undefined,
    }));

const listOf = (res: any): any[] => {
  if (res?.code !== 0) throw new Error(res?.msg);
  return Array.isArray(res.data) ? res.data : res.data?.list || [];
};

/** 这个角色被授权的菜单树（和侧栏一样的结构）。 */
function useRoleMenus(authorityId: number) {
  const [menus, setMenus] = useState<WorkplaceMenu[] | null>(null);
  const [error, setError] = useState(false);
  useEffect(() => {
    let alive = true;
    Promise.all([
      getMenuList({ pageInfo: { page: 1, pageSize: 9999 } }),
      getMenuAuthority({ authorityId }),
    ])
      .then(([all, granted]) => {
        const ids = new Set(listOf(granted).map((m: RawMenu) => m.ID));
        if (alive) setMenus(processMenuData(keepMenus(listOf(all), ids)));
      })
      .catch(() => alive && setError(true));
    return () => {
      alive = false;
    };
  }, [authorityId]);
  return { menus, error };
}

/**
 * 设计角色的默认工作台：和工作台的编辑模式一样拖动、移除、切换尺寸、从组件库添加。
 * 只能放这个角色有权限的页面与小组件（外部链接人人可用）；没改过工作台的人（以及“恢复默认”）用它，多个角色时各角色的拼在一起。
 */
export default function RoleDeskDrawer({
  role,
  roleName,
  onClose,
  onSaved,
}: {
  role: AuthorityItem;
  roleName: (id: number) => string;
  onClose: () => void;
  onSaved: () => void;
}) {
  useI18n();
  const { menus, error } = useRoleMenus(role.authorityId);
  if (!menus) {
    return (
      <Drawer
        open
        width="min(1280px, 96vw)"
        title={t('cms.roleDesk.title', { name: role.authorityName })}
        onClose={onClose}
      >
        {error ? (
          <Alert type="error" showIcon message={t('cms.roleDesk.loadFailed')} />
        ) : (
          <Spin style={{ display: 'block', marginTop: 80 }} />
        )}
      </Drawer>
    );
  }
  return (
    <RoleDeskEditor
      role={role}
      roleName={roleName}
      menus={menus}
      onClose={onClose}
      onSaved={onSaved}
    />
  );
}

function RoleDeskEditor({
  role,
  roleName,
  menus: roleMenus,
  onClose,
  onSaved,
}: {
  role: AuthorityItem;
  roleName: (id: number) => string;
  menus: WorkplaceMenu[];
  onClose: () => void;
  onSaved: () => void;
}) {
  useI18n();
  const { initialState } = useModel('@@initialState');
  const {
    menus,
    pages,
    allPages,
    links,
    pageByPath,
    canWidget,
    pathOf,
    resolveItems,
  } = useDeskPages(roleMenus);
  // 只留这个角色能用的（存下来的设计里，角色之后被收回的页面也一起去掉）；预置里的组件标识换成实际地址
  const usable = (list: DeskItem[]) =>
    resolveItems(list).filter((i) =>
      i.kind === 'app' ? pageByPath.has(i.path) : canWidget(i.widget),
    );
  const generic = () => defaultDesk({ widget: canWidget, page: () => true });
  const preset = () => presetDesk(role.authorityId) ?? generic();
  const [items, setItems] = useState<DeskItem[]>(() =>
    usable((resolveDesk(role.workplace) ?? preset()).items),
  );
  // 是预置的那一版（没设计过，或点了“恢复预置”还没再改）：保存时清掉角色上存的
  const [isPreset, setIsPreset] = useState(!resolveDesk(role.workplace));
  const [adding, setAdding] = useState<'widgets' | 'apps' | null>(null);
  const [saving, setSaving] = useState(false);
  const notices = useNotices();
  const change = (next: DeskItem[]) => {
    setItems(next);
    setIsPreset(false);
  };
  const save = async () => {
    setSaving(true);
    try {
      await saveRoleDesk(
        role.authorityId,
        isPreset ? null : { v: 1, items: usable(items) },
      );
      message.success(t('cms.roleDesk.saved'));
      onSaved();
      onClose();
    } catch (e) {
      message.error((e as Error)?.message || t('cms.desk.saveFailed'));
    } finally {
      setSaving(false);
    }
  };
  // 小组件在这里只是预览（内容是当前登录人的数据）
  const ctx: DeskContextValue = {
    notices,
    openNotice: () => {},
    server: { error: false, loading: false, refresh: () => {} },
    user: initialState?.currentUser,
    role: role.authorityName,
    pathOf,
    accountPath: allPages.find((m) => m.component === 'user/info')?.path,
    statePath: pages.find((m) => m.component === 'state')?.path,
    pagesCount: pages.length,
    openGallery: setAdding,
  };
  // 多个角色合并时的顺序：登记过优先顺序的角色在前，其余按编号排在后面
  const priority = rolePriority().map(roleName).join(' > ');
  return (
    <Drawer
      open
      width="min(1280px, 96vw)"
      title={
        <Space size={10}>
          {t('cms.roleDesk.title', { name: role.authorityName })}
          {isPreset ? (
            <Tag bordered={false}>
              {t(
                presetDesk(role.authorityId)
                  ? 'cms.roleDesk.preset'
                  : 'cms.roleDesk.none',
              )}
            </Tag>
          ) : (
            <Tag bordered={false} color="blue">
              {t('cms.roleDesk.custom')}
            </Tag>
          )}
        </Space>
      }
      onClose={onClose}
      maskClosable={false}
      extra={
        <Space>
          <Button icon={<PlusOutlined />} onClick={() => setAdding('widgets')}>
            {t('cms.desk.addTitle')}
          </Button>
          <Popconfirm
            title={t('cms.roleDesk.restoreConfirm')}
            onConfirm={() => {
              setItems(usable(preset().items));
              setIsPreset(true);
            }}
          >
            <Button disabled={isPreset}>{t('cms.roleDesk.restore')}</Button>
          </Popconfirm>
          <Button onClick={onClose}>{t('cms.cancel')}</Button>
          <Button type="primary" loading={saving} onClick={save}>
            {t('cms.roleDesk.save')}
          </Button>
        </Space>
      }
    >
      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 20 }}
        message={t('cms.roleDesk.hint')}
        description={t('cms.roleDesk.priority', { order: priority })}
      />
      <DeskContext.Provider value={ctx}>
        <Desk
          items={items}
          editing
          pages={pageByPath}
          canWidget={(i) => canWidget(i.widget)}
          onChange={change}
        />
        <Gallery
          tab={adding}
          onTab={setAdding}
          menus={menus}
          links={links}
          items={items}
          canWidget={canWidget}
          onAdd={(item) => {
            change([...items, item]);
            message.success(t('cms.desk.added'));
          }}
          onClose={() => setAdding(null)}
        />
      </DeskContext.Provider>
    </Drawer>
  );
}
