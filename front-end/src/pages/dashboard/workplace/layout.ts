import { request } from '@umijs/max';
import {
  type DeskPreset,
  type WidgetSize,
  rolePreset,
  rolePriority,
  starterDesk,
  widgetDef,
} from './registry';

// 工作台桌面：小组件与应用图标混排在一张网格上（见 Desk.tsx）。
// 应用图标占 1×1 格，打开一个页面或外部链接。
// 布局保存在本人账号上（sys_users.workplace），换电脑也一样；角色的默认工作台存在角色上（sys_authorities.workplace）。

export type { WidgetSize } from './registry';

/** 桌面上的一项：小组件（kind = widget）或应用图标（kind = app，打开 path：平台页面，或 https:// 外部链接）。 */
export type DeskItem =
  | { id: string; kind: 'widget'; widget: string; size: WidgetSize }
  | { id: string; kind: 'app'; path: string };

export interface Desk {
  v: 1;
  items: DeskItem[];
}

const MAX_ITEMS = 60;

let seq = 0;
export const newItemId = () =>
  `i${Date.now().toString(36)}${(seq++).toString(36)}`;

/** 小组件可用的尺寸（第一个是默认尺寸）；不认识的小组件返回空。 */
export const widgetSizes = (widget: string): WidgetSize[] =>
  widgetDef(widget)?.sizes ?? [];

export const widgetItem = (
  widget: string,
  size: WidgetSize = widgetSizes(widget)[0],
): DeskItem => ({ id: newItemId(), kind: 'widget', widget, size });
export const appItem = (path: string): DeskItem => ({
  id: newItemId(),
  kind: 'app',
  path,
});

const fromPreset = (p: DeskPreset): Desk => ({
  v: 1,
  items: [
    ...p.widgets
      .filter(([w]) => widgetDef(w))
      .map(([w, size]) => widgetItem(w, size)),
    ...p.apps.map(appItem),
  ],
});

/** 起步桌面（角色都没有默认工作台时）：只放本人能用的。 */
export function defaultDesk(can: {
  widget: (w: string) => boolean;
  page: (path: string) => boolean;
}): Desk {
  const p = starterDesk();
  return fromPreset({
    widgets: p.widgets.filter(([w]) => can.widget(w)),
    apps: p.apps.filter(can.page),
  });
}

// 按角色的默认工作台：没改过工作台的人（以及“恢复默认”）用本人角色的默认工作台。
// 管理员在“角色管理 → 默认工作台”里设计（存在角色上），只能放这个角色有权限的内容；没设计过的用预置的那一版。
// 一个人有多个角色时，把各角色的默认工作台按顺序拼起来（重复的小组件、图标只留第一个）。

/** 角色预置的默认工作台；没有预置的角色返回 null。 */
export function presetDesk(roleId: number): Desk | null {
  const p = rolePreset(roleId);
  return p ? fromPreset(p) : null;
}

const rank = (id: number) => {
  const order = rolePriority();
  const i = order.indexOf(id);
  return i < 0 ? order.length + id : i;
};

/** 按优先顺序排好的角色。 */
export const sortRoles = <T extends { authorityId: number }>(roles: T[]) =>
  [...roles].sort((a, b) => rank(a.authorityId) - rank(b.authorityId));

/** 本人角色的默认工作台：按优先顺序把各角色设计过的（没设计过用预置）拼起来，同一个小组件、图标只留第一个；都没有返回 null。 */
export function roleDesk(
  roles: { authorityId: number; workplace?: unknown }[],
): Desk | null {
  const items: DeskItem[] = [];
  const seen = new Set<string>();
  let found = false;
  for (const r of sortRoles(roles)) {
    const desk = resolveDesk(r.workplace) ?? presetDesk(r.authorityId);
    if (!desk) continue;
    found = true;
    for (const item of desk.items) {
      const key =
        item.kind === 'widget' ? `w:${item.widget}` : `a:${item.path}`;
      if (seen.has(key)) continue;
      seen.add(key);
      items.push({ ...item, id: newItemId() });
    }
  }
  return found ? { v: 1, items: items.slice(0, MAX_ITEMS) } : null;
}

/** 应用图标能打开的：站内 path、页面的组件标识（如 sys/user，显示时换成实际地址）、https:// 外部链接。 */
const isAppPath = (p: string) =>
  (p.startsWith('/') && !p.startsWith('//')) ||
  p.startsWith('https://') ||
  /^[a-z][\w-]*(\/[\w-]+)*$/i.test(p);

/** 把存下的桌面整理成有效的：去掉不认识的小组件、重复的应用、不支持的尺寸；没存过（或不认识的格式）返回 null。 */
export function resolveDesk(stored: unknown): Desk | null {
  if (!stored || typeof stored !== 'object') return null;
  const raw = stored as Record<string, unknown>;
  if (raw.v !== 1 || !Array.isArray(raw.items)) return null;
  const items: DeskItem[] = [];
  const ids = new Set<string>();
  const paths = new Set<string>();
  for (const x of raw.items as Record<string, unknown>[]) {
    if (!x || typeof x !== 'object') continue;
    const id =
      typeof x.id === 'string' && x.id && !ids.has(x.id) ? x.id : newItemId();
    const sizes = typeof x.widget === 'string' ? widgetSizes(x.widget) : [];
    if (x.kind === 'widget' && sizes.length) {
      const size = sizes.includes(x.size as WidgetSize)
        ? (x.size as WidgetSize)
        : sizes[0];
      items.push({ id, kind: 'widget', widget: x.widget as string, size });
    } else if (
      x.kind === 'app' &&
      typeof x.path === 'string' &&
      isAppPath(x.path) &&
      !paths.has(x.path)
    ) {
      paths.add(x.path);
      items.push({ id, kind: 'app', path: x.path });
    } else continue;
    ids.add(id);
  }
  return { v: 1, items: items.slice(0, MAX_ITEMS) };
}

/** 保存本人的桌面。 */
export async function saveDesk(desk: Desk) {
  const res = await request<{ code: number; msg?: string }>(
    '/api/v1/sys/user/workplace',
    { method: 'PUT', data: { workplace: desk } },
  );
  if (res.code !== 0) throw new Error(res.msg);
}

/** 保存角色的默认工作台；null 恢复预置。 */
export async function saveRoleDesk(authorityId: number, desk: Desk | null) {
  const res = await request<{ code: number; msg?: string }>(
    '/api/v1/sys/authority/workplace',
    { method: 'PUT', data: { authorityId, workplace: desk } },
  );
  if (res.code !== 0) throw new Error(res.msg);
}
