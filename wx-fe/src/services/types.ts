export interface ApiResponse<T = any> {
  code?: number;
  message?: string;
  data: T;
}

export interface PaginationResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export type PaginatedResponse<T> = ApiResponse<PaginationResponse<T>>;

export interface Product {
  id: number;
  name: string;
  category_id: number;
  price: string | number;
  original_price?: string | number;
  images?: string;
  stock?: number;
  status?: number;
}

export interface Category {
  id: number;
  name: string;
  parent_id?: number;
  sort?: number;
  status?: number;
}

export interface CartItem {
  id: number;
  product_id: number;
  sku_id?: number;
  quantity: number;
}

export interface Order {
  id: number;
  order_no: string;
  pay_amount: number | string;
  total_amount?: number | string;
  discount_amount?: number | string;
  delivery_fee?: number | string;
  status?: number | string;
  pay_status?: number | string;
  store_id?: number;
  delivery_type?: number;
  address_info?: string | Record<string, any>;
  remark?: string;
  order_type?: number;
  created_at?: string;
  paid_at?: string;
  completed_at?: string;
  cancelled_at?: string;
  cancel_reason?: string;
  items?: OrderItem[];
}

export interface OrderItem {
  id: number;
  order_id?: number;
  product_id: number;
  sku_id?: number;
  product_name?: string;
  sku_name?: string;
  quantity: number;
  price?: number | string;
  amount?: number | string;
  image?: string;
}

export interface OrderDetailPayload {
  order: Order;
  items: OrderItem[];
}

export interface ProductSku {
  id: number;
  sku_code?: string;
  name?: string;
  price?: number | string;
  stock?: number;
}

export interface User {
  id: number;
  uid?: string;
  open_id?: string;
  nickname?: string;
  avatar?: string;
  phone?: string;
  gender?: number;
  balance?: number;
  points?: number;
  role?: string;
}

export interface AuthResponse {
  token?: string;
  user?: User;
}

export interface Coupon {
  id: number;
  title?: string;
  discount?: number;
  expire_at?: string;
}

// 小程序侧用户持有优惠券（与后端 UserCoupon/Coupon 模型对齐的精简版）
export interface UserCoupon {
  id: number; // user_coupon id
  status: number; // 1未使用 2已使用 3已过期
  used_at?: string | null;
  order_id?: number | null;
  coupon: {
    id: number;
    store_id?: number | null; // 为空表示平台券，非空表示门店券
    name: string;
    type: number; // 1满减 2折扣 3免单
    amount: string; // JSON 序列化后的金额，前端按需展示
    discount: string;
    min_amount: string;
    total_count: number;
    used_count: number;
    status: number;
    start_time: string;
    end_time: string;
    description?: string;
  };
}

export interface AvailableCouponInfo {
  user_coupon_id: number;
  reason: string;
}

export interface AvailableCouponsResponse {
  available: UserCoupon[];
  unavailable: AvailableCouponInfo[];
}

export interface Store {
  id: number;
  name: string;
  address?: string;
  latitude?: number;
  longitude?: number;
}

export interface Activity {
  id: number;
  store_id?: number;
  name: string;
  type: number;
  start_time: string;
  end_time: string;
  rules?: string;
  status?: number;
  priority?: number;
  description?: string;
}
