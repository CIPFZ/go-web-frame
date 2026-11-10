export namespace API {
  // API 项类型 (基于你的响应)
  export type ApiItem = {
    ID: number;
    path: string;
    description: string;
    apiGroup: string;
    method: string;
  };

  // 公共响应结构
  export type CommonResponse = {
    code: number;
    msg: string;
    data: any;
  }

  // 用户权限信息
  export type Authority = {
    authorityId: number;
    authorityName: string;
    parentId: number;
    dataAuthorityId: any;
    children: any;
    menus: any;
    defaultRouter: string;
    /** 角色的默认工作台（见 pages/dashboard/workplace/layout.ts）；为空时用预置的 */
    workplace?: unknown;
  };

  // -------- 用户登录 ------------------
  export type LoginParams = {
    username?: string;
    password?: string;
    autoLogin?: boolean;
    type?: string;
  };
  // 用户信息
  export type UserInfo = {
    ID?: number;
    uuid?: string;
    username?: string;
    nickName?: string;
    avatar?: string;
    authorityId?: number;
    authority?: Authority;
    authorities?: Authority[];
    phone?: string;
    email?: string;
    status: number;
    settings?: any;
    /** 本人的工作台布局（见 pages/dashboard/workplace/layout.ts） */
    workplace?: unknown;
  };

  export type CurrentUser = UserInfo;

  // 记录前端登录的状态
  export type LoginResult = {
    // -1 还没有操作 0：失败  1：成功
    code: number;
    msg: string;
  };

  // 分页查询信息
  export type PageResult = {
    list: UserInfo[];
    total: number;
    page: number;
    pageSize: number;
  };

  // 定义请求参数类型
  export type UpdateSelfInfoParams = {
    nickName: string;
    bio?: string;
  }

  export type UpdateUiConfigParams = {
    settings: Record<string, any>;
  }

}
