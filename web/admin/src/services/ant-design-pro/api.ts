// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** 获取当前的用户 GET /api/currentUser */
export async function currentUser(options?: { [key: string]: any }) {
  const res = await request<API.UserDetail>('/api/users/detail', {
    method: 'GET',
    ...(options || {}),
  });
  return {
    data: {
      name: res.nickName || res.mobile,
      userid: String(res.id),
      access: Number(res.role) === 2 ? 'admin' : 'user',
      permissions: res.permissions || [],
    } as API.CurrentUser,
  };
}

/** 退出登录接口 POST /api/login/outLogin */
export async function outLogin(options?: { [key: string]: any }) {
  localStorage.removeItem('kratos_admin_token');
  return {};
}

/** 登录接口 POST /api/users/login */
export async function login(body: API.LoginParams, options?: { [key: string]: any }) {
  const res = await request<API.LoginResult>('/api/users/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
  if (res.token) {
    localStorage.setItem('kratos_admin_token', res.token);
  }
  return {
    status: 'ok',
    ...res,
  };
}

/** 此处后端没有提供注释 GET /api/notices */
export async function getNotices(options?: { [key: string]: any }) {
  return request<API.NoticeIconList>('/api/notices', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 获取规则列表 GET /api/rule */
export async function rule(
  params: {
    // query
    /** 当前的页码 */
    current?: number;
    /** 页面的容量 */
    pageSize?: number;
  },
  options?: { [key: string]: any },
) {
  return request<API.RuleList>('/api/rule', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 更新规则 PUT /api/rule */
export async function updateRule(options?: { [key: string]: any }) {
  return request<API.RuleListItem>('/api/rule', {
    method: 'POST',
    data: {
      method: 'update',
      ...(options || {}),
    },
  });
}

/** 新建规则 POST /api/rule */
export async function addRule(options?: { [key: string]: any }) {
  return request<API.RuleListItem>('/api/rule', {
    method: 'POST',
    data: {
      method: 'post',
      ...(options || {}),
    },
  });
}

/** 删除规则 DELETE /api/rule */
export async function removeRule(options?: { [key: string]: any }) {
  return request<Record<string, any>>('/api/rule', {
    method: 'POST',
    data: {
      method: 'delete',
      ...(options || {}),
    },
  });
}

/** 订单管理 */
export async function adminOrderList(
  params: { page?: number; pageSize?: number; status?: number },
  options?: { [key: string]: any },
) {
  return request<API.OrderList>('/api/order/list', {
    method: 'GET',
    params,
    ...(options || {}),
  });
}

export async function adminOrderDetail(orderSn: string, options?: { [key: string]: any }) {
  return request<API.OrderInfo>('/api/order/detail', {
    method: 'GET',
    params: { orderSn },
    ...(options || {}),
  });
}

export async function adminShipOrder(orderSn: string, post: string, options?: { [key: string]: any }) {
  return request<{ success?: boolean }>('/api/order/ship', {
    method: 'POST',
    data: { orderSn, post },
    ...(options || {}),
  });
}

export async function adminRefundOrder(orderSn: string, options?: { [key: string]: any }) {
  return request<{ success?: boolean }>('/api/order/refund', {
    method: 'POST',
    data: { orderSn },
    ...(options || {}),
  });
}

/** 商品管理 */
export async function adminGoodsList(
  params: {
    page?: number;
    pageSize?: number;
    keywords?: string;
    categoryId?: number;
    brandId?: number;
    skuCode?: string;
  },
  options?: { [key: string]: any },
) {
  return request<API.GoodsList>('/api/goods/list', {
    method: 'GET',
    params,
    ...(options || {}),
  });
}

export async function adminGoodsCategories(options?: { [key: string]: any }) {
  return request<API.CategoryList>('/api/goods/categories', {
    method: 'GET',
    ...(options || {}),
  });
}

export async function adminGoodsBrands(options?: { [key: string]: any }) {
  return request<API.BrandList>('/api/goods/brands', {
    method: 'GET',
    ...(options || {}),
  });
}

export async function adminCreateCategory(
  body: Record<string, any>,
  options?: { [key: string]: any },
) {
  return request<API.CategoryItem>('/api/goods/category/create', {
    method: 'POST',
    data: body,
    ...(options || {}),
  });
}

export async function adminUpdateCategory(
  body: Record<string, any>,
  options?: { [key: string]: any },
) {
  return request<{ success?: boolean }>('/api/goods/category/update', {
    method: 'PUT',
    data: body,
    ...(options || {}),
  });
}

export async function adminDeleteCategory(id: number, options?: { [key: string]: any }) {
  return request<{ success?: boolean }>('/api/goods/category/delete', {
    method: 'DELETE',
    params: { id },
    ...(options || {}),
  });
}

export async function adminGoodsDetail(id: number, options?: { [key: string]: any }) {
  return request<API.GoodsDetail>('/api/goods/detail', {
    method: 'GET',
    params: { id },
    ...(options || {}),
  });
}

export async function adminUpdateGoodsStatus(
  id: number,
  onSale: boolean,
  options?: { [key: string]: any },
) {
  return request<{ success?: boolean }>('/api/goods/status', {
    method: 'PUT',
    data: { id, onSale },
    ...(options || {}),
  });
}

export async function adminDeleteGoods(id: number, options?: { [key: string]: any }) {
  return request<{ success?: boolean }>('/api/goods/delete', {
    method: 'DELETE',
    params: { id },
    ...(options || {}),
  });
}

export async function adminCreateGoods(
  body: Record<string, any>,
  options?: { [key: string]: any },
) {
  return request<{ success?: boolean }>('/api/goods/create', {
    method: 'POST',
    data: body,
    ...(options || {}),
  });
}

/** 用户管理 */
export async function adminUserList(
  params: { page?: number; pageSize?: number },
  options?: { [key: string]: any },
) {
  return request<API.UserList>('/api/user/list', {
    method: 'GET',
    params,
    ...(options || {}),
  });
}

export async function adminUserAddressList(
  uid: number,
  options?: { [key: string]: any },
) {
  return request<API.UserAddressList>('/api/user/address/list', {
    method: 'GET',
    params: { uid },
    ...(options || {}),
  });
}

export async function adminUserAddressDelete(
  id: number,
  uid: number,
  options?: { [key: string]: any },
) {
  return request<{ success?: boolean }>('/api/user/address/delete', {
    method: 'DELETE',
    params: { id, uid },
    ...(options || {}),
  });
}

/** 运营看板 */
export async function adminDashboardStats(options?: { [key: string]: any }) {
  return request<API.DashboardStats>('/api/dashboard/stats', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 系统权限 */
export async function adminPermissionList(options?: { [key: string]: any }) {
  return request<API.PermissionList>('/api/system/permissions', {
    method: 'GET',
    ...(options || {}),
  });
}

export async function adminRolePermissionList(
  role: number,
  options?: { [key: string]: any },
) {
  return request<API.RolePermissionList>('/api/system/role-permissions', {
    method: 'GET',
    params: { role },
    ...(options || {}),
  });
}

export async function adminRolePermissionUpdate(
  role: number,
  codes: string[],
  options?: { [key: string]: any },
) {
  return request<{ success?: boolean }>('/api/system/role-permissions', {
    method: 'PUT',
    data: { role, codes },
    ...(options || {}),
  });
}
