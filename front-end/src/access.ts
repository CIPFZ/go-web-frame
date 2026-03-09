/**
 * @see https://umijs.org/docs/max/access#access
 * 权限控制
 * 根据当前用户的角色获取不同的权限
 * */
export default function access(
  initialState: { currentUser?: API.CurrentUser } | undefined,
) {
  const { currentUser } = initialState ?? {};
  // return {
  //   // 是否为 admin
  //   canAdmin: currentUser && currentUser.access === 'admin',
  // };
  return {};
}
