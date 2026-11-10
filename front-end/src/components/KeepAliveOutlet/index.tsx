import { useLocation, useOutlet } from '@umijs/max';
import React, { Suspense, createContext, useContext, useLayoutEffect, useRef, useState } from 'react';

// 菜单分组的“保活”子页面：去过的子页面切走时不卸载，只是隐藏，切回来直接是上次的数据和图表，不重新请求、不重画。
// 隐藏的页面移到屏幕外、保持离开时的宽度：尺寸不变，自适应宽度的图表就不会重画。
// 每个页面有自己的 Suspense：打开一个还没加载代码的页面时，不会把已经打开的页面一起藏起来（那样图表也会重画）。

const ActiveContext = createContext(true);

/** 当前子页面是否显示中（不在保活分组里时总是 true）：轮询等只在显示时进行。 */
export const useRouteActive = () => useContext(ActiveContext);

function Page({ active, children }: { active: boolean; children: React.ReactNode }) {
  const ref = useRef<HTMLDivElement>(null);
  const [width, setWidth] = useState<number>();
  // 显示时记下宽度，隐藏时按这个宽度放到屏幕外
  useLayoutEffect(() => {
    const el = ref.current;
    if (!active || !el) return;
    setWidth(el.getBoundingClientRect().width);
    const ro = new ResizeObserver(([e]) => setWidth(e.contentRect.width));
    ro.observe(el);
    return () => ro.disconnect();
  }, [active]);
  const style: React.CSSProperties | undefined = active
    ? undefined
    : { position: 'fixed', top: 0, left: -100000, width, visibility: 'hidden', pointerEvents: 'none' };
  return (
    <div ref={ref} style={style} aria-hidden={!active}>
      <ActiveContext.Provider value={active}>
        <Suspense fallback={null}>{children}</Suspense>
      </ActiveContext.Provider>
    </div>
  );
}

export default function KeepAliveOutlet() {
  const { pathname } = useLocation();
  const outlet = useOutlet();
  const pages = useRef(new Map<string, React.ReactNode>());
  if (outlet && !pages.current.has(pathname)) pages.current.set(pathname, outlet);
  return (
    <>
      {[...pages.current.entries()].map(([path, element]) => (
        <Page key={path} active={path === pathname}>
          {element}
        </Page>
      ))}
    </>
  );
}
