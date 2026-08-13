/**
 * @see https://umijs.org/docs/max/access#access
 * */
export default function access(
  initialState: { currentUser?: API.CurrentUser } | undefined,
) {
  const { currentUser } = initialState ?? {};
  return {
    canAdmin: currentUser && currentUser.access === 'admin',
    can: (code: string) => currentUser?.permissions?.includes(code) || false,
    canPermissionManage:
      currentUser?.permissions?.includes('system:permission:manage') || false,
    canCategoryManage:
      currentUser?.permissions?.includes('goods:category:manage') || false,
  };
}
