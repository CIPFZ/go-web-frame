// src/utils/menuDataStore.ts
import { getMenu } from '@/services/api/menu';

// 用于缓存正在进行的请求 Promise 或 结果
let menuPromiseCache: Promise<any[]> | null = null;

/**
 * 获取菜单数据
 * @param forceRefresh 是否强制刷新（忽略缓存）
 */
export const fetchMenuData = async (forceRefresh = false): Promise<any[]> => {
  // 1. 如果强制刷新，先清除缓存
  if (forceRefresh) {
    menuPromiseCache = null;
  }

  // 2. 如果缓存中已有 Promise（无论是正在请求中，还是已完成），直接返回
  // 这保证了并发调用（如 app 启动时）只发送一次网络请求
  if (menuPromiseCache) {
    return menuPromiseCache;
  }

  // 3. 发起新请求并缓存 Promise
  menuPromiseCache = getMenu()
    .then((res) => {
      if (res.code === 0) {
        return res.data || [];
      }
      return [];
    })
    .catch((err) => {
      console.error('Fetch menu error:', err);
      return [];
    });

  return menuPromiseCache;
};

/**
 * 清除菜单缓存
 * 在执行增删改操作后调用此方法
 */
export const clearMenuCache = () => {
  menuPromiseCache = null;
};
