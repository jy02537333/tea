export interface User {
  id: number;
  username?: string;
  nickname?: string;
  role?: string;
  // 统一来源：GET /api/v1/users/me/summary（user/info 契约对齐）
  uid?: string;
  open_id?: string;
  phone?: string;
  gender?: number;
  balance?: number;
  points?: number;
  membership_package_id?: number | null;
  partner_level_id?: number | null;
  membership_level_name?: string;
  discount_rate?: number;
  purchase_discount_rate?: number;
  direct_commission_rate?: number;
  team_commission_rate?: number;
  upgrade_reward_rate?: number;
}
