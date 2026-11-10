import type React from 'react';

// 工作台的扩展点：小组件、组件库分组、应用图标的颜色与角标、外部链接、角色预置的默认工作台。
// 框架自带的小组件在 widgets/index.tsx 注册；业务项目在 src/workplace.tsx 注册自己的。
// 注册都在模块加载时完成（页面渲染前），之后不再变化。

/** 小组件尺寸：小 2×2、中 4×2、大 4×4 格；宽占满一行、3 格高。 */
export type WidgetSize = 's' | 'm' | 'l' | 'w';
export const WIDGET_SIZES: WidgetSize[] = ['s', 'm', 'l', 'w'];

export interface WidgetDef {
  /** 唯一的键，存进布局里，改了已保存的布局就找不到它 */
  key: string;
  /** 组件库里的分组（见 registerWidgetGroup） */
  group: string;
  title: () => string;
  description: () => string;
  icon: React.ReactNode;
  /** 图标底色（渐变的起点与终点） */
  tint: [string, string];
  /** 可用尺寸，第一个是默认尺寸 */
  sizes: WidgetSize[];
  /**
   * 点小组件打开的页面：path（以 / 开头），或页面的组件标识（如 'state'，按后端菜单找到实际地址，管理员改了菜单地址也能找到）。
   * 有这个页面的权限才能放这个小组件。
   */
  to?: string;
  /** 不需要页面权限：人人都能放（如当前账号） */
  public?: boolean;
  /** 自己画整张卡片（没有标题行与默认的边框） */
  bare?: boolean;
  Body: React.ComponentType<{ size: WidgetSize }>;
}

/** 外部链接（新窗口打开），像应用图标一样放到桌面上；不需要页面权限。 */
export interface DeskLink {
  url: string;
  name: () => string;
  icon: React.ReactNode;
  tint?: [string, string];
}

/** 一份预置的桌面：小组件（键与尺寸）在前，应用图标在后（页面 path、页面的组件标识如 'sys/user'，或 https:// 外部链接）。 */
export interface DeskPreset {
  widgets: [string, WidgetSize][];
  apps: string[];
}

const widgets = new Map<string, WidgetDef>();
const groups: { key: string; label: () => string }[] = [];
const tints = new Map<string, [string, string]>();
const tintPrefixes: [string, [string, string]][] = [];
const badges: ((path: string) => number)[] = [];
const links: DeskLink[] = [];
const presets = new Map<number, DeskPreset>();
let priority: number[] = [];
let starter: DeskPreset = { widgets: [], apps: [] };

export function registerWidget(def: WidgetDef) {
  if (!def.sizes.length) throw new Error(`widget ${def.key} has no sizes`);
  widgets.set(def.key, def);
}

/** 组件库里的分组，按注册顺序显示。 */
export function registerWidgetGroup(key: string, label: () => string) {
  if (!groups.some((g) => g.key === key)) groups.push({ key, label });
}

/** 应用图标的颜色：path 以 '/' 结尾时对这个前缀下的所有页面生效。 */
export function registerAppTint(path: string, tint: [string, string]) {
  if (path.endsWith('/')) tintPrefixes.push([path, tint]);
  else tints.set(path, tint);
}

/**
 * 应用图标右上角的数字（像手机上的未读数）。useBadge 是 React Hook，每次渲染都会对每个图标调用：
 * 不是自己负责的 path 返回 0，并且不要按 path 条件调用其他 Hook（可以给数据查询传空的 key）。
 */
export function registerAppBadge(useBadge: (path: string) => number) {
  badges.push(useBadge);
}

export function registerLink(link: DeskLink) {
  if (!link.url.startsWith('https://'))
    throw new Error('desk links must be https:// URLs');
  links.push(link);
}

/** 某个角色预置的默认工作台（管理员没在“角色管理 → 默认工作台”里设计过时用它）。 */
export function registerRolePreset(roleId: number, preset: DeskPreset) {
  presets.set(roleId, preset);
}

/** 一个人有多个角色时，各角色的默认工作台按这个顺序拼起来；不在列表里的角色排在后面（按编号）。 */
export function setRolePriority(roleIds: number[]) {
  priority = [...roleIds];
}

/** 所有角色都没有默认工作台时的起步桌面（只放本人能用的）。 */
export function setStarterDesk(preset: DeskPreset) {
  starter = preset;
}

export const listWidgets = () => [...widgets.values()];
export const widgetDef = (key: string) => widgets.get(key);
export const widgetGroups = () => groups;
export const deskLinks = () => links;
export const rolePreset = (roleId: number) => presets.get(roleId);
export const rolePriority = () => priority;
export const starterDesk = () => starter;

const DEFAULT_TINT: [string, string] = ['#94a3b8', '#64748b'];

export function appTint(path: string): [string, string] {
  const exact = tints.get(path) ?? links.find((l) => l.url === path)?.tint;
  if (exact) return exact;
  const prefix = tintPrefixes
    .filter(([p]) => path.startsWith(p))
    .sort((a, b) => b[0].length - a[0].length)[0];
  return prefix ? prefix[1] : DEFAULT_TINT;
}

/** 应用图标的数字：第一个非 0 的。所有角标 Hook 每次都按注册顺序调用。 */
export function useAppBadge(path: string): number {
  // biome-ignore lint/correctness/useHookAtTopLevel: 角标 Hook 在模块加载时登记、之后不变，每次渲染的调用顺序相同
  const values = badges.map((useBadge) => useBadge(path));
  return values.find((n) => n > 0) ?? 0;
}
