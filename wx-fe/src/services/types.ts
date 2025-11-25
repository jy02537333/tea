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

export interface CartItem {
  id: number;
  product_id: number;
  sku_id?: number;
  quantity: number;
}

export interface Order {
  id: number;
  order_no: string;
  pay_amount: number;
  status?: number | string;
  address_info?: string;
  items?: CartItem[];
}

export interface OrderItem {
  id: number;
  product_id: number;
  sku_id?: number;
  quantity: number;
  price?: number;
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
  nickname?: string;
  avatar?: string;
  phone?: string;
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

export interface Store {
  id: number;
  name: string;
  address?: string;
  latitude?: number;
  longitude?: number;
}
