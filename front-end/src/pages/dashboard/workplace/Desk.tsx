import { t } from '@/i18n';
import { menuLabel } from '@/i18n/menus';
import {
  DndContext,
  KeyboardSensor,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core';
import {
  SortableContext,
  arrayMove,
  rectSortingStrategy,
  sortableKeyboardCoordinates,
  useSortable,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { AppstoreOutlined, MinusOutlined } from '@ant-design/icons';
import { Link, history } from '@umijs/max';
import { Badge, Segmented } from 'antd';
import { createStyles, keyframes } from 'antd-style';
import React, { useContext } from 'react';
import { type DeskItem, widgetSizes } from './layout';
import type { WorkplaceMenu } from './menus';
import {
  WIDGET_SIZES,
  type WidgetSize,
  appTint,
  useAppBadge,
  widgetDef,
} from './registry';
import { DeskContext } from './widgets/context';

// 桌面网格：8 列（窄屏 4 列），每格 88px 高；应用图标 1×1，小组件 小 2×2、中 4×2、大 4×4；dense 排布填满空隙。
// 编辑模式像手机桌面：图块轻轻晃动，拖动换位置，左上角移除，右上角切换尺寸。
// 编辑模式下只能编辑布局：图块里没有链接（应用图标不渲染成链接，小组件内容设为 inert），点击、拖动都不会打开页面。

const wiggle = keyframes`
  0% { transform: rotate(-0.4deg); }
  50% { transform: rotate(0.4deg); }
  100% { transform: rotate(-0.4deg); }
`;

/** 每种图块占的列数与行数；宽（问候）占满一行（窄屏也是），3 格高。 */
export const SPAN: Record<WidgetSize | 'app', [number | 'row', number]> = {
  app: [1, 1],
  s: [2, 2],
  m: [4, 2],
  l: [4, 4],
  w: ['row', 3],
};

export const useDeskStyles = createStyles(({ token, css }) => ({
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(8, minmax(0, 1fr))',
    gridAutoRows: 88,
    gridAutoFlow: 'row dense',
    gap: 16,
    '@media (max-width: 1100px)': {
      gridTemplateColumns: 'repeat(4, minmax(0, 1fr))',
    },
  },
  cell: { position: 'relative', minWidth: 0, minHeight: 0 },
  editing: css`
    & > div {
      animation: ${wiggle} 0.32s ease-in-out infinite;
    }
    &:nth-child(2n) > div {
      animation-delay: -0.16s;
    }
  `,
  dragging: {
    zIndex: 5,
    '& > div': { animation: 'none !important', boxShadow: token.boxShadow },
  },
  widget: {
    height: '100%',
    display: 'flex',
    flexDirection: 'column',
    gap: 10,
    padding: '14px 16px',
    borderRadius: 20,
    background: token.colorBgContainer,
    border: `1px solid ${token.colorBorderSecondary}`,
    boxShadow:
      '0 1px 2px rgba(15, 23, 42, 0.04), 0 4px 16px rgba(15, 23, 42, 0.04)',
    overflow: 'hidden',
    cursor: 'pointer',
    transition: 'box-shadow .2s, transform .2s',
    '&:hover': {
      boxShadow:
        '0 2px 4px rgba(15, 23, 42, 0.06), 0 10px 28px rgba(15, 23, 42, 0.08)',
    },
  },
  head: { display: 'flex', alignItems: 'center', gap: 8, minHeight: 24 },
  headIcon: {
    display: 'grid',
    placeItems: 'center',
    width: 24,
    height: 24,
    borderRadius: 7,
    color: '#fff',
    fontSize: 13,
    flex: 'none',
  },
  headTitle: {
    fontSize: 13,
    fontWeight: 600,
    flex: 1,
    minWidth: 0,
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  app: {
    height: '100%',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 6,
    borderRadius: 16,
    cursor: 'pointer',
    textAlign: 'center',
    '&:hover > span:first-child': { transform: 'translateY(-2px)' },
    '&:focus-visible': { outline: `2px solid ${token.colorPrimary}` },
  },
  appIcon: {
    display: 'grid',
    placeItems: 'center',
    width: 50,
    height: 50,
    borderRadius: 14,
    color: '#fff',
    fontSize: 22,
    boxShadow: '0 4px 12px rgba(15, 23, 42, 0.14)',
    transition: 'transform .2s',
  },
  appLabel: {
    fontSize: 12,
    color: token.colorText,
    maxWidth: '100%',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
    padding: '0 4px',
  },
  remove: {
    position: 'absolute',
    top: -8,
    left: -8,
    zIndex: 2,
    display: 'grid',
    placeItems: 'center',
    width: 22,
    height: 22,
    borderRadius: '50%',
    border: `1px solid ${token.colorBorder}`,
    background: token.colorBgElevated,
    color: token.colorText,
    fontSize: 11,
    cursor: 'pointer',
    boxShadow: token.boxShadowTertiary,
    '&:hover': { background: token.colorErrorBg, color: token.colorError },
  },
  sizes: {
    position: 'absolute',
    top: 10,
    right: 10,
    zIndex: 2,
    margin: 0,
    padding: 0,
    border: 0,
    minWidth: 0,
  },
  empty: {
    gridColumn: '1 / -1',
    gridRow: 'span 2',
    display: 'grid',
    placeItems: 'center',
    borderRadius: 20,
    border: `1px dashed ${token.colorBorder}`,
    color: token.colorTextTertiary,
  },
}));

const gradient = ([a, b]: [string, string]) =>
  `linear-gradient(145deg, ${a}, ${b})`;

/** 一个小组件：标题行 + 内容；点卡片打开对应页面（标题是可聚焦的链接；编辑时不打开）。 */
export function WidgetTile({
  item,
  editing,
  preview,
}: {
  item: Extract<DeskItem, { kind: 'widget' }>;
  editing?: boolean;
  preview?: boolean;
}) {
  const { styles } = useDeskStyles();
  // 实际地址按后端菜单解析（小组件可以按组件标识指向页面）
  const to = useContext(DeskContext)?.pathOf(item.widget);
  const def = widgetDef(item.widget);
  if (!def) return null;
  const active = !editing && !preview && !!to;
  const inert = editing || preview;
  if (def.bare)
    return (
      <div
        style={{ height: '100%', pointerEvents: inert ? 'none' : undefined }}
        inert={inert}
      >
        <def.Body size={item.size} />
      </div>
    );
  return (
    <section
      className={styles.widget}
      aria-label={def.title()}
      onClick={() => active && history.push(to as string)}
    >
      <div className={styles.head}>
        <span
          className={styles.headIcon}
          style={{ background: gradient(def.tint) }}
        >
          {def.icon}
        </span>
        {active ? (
          <Link
            to={to as string}
            className={styles.headTitle}
            style={{ color: 'inherit' }}
            onClick={(e) => e.stopPropagation()}
          >
            {def.title()}
          </Link>
        ) : (
          <span className={styles.headTitle}>{def.title()}</span>
        )}
      </div>
      <div
        style={{
          flex: 1,
          minHeight: 0,
          display: 'flex',
          flexDirection: 'column',
          pointerEvents: inert ? 'none' : undefined,
        }}
        inert={inert}
      >
        <def.Body size={item.size} />
      </div>
    </section>
  );
}

/** 一个应用图标：圆角图标 + 名称，右上角数字角标；外部链接（https://）在新窗口打开。 */
export function AppTile({
  menu,
  editing,
}: {
  menu: WorkplaceMenu;
  editing?: boolean;
}) {
  const { styles } = useDeskStyles();
  const badge = useAppBadge(menu.path);
  const icon = (
    <>
      <Badge count={badge} size="small" offset={[-2, 2]}>
        <span
          className={styles.appIcon}
          style={{ background: gradient(appTint(menu.path)) }}
        >
          {menu.icon || <AppstoreOutlined />}
        </span>
      </Badge>
      <span className={styles.appLabel}>
        {menuLabel(menu)}
        {menu.path.startsWith('https://') ? ' ↗' : ''}
      </span>
    </>
  );
  if (editing) return <div className={styles.app}>{icon}</div>;
  if (menu.path.startsWith('https://'))
    return (
      <a
        href={menu.path}
        target="_blank"
        rel="noopener noreferrer"
        className={styles.app}
        aria-label={menuLabel(menu)}
        draggable={false}
      >
        {icon}
      </a>
    );
  return (
    <Link
      to={menu.path}
      className={styles.app}
      aria-label={menuLabel(menu)}
      draggable={false}
    >
      {icon}
    </Link>
  );
}

export const SIZE_LABEL: Record<WidgetSize, () => string> = {
  s: () => t('cms.desk.small'),
  m: () => t('cms.desk.medium'),
  l: () => t('cms.desk.large'),
  w: () => t('cms.desk.wide'),
};

function Cell({
  item,
  menu,
  editing,
  onRemove,
  onResize,
}: {
  item: DeskItem;
  menu?: WorkplaceMenu;
  editing: boolean;
  onRemove: () => void;
  onResize: (size: WidgetSize) => void;
}) {
  const { styles, cx } = useDeskStyles();
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: item.id, disabled: !editing });
  const [cols, rows] = SPAN[item.kind === 'app' ? 'app' : item.size];
  const name =
    item.kind === 'app'
      ? menuLabel(menu as WorkplaceMenu)
      : (widgetDef(item.widget)?.title() ?? item.widget);
  return (
    <div
      ref={setNodeRef}
      className={cx(
        styles.cell,
        editing && styles.editing,
        isDragging && styles.dragging,
      )}
      style={{
        gridColumn: cols === 'row' ? '1 / -1' : `span ${cols}`,
        gridRow: `span ${rows}`,
        transform: CSS.Translate.toString(transform),
        transition,
        cursor: editing ? 'grab' : undefined,
      }}
      {...(editing
        ? {
            ...attributes,
            ...listeners,
            'aria-label': t('cms.desk.dragOne', { name }),
          }
        : {})}
    >
      {editing ? (
        <button
          type="button"
          className={styles.remove}
          aria-label={t('cms.desk.removeOne', { name })}
          onPointerDown={(e) => e.stopPropagation()}
          onClick={onRemove}
        >
          <MinusOutlined />
        </button>
      ) : null}
      {editing &&
      item.kind === 'widget' &&
      widgetSizes(item.widget).length > 1 ? (
        <fieldset
          className={styles.sizes}
          aria-label={t('cms.desk.sizeOf', { name })}
          onPointerDown={(e) => e.stopPropagation()}
          onKeyDown={(e) => e.stopPropagation()}
        >
          <Segmented<WidgetSize>
            size="small"
            value={item.size}
            onChange={onResize}
            options={WIDGET_SIZES.filter((s) =>
              widgetSizes(item.widget).includes(s),
            ).map((s) => ({ value: s, label: SIZE_LABEL[s]() }))}
          />
        </fieldset>
      ) : null}
      <div style={{ height: '100%' }}>
        {item.kind === 'app' ? (
          <AppTile menu={menu as WorkplaceMenu} editing={editing} />
        ) : (
          <WidgetTile item={item} editing={editing} />
        )}
      </div>
    </div>
  );
}

/** 桌面：只显示本人能用的（应用图标需要页面权限；小组件需要它对应页面的权限）。 */
export default function Desk({
  items,
  editing,
  pages,
  canWidget,
  onChange,
}: {
  items: DeskItem[];
  editing: boolean;
  pages: Map<string, WorkplaceMenu>;
  canWidget: (item: Extract<DeskItem, { kind: 'widget' }>) => boolean;
  onChange: (items: DeskItem[]) => void;
}) {
  const { styles } = useDeskStyles();
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 6 } }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );
  const shown = items.filter((i) =>
    i.kind === 'app' ? pages.has(i.path) : canWidget(i),
  );
  const onDragEnd = ({ active, over }: DragEndEvent) => {
    if (!over || active.id === over.id) return;
    const from = items.findIndex((i) => i.id === active.id);
    const to = items.findIndex((i) => i.id === over.id);
    onChange(arrayMove(items, from, to));
  };
  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragEnd={onDragEnd}
    >
      <SortableContext
        items={shown.map((i) => i.id)}
        strategy={rectSortingStrategy}
      >
        <div className={styles.grid}>
          {shown.map((item) => (
            <Cell
              key={item.id}
              item={item}
              menu={item.kind === 'app' ? pages.get(item.path) : undefined}
              editing={editing}
              onRemove={() => onChange(items.filter((i) => i.id !== item.id))}
              onResize={(size) =>
                onChange(
                  items.map((i) =>
                    i.id === item.id && i.kind === 'widget'
                      ? { ...i, size }
                      : i,
                  ),
                )
              }
            />
          ))}
          {!shown.length ? (
            <div className={styles.empty}>{t('cms.desk.empty')}</div>
          ) : null}
        </div>
      </SortableContext>
    </DndContext>
  );
}
