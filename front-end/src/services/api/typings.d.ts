declare namespace API {
  // API 项类型 (基于你的响应)
  type ApiItem = {
    ID: number;
    path: string;
    description: string;
    apiGroup: string;
    method: string;
  };

  interface OsInfo {
    goos: string;
    numCpu: number;
    compiler: string;
    goVersion: string;
    numGoroutine: number;
  }

  interface CpuInfo {
    cpus: number[];
    cores: number;
    load1?: number;
    load5?: number;
    load15?: number;
  }

  interface RamInfo {
    used: number;
    total: number;
    usedPercent?: number;
  }

  interface DiskInfo {
    mountPoint: string;
    used: number;
    total: number;
    usedPercent?: number;
    readBytes?: number;
    writeBytes?: number;
  }

  interface ServerInfo {
    os: OsInfo;
    cpu: CpuInfo;
    ram: RamInfo;
    disk: DiskInfo[];
    io?: {
      readBytes: number;
      writeBytes: number;
    };
  }

  // 公共响应结构
  type CommonResponse = {
    code: number;
    msg: string;
    data: any;
  }

  // 用户权限信息
  type Authority = {
    authorityId: number;
    authorityName: string;
    parentId: number;
    dataAuthorityId: any;
    children: any;
    menus: any;
    defaultRouter: string;
  };

  // -------- 用户登录 ------------------
  type LoginParams = {
    username?: string;
    password?: string;
    autoLogin?: boolean;
    type?: string;
  };
  // 用户信息
  type UserInfo = {
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
  };

  type CurrentUser = UserInfo;

  // 记录前端登录的状态
  type LoginResult = {
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
