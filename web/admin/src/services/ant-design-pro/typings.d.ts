// @ts-ignore
/* eslint-disable */

declare namespace API {
  type CurrentUser = {
    name?: string;
    avatar?: string;
    userid?: string;
    email?: string;
    signature?: string;
    title?: string;
    group?: string;
    tags?: { key?: string; label?: string }[];
    notifyCount?: number;
    unreadCount?: number;
    country?: string;
    access?: string;
    permissions?: string[];
    geographic?: {
      province?: { label?: string; key?: string };
      city?: { label?: string; key?: string };
    };
    address?: string;
    phone?: string;
  };

  type LoginResult = {
    status?: string;
    type?: string;
    currentAuthority?: string;
    token?: string;
    id?: number;
    mobile?: string;
    username?: string;
    expiredAt?: number;
  };

  type UserDetail = {
    id?: number;
    mobile?: string;
    nickName?: string;
    birthday?: number;
    gender?: string;
    role?: number;
    permissions?: string[];
  };

  type OrderGoods = {
    id?: number;
    skuId?: number;
    skuName?: string;
    skuPrice?: number;
    num?: number;
    totalPrice?: number;
  };

  type OrderInfo = {
    id?: number;
    userId?: number;
    orderSn?: string;
    status?: string;
    post?: string;
    total?: number;
    address?: string;
    name?: string;
    mobile?: string;
    addTime?: string;
    goods?: OrderGoods[];
  };

  type OrderList = {
    total?: number;
    list?: OrderInfo[];
  };

  type GoodsItem = {
    id?: number;
    name?: string;
    goodsSn?: string;
    marketPrice?: number;
    onSale?: boolean;
    isNew?: boolean;
    isHot?: boolean;
    categoryId?: number;
    brandId?: number;
    categoryName?: string;
    brandName?: string;
    skuCount?: number;
    soldNum?: number;
  };

  type GoodsList = {
    total?: number;
    list?: GoodsItem[];
  };

  type CategoryItem = {
    id?: number;
    name?: string;
    parentCategory?: number;
    level?: number;
  };

  type CategoryList = {
    list?: CategoryItem[];
  };

  type BrandItem = {
    id?: number;
    name?: string;
    logo?: string;
  };

  type BrandList = {
    list?: BrandItem[];
  };

  type GoodsSkuDetail = {
    id?: number;
    goodsId?: number;
    skuName?: string;
    skuCode?: string;
    barCode?: string;
    price?: number;
    promotionPrice?: number;
    inventory?: number;
    onSale?: boolean;
    pic?: string;
  };

  type GoodsImageDetail = {
    id?: number;
    url?: string;
    position?: number;
    isMaster?: boolean;
  };

  type GoodsDetail = {
    id?: number;
    categoryId?: number;
    brandId?: number;
    typeId?: number;
    name?: string;
    goodsSn?: string;
    marketPrice?: number;
    goodsBrief?: string;
    goodsFrontImage?: string;
    goodsImages?: string[];
    onSale?: boolean;
    isNew?: boolean;
    isHot?: boolean;
    soldNum?: number;
    brandName?: string;
    categoryName?: string;
    skus?: GoodsSkuDetail[];
    images?: GoodsImageDetail[];
  };

  type UserInfo = {
    id?: number;
    mobile?: string;
    nickName?: string;
    birthday?: number;
    gender?: string;
    role?: number;
  };

  type UserList = {
    total?: number;
    list?: UserInfo[];
  };

  type AddressInfo = {
    id?: number;
    name?: string;
    mobile?: string;
    Province?: string;
    City?: string;
    Districts?: string;
    address?: string;
    post_code?: string;
    is_default?: number;
  };

  type UserAddressList = {
    list?: AddressInfo[];
  };

  type StatusCount = {
    status?: number;
    count?: number;
  };

  type DailySales = {
    date?: string;
    orderCount?: number;
    amount?: number;
  };

  type TopGoods = {
    skuId?: number;
    skuName?: string;
    num?: number;
    amount?: number;
  };

  type DashboardStats = {
    totalOrders?: number;
    totalSales?: number;
    todayOrders?: number;
    todaySales?: number;
    totalUsers?: number;
    statusCounts?: StatusCount[];
    last30Days?: DailySales[];
    topGoods?: TopGoods[];
  };

  type PermissionItem = {
    id?: number;
    code?: string;
    name?: string;
    groupName?: string;
    sort?: number;
  };

  type PermissionList = {
    list?: PermissionItem[];
  };

  type RolePermissionList = {
    codes?: string[];
  };

  type PageParams = {
    current?: number;
    pageSize?: number;
  };

  type RuleListItem = {
    key?: number;
    disabled?: boolean;
    href?: string;
    avatar?: string;
    name?: string;
    owner?: string;
    desc?: string;
    callNo?: number;
    status?: number;
    updatedAt?: string;
    createdAt?: string;
    progress?: number;
  };

  type RuleList = {
    data?: RuleListItem[];
    /** 列表的内容总数 */
    total?: number;
    success?: boolean;
  };

  type FakeCaptcha = {
    code?: number;
    status?: string;
  };

  type LoginParams = {
    username?: string;
    password?: string;
    autoLogin?: boolean;
    type?: string;
  };

  type ErrorResponse = {
    /** 业务约定的错误码 */
    errorCode: string;
    /** 业务上的错误信息 */
    errorMessage?: string;
    /** 业务上的请求是否成功 */
    success?: boolean;
  };

  type NoticeIconList = {
    data?: NoticeIconItem[];
    /** 列表的内容总数 */
    total?: number;
    success?: boolean;
  };

  type NoticeIconItemType = 'notification' | 'message' | 'event';

  type NoticeIconItem = {
    id?: string;
    extra?: string;
    key?: string;
    read?: boolean;
    avatar?: string;
    title?: string;
    status?: string;
    datetime?: string;
    description?: string;
    type?: NoticeIconItemType;
  };
}
