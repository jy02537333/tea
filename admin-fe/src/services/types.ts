export interface PaginationResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

// Standard API response wrapper used by tea-api
export interface ApiResponse<T = any> {
  code?: number;
  message?: string;
  data: T;
}

// Paginated wrapper where `data` is the array payload
export type PaginatedResponse<T> = ApiResponse<{
  data: T[];
  total: number;
  page: number;
  limit: number;
}>;

export interface User {
  id: number;
  nickname?: string;
  avatar?: string;
  phone?: string;
}

export interface Category {
  id: number;
  name: string;
  description?: string;
  parent_id?: number;
}

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

export interface OrderSummary {
  id: number;
  order_no: string;
  pay_amount: number;
  status?: number | string;
  address_info?: string;
}

export interface Store {
  id: number;
  name: string;
  address?: string;
  latitude?: number;
  longitude?: number;
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

export interface Order {
  id: number;
  order_no: string;
  pay_amount: number;
  status?: number | string;
  address_info?: string;
  items?: OrderItem[];
}

export interface Role {
  id: number;
  name: string;
  display_name?: string;
}

export interface Permission {
  id: number;
  name: string;
  display_name?: string;
  description?: string;
}

export interface Coupon {
  id: number;
  title?: string;
  discount?: number;
  expire_at?: string;
}

export interface StoreOrderStats {
  order_count?: number;
  total_amount?: number;
}

export interface PaymentIntent {
  id?: string;
  client_secret?: string;
  amount?: number;
}

export interface PaymentCallbackResult {
  success: boolean;
  message?: string;
}

export interface AccrualSummary {
  record_count?: number;
  user_count?: number;
  total_interest?: string | number;
}
