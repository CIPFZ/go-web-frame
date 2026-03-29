const adminRoleIds = new Set([1, 9528]);
const requesterRoleIds = new Set([10010]);
const reviewerRoleIds = new Set([10013]);
const publisherRoleIds = new Set([10014]);

const collectAuthorityIds = (currentUser?: API.UserInfo) => {
  const ids = new Set<number>();
  if (currentUser?.authorityId) ids.add(currentUser.authorityId);
  (currentUser?.authorities || []).forEach((item: any) => {
    if (item?.authorityId) ids.add(item.authorityId);
  });
  return ids;
};

export default function access(
  initialState: { currentUser?: API.UserInfo } | undefined,
) {
  const { currentUser } = initialState ?? {};
  const authorityIds = collectAuthorityIds(currentUser);
  const hasAny = (targets: Set<number>) => Array.from(authorityIds).some((id) => targets.has(id));
  const isAdmin = hasAny(adminRoleIds);
  const canManagePluginProjects = isAdmin || hasAny(requesterRoleIds);
  const canReviewPlugins = isAdmin || hasAny(reviewerRoleIds);
  const canPublishPlugins = isAdmin || hasAny(publisherRoleIds);

  return {
    canManagePluginProjects,
    canReviewPlugins,
    canPublishPlugins,
    canAccessPluginProjectDetail: canManagePluginProjects || canReviewPlugins || canPublishPlugins,
  };
}
