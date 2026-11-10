/**
 * 子页面保活的菜单分组（分组的 path，如 '/reports'）：同一分组下切换页面时，去过的页面不卸载，
 * 切回来直接是上次的数据和图表，不重新请求、不重画。适合页面间频繁切换、图表较多的分组。
 * 保活的页面可以用 useRouteActive()（components/KeepAliveOutlet）判断自己是否正在显示，只在显示时轮询。
 * 由项目按需登记；默认不保活。
 */
export const KEEP_ALIVE_GROUPS: ReadonlySet<string> = new Set<string>([]);
