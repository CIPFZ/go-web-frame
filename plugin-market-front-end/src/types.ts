export interface CompatibilityItem {
  targetType: string;
  productCode: string;
  productName: string;
  versionConstraint: string;
}

export interface PublishedPluginItem {
  ID: number;
  pluginId: number;
  code: string;
  nameZh: string;
  nameEn: string;
  descriptionZh: string;
  descriptionEn: string;
  latestVersion: string;
  releasedAt?: string;
  packageX86Url?: string;
  packageArmUrl?: string;
  compatibleItems: CompatibilityItem[];
}

export interface PublishedVersion {
  releaseId: number;
  version: string;
  publisher?: string;
  versionConstraint?: string;
  changelogZh?: string;
  changelogEn?: string;
  performanceSummaryZh?: string;
  performanceSummaryEn?: string;
  testReportUrl?: string;
  packageX86Url?: string;
  packageArmUrl?: string;
  releasedAt?: string;
  compatibleItems: CompatibilityItem[];
}

export interface PublishedPluginDetail {
  plugin: {
    ID: number;
    pluginId: number;
    code: string;
    nameZh: string;
    nameEn: string;
    descriptionZh: string;
    descriptionEn: string;
    capabilityZh?: string;
    capabilityEn?: string;
    ownerName?: string;
  };
  release: PublishedVersion;
  versions: PublishedVersion[];
}

export interface ApiResponse<T> {
  code: number;
  msg: string;
  data: T;
}

export interface PageResult<T> {
  list: T[];
  total: number;
  page: number;
  pageSize: number;
}
