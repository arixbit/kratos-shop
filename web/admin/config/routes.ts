/**
 * @name umi 的路由配置
 * @doc https://umijs.org/docs/guides/routes
 */
export default [
  {
    path: '/user',
    layout: false,
    routes: [
      {
        path: '/user/login',
        name: 'login',
        component: './user/login',
      },
      {
        path: '/user',
        redirect: '/user/login',
      },
    ],
  },
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    icon: 'dashboard',
    component: './dashboard/index',
  },
  {
    path: '/users',
    name: 'users',
    icon: 'team',
    component: './users/index',
  },
  {
    path: '/orders',
    name: 'orders',
    icon: 'shoppingCart',
    component: './orders/index',
  },
  {
    path: '/goods',
    name: 'goods',
    icon: 'shopping',
    component: './goods/index',
  },
  {
    path: '/goods/categories',
    name: 'categories',
    icon: 'appstore',
    access: 'canCategoryManage',
    component: './goods/categories',
  },
  {
    path: '/system/permissions',
    name: 'permissions',
    icon: 'safety',
    access: 'canPermissionManage',
    component: './system/permissions',
  },
  {
    path: '*',
    component: './exception/404',
  },
];
