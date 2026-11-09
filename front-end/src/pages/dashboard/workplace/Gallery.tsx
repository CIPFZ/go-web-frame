import { t } from '@/i18n';
import { menuLabel } from '@/i18n/menus';
import {
  AppstoreOutlined,
  CheckCircleFilled,
  PlusOutlined,
} from '@ant-design/icons';
import { Button, Drawer, Segmented, Tabs, Typography } from 'antd';
import { createStyles } from 'antd-style';
import React, { useState } from 'react';
import { SIZE_LABEL, WidgetTile } from './Desk';
import { type DeskItem, appItem, newItemId } from './layout';
import { type WorkplaceMenu, pageGroups } from './menus';
import {
  WIDGET_SIZES,
  type WidgetSize,
  appTint,
  listWidgets,
  widgetGroups,
} from './registry';

const useStyles = createStyles(({ token }) => ({
  group: { marginBottom: 24 },
  groupLabel: {
    color: token.colorTextTertiary,
    fontSize: 12,
    margin: '0 0 10px',
  },
  card: {
    padding: 16,
    marginBottom: 16,
    borderRadius: 16,
    background: token.colorFillQuaternary,
  },
  cardHead: {
    display: 'flex',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    gap: 12,
    marginBottom: 12,
  },
  preview: { pointerEvents: 'none' },
  apps: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(96px, 1fr))',
    gap: 12,
  },
  app: {
    position: 'relative',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    gap: 6,
    padding: '12px 4px',
    border: 0,
    borderRadius: 14,
    background: 'transparent',
    cursor: 'pointer',
    fontSize: 12,
    color: token.colorText,
    '&:hover': { background: token.colorFillTertiary },
    '&:disabled': { cursor: 'default', opacity: 0.55 },
  },
  appIcon: {
    display: 'grid',
    placeItems: 'center',
    width: 44,
    height: 44,
    borderRadius: 12,
    color: '#fff',
    fontSize: 19,
  },
  added: { position: 'absolute', top: 6, right: 14, color: token.colorSuccess },
}));

const PREVIEW: Record<WidgetSize, React.CSSProperties> = {
  s: { width: 200, height: 192 },
  m: { width: '100%', height: 192 },
  l: { width: '100%', height: 400 },
  w: { width: '100%', height: 296 },
};

/** 添加小组件与应用图标：像手机的小组件库，每个小组件可以预览三种尺寸再添加；应用图标按菜单分组，已在桌面上的打勾。 */
export default function Gallery({
  tab,
  onTab,
  menus,
  links,
  items,
  canWidget,
  onAdd,
  onClose,
}: {
  /** 打开在哪个页签；null 是关着 */
  tab: 'widgets' | 'apps' | null;
  onTab: (tab: 'widgets' | 'apps') => void;
  menus: WorkplaceMenu[];
  /** 本人能放的外部链接，按应用图标的样子列在最后 */
  links: WorkplaceMenu[];
  items: DeskItem[];
  canWidget: (key: string) => boolean;
  onAdd: (item: DeskItem) => void;
  onClose: () => void;
}) {
  const { styles } = useStyles();
  const [sizes, setSizes] = useState<Record<string, WidgetSize>>({});
  const onDesk = new Set(
    items.flatMap((i) => (i.kind === 'app' ? [i.path] : [])),
  );
  const widgets = (
    <>
      {widgetGroups().map(({ key: group, label }) => {
        const defs = listWidgets().filter(
          (w) => w.group === group && canWidget(w.key),
        );
        if (!defs.length) return null;
        return (
          <div key={group} className={styles.group}>
            <div className={styles.groupLabel}>{label()}</div>
            {defs.map((def) => {
              const size = sizes[def.key] ?? def.sizes[0];
              return (
                <section
                  key={def.key}
                  className={styles.card}
                  aria-label={def.title()}
                >
                  <div className={styles.cardHead}>
                    <div>
                      <Typography.Text strong>{def.title()}</Typography.Text>
                      <div>
                        <Typography.Text
                          type="secondary"
                          style={{ fontSize: 12 }}
                        >
                          {def.description()}
                        </Typography.Text>
                      </div>
                    </div>
                    <Button
                      type="primary"
                      size="small"
                      icon={<PlusOutlined />}
                      onClick={() =>
                        onAdd({
                          id: newItemId(),
                          kind: 'widget',
                          widget: def.key,
                          size,
                        })
                      }
                    >
                      {t('cms.desk.add')}
                    </Button>
                  </div>
                  <Segmented<WidgetSize>
                    size="small"
                    value={size}
                    onChange={(v) => setSizes((s) => ({ ...s, [def.key]: v }))}
                    options={WIDGET_SIZES.filter((s) =>
                      def.sizes.includes(s),
                    ).map((s) => ({ value: s, label: SIZE_LABEL[s]() }))}
                    style={{ marginBottom: 12 }}
                  />
                  {/* 预览不可交互，读屏跳过 */}
                  <div
                    className={styles.preview}
                    style={PREVIEW[size]}
                    aria-hidden
                  >
                    <WidgetTile
                      item={{
                        id: 'preview',
                        kind: 'widget',
                        widget: def.key,
                        size,
                      }}
                      preview
                    />
                  </div>
                </section>
              );
            })}
          </div>
        );
      })}
    </>
  );
  const apps = (
    <>
      {[
        ...pageGroups(menus, t('cms.desk.groupOther')),
        ...(links.length
          ? [{ key: '_links', label: t('cms.desk.groupLinks'), pages: links }]
          : []),
      ].map((g) => (
        <div key={g.key} className={styles.group}>
          <div className={styles.groupLabel}>{g.label}</div>
          <div className={styles.apps}>
            {g.pages
              .filter((p) => p.component !== 'dashboard/workplace')
              .map((p) => {
                const added = onDesk.has(p.path);
                return (
                  <button
                    key={p.path}
                    type="button"
                    className={styles.app}
                    disabled={added}
                    onClick={() => onAdd(appItem(p.path))}
                  >
                    <span
                      className={styles.appIcon}
                      style={{
                        background: `linear-gradient(145deg, ${appTint(p.path).join(', ')})`,
                      }}
                    >
                      {p.icon || <AppstoreOutlined />}
                    </span>
                    {menuLabel(p)}
                    {added ? (
                      <CheckCircleFilled
                        className={styles.added}
                        aria-label={t('cms.desk.onDesk')}
                      />
                    ) : null}
                  </button>
                );
              })}
          </div>
        </div>
      ))}
    </>
  );
  return (
    <Drawer
      open={!!tab}
      width={620}
      title={t('cms.desk.addTitle')}
      onClose={onClose}
      destroyOnClose
    >
      <Tabs
        activeKey={tab || 'widgets'}
        onChange={(k) => onTab(k as 'widgets' | 'apps')}
        items={[
          { key: 'widgets', label: t('cms.desk.widgets'), children: widgets },
          { key: 'apps', label: t('cms.desk.apps'), children: apps },
        ]}
      />
    </Drawer>
  );
}
