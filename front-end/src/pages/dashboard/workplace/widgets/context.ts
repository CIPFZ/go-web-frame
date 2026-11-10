import type { MyNotice } from '@/services/system/notice';
import { createContext, useContext } from 'react';
import type { useNotices } from '../useNotices';

/** 工作台页面给小组件的共享内容：通知（与页头的个人通知数同一份）、服务状态、当前账号。 */
export interface DeskContextValue {
  notices: ReturnType<typeof useNotices>;
  openNotice: (n: MyNotice) => void;
  server: {
    checks?: Record<string, string>;
    checkedAt?: Date;
    error: boolean;
    loading: boolean;
    refresh: () => void;
  };
  user?: {
    nickName?: string;
    username?: string;
    avatar?: string;
    email?: string;
  };
  role: string;
  accountPath?: string;
  statePath?: string;
  /** 小组件打开的页面的实际地址（见 useDeskPages） */
  pathOf: (widget: string) => string | undefined;
  /** 本人能打开的页面数 */
  pagesCount: number;
  /** 进入编辑并打开组件库（小组件或应用图标） */
  openGallery: (tab: 'widgets' | 'apps') => void;
}

export const DeskContext = createContext<DeskContextValue | null>(null);

export function useDesk(): DeskContextValue {
  const v = useContext(DeskContext);
  if (!v) throw new Error('useDesk outside the workplace');
  return v;
}
