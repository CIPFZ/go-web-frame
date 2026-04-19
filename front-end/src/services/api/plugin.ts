import { request } from '@umijs/max';

export type PluginItem = {
  ID: number;
  code: string;
  repositoryUrl: string;
  nameZh: string;
  nameEn: string;
  descriptionZh: string;
  descriptionEn: string;
  departmentId: number;
  department?: string;
  ownerId: number;
  createdBy: number;
  createdAt: string;
};

export type ProductItem = {
  ID: number;
  code: string;
  name: string;
  description?: string;
};

export type DepartmentItem = {
  ID: number;
  name: string;
  productLine: string;
};

export type CompatibleProductItem = {
  ID?: number;
  productId: number;
  productCode?: string;
  productName?: string;
  versionConstraint?: string;
};

export type PluginReleaseItem = {
  ID: number;
  pluginId: number;
  pluginCode: string;
  pluginNameZh: string;
  requestType: number;
  status: number;
  processStatus: number;
  version: string;
  claimerId?: number;
  reviewComment?: string;
  testReportUrl?: string;
  packageX86Url?: string;
  packageArmUrl?: string;
  changelogZh?: string;
  changelogEn?: string;
  offlineReasonZh?: string;
  offlineReasonEn?: string;
  tdId?: string;
  submittedAt?: string;
  approvedAt?: string;
  releasedAt?: string;
  offlinedAt?: string;
  claimedAt?: string;
  compatibleItems?: CompatibleProductItem[];
  createdBy: number;
  createdAt: string;
};

export type ProjectDetail = {
  plugin: PluginItem;
  selectedRelease?: PluginReleaseItem;
  releases: PluginReleaseItem[];
  events: Array<{
    ID: number;
    action: string;
    comment?: string;
    createdAt: string;
  }>;
};

export type PublishedPluginItem = {
  ID: number;
  code: string;
  nameZh: string;
  nameEn: string;
  descriptionZh: string;
  descriptionEn: string;
  latestVersion: string;
  compatibleItems?: CompatibleProductItem[];
};

export type PublishedPluginDetail = {
  plugin: PluginItem;
  release: {
    ID: number;
    version: string;
    changelogZh?: string;
    changelogEn?: string;
    testReportUrl?: string;
    packageX86Url?: string;
    packageArmUrl?: string;
    releasedAt?: string;
    compatibleItems?: CompatibleProductItem[];
  };
  versions: Array<{
    ID: number;
    version: string;
    changelogZh?: string;
    changelogEn?: string;
    testReportUrl?: string;
    packageX86Url?: string;
    packageArmUrl?: string;
    releasedAt?: string;
    compatibleItems?: CompatibleProductItem[];
  }>;
};

const jsonRequest = (url: string, data: any, method = 'POST', options?: Record<string, any>) =>
  request<API.CommonResponse>(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    data,
    ...(options || {}),
  });

export const getPluginList = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/plugin/getPluginList', data, 'POST', options);
export const getProjectDetail = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/plugin/getProjectDetail', data, 'POST', options);
export const createPlugin = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/plugin/createPlugin', data, 'POST', options);
export const updatePlugin = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/plugin/updatePlugin', data, 'PUT', options);
export const getReleaseDetail = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/release/getReleaseDetail', data, 'POST', options);
export const createRelease = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/release/createRelease', data, 'POST', options);
export const updateRelease = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/release/updateRelease', data, 'PUT', options);
export const transitionRelease = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/release/transition', data, 'POST', options);
export const claimWorkOrder = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/release/claim', data, 'POST', options);
export const resetWorkOrder = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/release/reset', data, 'POST', options);
export const getWorkOrderPool = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/work-order/getWorkOrderPool', data, 'POST', options);
export const getProductList = (data: any = { page: 1, pageSize: 999 }, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/product/getProductList', data, 'POST', options);
export const getDepartmentList = (data: any = { page: 1, pageSize: 999 }, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/department/getDepartmentList', data, 'POST', options);
export const getPublishedPluginList = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/public/getPublishedPluginList', data, 'POST', options);
export const getPublishedPluginDetail = (data: any, options?: Record<string, any>) =>
  jsonRequest('/api/v1/plugin/public/getPublishedPluginDetail', data, 'POST', options);
