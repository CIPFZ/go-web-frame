import type { ReleaseRequestType, ReleaseStatus } from '@/services/api/plugin';

type DetailEntrySource = 'project' | 'review' | 'publish';
type WorkbenchAssignmentScope = 'assigned' | 'all';

type PublishWorkbenchQueryInput = {
  page: number;
  pageSize: number;
  keyword: string;
  requestType: ReleaseRequestType | 'all';
  status: ReleaseStatus | 'all';
  scope: WorkbenchAssignmentScope;
  currentUserId?: number;
};

type ReviewWorkbenchQueryInput = {
  keyword: string;
  requestType: ReleaseRequestType | 'all';
  scope: WorkbenchAssignmentScope;
  currentUserId?: number;
};

type ReleaseTransitionPayloadInput = {
  id: number;
  action: 'submit_review' | 'approve' | 'reject' | 'release' | 'revise';
  reviewComment?: string;
  reviewerId?: number;
  publisherId?: number;
};

type ReleaseAssignmentPayloadInput = {
  id: number;
  reviewerId?: number;
  publisherId?: number;
  comment?: string;
};

export const resolveInitialTab = (from?: string | null, tab?: string | null) => {
  if (tab) return tab;
  if (from === 'review' || from === 'publish') return 'review';
  return 'overview';
};

export const resolveDetailEntryMeta = (rawSource?: string | null) => {
  const source: DetailEntrySource =
    rawSource === 'review' || rawSource === 'publish' ? rawSource : 'project';

  if (source === 'review') {
    return {
      source,
      backTarget: '/plugin/review-workbench',
      backLabel: '返回审核工作台',
      sectionLabel: '审核工作台',
      pageContent:
        '当前从审核工作台进入，页面已自动定位到待审核版本。你可以在完整项目上下文里查看资料、核对时间轴并完成审核。',
    };
  }

  if (source === 'publish') {
    return {
      source,
      backTarget: '/plugin/publish-workbench',
      backLabel: '返回发布工作台',
      sectionLabel: '发布工作台',
      pageContent:
        '当前从发布工作台进入，页面已自动定位到待执行版本。你可以在完整项目上下文里核对资料、查看时间轴并执行发布或下架。',
    };
  }

  return {
    source,
    backTarget: '/plugin/center',
    backLabel: '返回项目管理',
    sectionLabel: '项目管理',
    pageContent:
      '项目详情页承载项目基础信息、版本列表和流程信息。先选择版本，再查看对应资料、审核记录与时间轴。',
  };
};

export const buildPublishWorkbenchQuery = ({
  page,
  pageSize,
  keyword,
  requestType,
  status,
  scope,
  currentUserId,
}: PublishWorkbenchQueryInput) => {
  const trimmedKeyword = keyword.trim();

  return {
    page,
    pageSize,
    keyword: trimmedKeyword || undefined,
    requestType: requestType === 'all' ? undefined : requestType,
    status: status === 'all' ? undefined : status,
    publisherId: scope === 'assigned' ? currentUserId : undefined,
  };
};

export const buildReviewWorkbenchBaseQuery = ({
  keyword,
  requestType,
  scope,
  currentUserId,
}: ReviewWorkbenchQueryInput) => {
  const trimmedKeyword = keyword.trim();

  return {
    keyword: trimmedKeyword || undefined,
    requestType: requestType === 'all' ? undefined : requestType,
    reviewerId: scope === 'assigned' ? currentUserId : undefined,
  };
};

export const buildReleaseTransitionPayload = ({
  id,
  action,
  reviewComment,
  reviewerId,
  publisherId,
}: ReleaseTransitionPayloadInput) => {
  const trimmedComment = reviewComment?.trim();

  return {
    id,
    action,
    reviewComment: trimmedComment || undefined,
    reviewerId:
      action === 'submit_review' && typeof reviewerId === 'number' && reviewerId > 0
        ? reviewerId
        : undefined,
    publisherId:
      action === 'approve' && typeof publisherId === 'number' && publisherId > 0
        ? publisherId
        : undefined,
  };
};

export const buildReleaseAssignmentPayload = ({
  id,
  reviewerId,
  publisherId,
  comment,
}: ReleaseAssignmentPayloadInput) => {
  const trimmedComment = comment?.trim();

  return {
    id,
    reviewerId: typeof reviewerId === 'number' && reviewerId > 0 ? reviewerId : undefined,
    publisherId: typeof publisherId === 'number' && publisherId > 0 ? publisherId : undefined,
    comment: trimmedComment || undefined,
  };
};
