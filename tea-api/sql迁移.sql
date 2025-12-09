-- 茶心阁 SQL 迁移脚本（手动执行）
-- 说明：此文件由代码生成草稿生成，包含项目主要表的 CREATE TABLE 语句。
-- 在生产环境执行前请先 review，并根据业务需求补充索引/外键/分区/权限等。

SET FOREIGN_KEY_CHECKS=0;

-- 基础用户/权限表
CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `username` VARCHAR(100) DEFAULT NULL,
  `password_hash` VARCHAR(255) DEFAULT NULL,
  `open_id` VARCHAR(50) DEFAULT NULL,
  `union_id` VARCHAR(50) DEFAULT NULL,
  `phone` VARCHAR(20) DEFAULT NULL,
  `nickname` VARCHAR(50) DEFAULT NULL,
  `avatar` VARCHAR(500) DEFAULT NULL,
  `gender` TINYINT DEFAULT 0,
  `birthday` DATETIME NULL,
  `province` VARCHAR(50) DEFAULT NULL,
  `city` VARCHAR(50) DEFAULT NULL,
  `country` VARCHAR(50) DEFAULT NULL,
  `default_address` JSON DEFAULT NULL,
  `default_address_updated_at` DATETIME NULL,
  `status` TINYINT DEFAULT 1,
  `last_login_at` DATETIME NULL,
  `balance` DECIMAL(12,2) DEFAULT 0,
  `interest_rate` DECIMAL(8,6) DEFAULT 0,
  `points` INT DEFAULT 0,
  `role` VARCHAR(30) DEFAULT 'user',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_uid` (`uid`),
  UNIQUE KEY `idx_users_username` (`username`),
  UNIQUE KEY `idx_users_openid` (`open_id`),
  UNIQUE KEY `idx_users_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `roles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `name` VARCHAR(50) NOT NULL,
  `display_name` VARCHAR(100) DEFAULT NULL,
  `description` TEXT,
  `status` TINYINT DEFAULT 1,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_roles_name` (`name`),
  UNIQUE KEY `idx_roles_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `permissions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `name` VARCHAR(100) NOT NULL,
  `display_name` VARCHAR(100) DEFAULT NULL,
  `description` TEXT,
  `module` VARCHAR(50) DEFAULT NULL,
  `action` VARCHAR(50) DEFAULT NULL,
  `resource` VARCHAR(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_permissions_name` (`name`),
  UNIQUE KEY `idx_permissions_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `user_roles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `user_id` BIGINT UNSIGNED NOT NULL,
  `role_id` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_roles_user_id` (`user_id`),
  KEY `idx_user_roles_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `role_permissions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `role_id` BIGINT UNSIGNED NOT NULL,
  `permission_id` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_role_permissions_role_id` (`role_id`),
  KEY `idx_role_permissions_permission_id` (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 商品相关
CREATE TABLE IF NOT EXISTS `categories` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `name` VARCHAR(50) NOT NULL,
  `description` TEXT,
  `image` VARCHAR(500) DEFAULT NULL,
  `sort` INT DEFAULT 0,
  `status` TINYINT DEFAULT 1,
  `parent_id` BIGINT UNSIGNED DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_categories_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `products` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `category_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `description` TEXT,
  `images` TEXT,
  `price` DECIMAL(10,2) NOT NULL,
  `original_price` DECIMAL(10,2) DEFAULT NULL,
  `stock` INT DEFAULT 0,
  `sales` INT DEFAULT 0,
  `status` TINYINT DEFAULT 1,
  `sort` INT DEFAULT 0,
  `is_hot` TINYINT(1) DEFAULT 0,
  `is_new` TINYINT(1) DEFAULT 0,
  `is_recommend` TINYINT(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_products_category_id` (`category_id`),
  UNIQUE KEY `idx_products_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `product_skus` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `product_id` BIGINT UNSIGNED NOT NULL,
  `sku_name` VARCHAR(100) DEFAULT NULL,
  `sku_code` VARCHAR(50) DEFAULT NULL,
  `price` DECIMAL(10,2) NOT NULL,
  `stock` INT DEFAULT 0,
  `sales` INT DEFAULT 0,
  `attrs` JSON DEFAULT NULL,
  `image` VARCHAR(500) DEFAULT NULL,
  `status` TINYINT DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `idx_product_skus_product_id` (`product_id`),
  UNIQUE KEY `idx_product_skus_sku_code` (`sku_code`),
  UNIQUE KEY `idx_product_skus_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `product_images` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `product_id` BIGINT UNSIGNED NOT NULL,
  `image_url` VARCHAR(500) NOT NULL,
  `sort` INT DEFAULT 0,
  `is_main` TINYINT(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_product_images_product_id` (`product_id`),
  UNIQUE KEY `idx_product_images_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 订单与购物车
CREATE TABLE IF NOT EXISTS `orders` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `order_no` VARCHAR(32) NOT NULL,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `store_id` BIGINT UNSIGNED DEFAULT 0,
  `total_amount` DECIMAL(10,2) NOT NULL,
  `pay_amount` DECIMAL(10,2) NOT NULL,
  `discount_amount` DECIMAL(10,2) DEFAULT 0,
  `delivery_fee` DECIMAL(10,2) DEFAULT 0,
  `status` TINYINT DEFAULT 1,
  `pay_status` TINYINT DEFAULT 1,
  `order_type` TINYINT DEFAULT 1,
  `delivery_type` TINYINT DEFAULT 1,
  `delivery_time` DATETIME NULL,
  `address_info` JSON DEFAULT NULL,
  `remark` TEXT,
  `paid_at` DATETIME NULL,
  `delivered_at` DATETIME NULL,
  `completed_at` DATETIME NULL,
  `cancelled_at` DATETIME NULL,
  `cancel_reason` VARCHAR(200) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_orders_order_no` (`order_no`),
  KEY `idx_orders_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `order_items` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `order_id` BIGINT UNSIGNED NOT NULL,
  `product_id` BIGINT UNSIGNED NOT NULL,
  `sku_id` BIGINT UNSIGNED DEFAULT NULL,
  `product_name` VARCHAR(100) NOT NULL,
  `sku_name` VARCHAR(100) DEFAULT NULL,
  `price` DECIMAL(10,2) NOT NULL,
  `quantity` INT NOT NULL,
  `amount` DECIMAL(10,2) NOT NULL,
  `image` VARCHAR(500) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order_items_order_id` (`order_id`),
  KEY `idx_order_items_product_id` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `carts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `user_id` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_carts_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `cart_items` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `cart_id` BIGINT UNSIGNED NOT NULL,
  `product_id` BIGINT UNSIGNED NOT NULL,
  `sku_id` BIGINT UNSIGNED DEFAULT NULL,
  `quantity` INT NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_cart_items_cart_id` (`cart_id`),
  KEY `idx_cart_items_product_id` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 支付/退款/提现
CREATE TABLE IF NOT EXISTS `payments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `order_id` BIGINT UNSIGNED NOT NULL,
  `payment_no` VARCHAR(64) NOT NULL,
  `payment_method` TINYINT NOT NULL,
  `amount` DECIMAL(10,2) NOT NULL,
  `status` TINYINT DEFAULT 1,
  `third_pay_no` VARCHAR(64) DEFAULT NULL,
  `third_response` TEXT,
  `paid_at` DATETIME NULL,
  `notify_at` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_payments_payment_no` (`payment_no`),
  KEY `idx_payments_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `refunds` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `order_id` BIGINT UNSIGNED NOT NULL,
  `payment_id` BIGINT UNSIGNED NOT NULL,
  `refund_no` VARCHAR(64) NOT NULL,
  `refund_amount` DECIMAL(10,2) NOT NULL,
  `refund_reason` VARCHAR(200) DEFAULT NULL,
  `status` TINYINT DEFAULT 1,
  `third_refund_no` VARCHAR(64) DEFAULT NULL,
  `third_response` TEXT,
  `refunded_at` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_refunds_refund_no` (`refund_no`),
  KEY `idx_refunds_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `withdraw_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `user_id` BIGINT UNSIGNED NOT NULL,
  `withdraw_no` VARCHAR(64) NOT NULL,
  `amount` DECIMAL(10,2) NOT NULL,
  `fee` DECIMAL(10,2) DEFAULT 0,
  `actual_amount` DECIMAL(10,2) NOT NULL,
  `withdraw_type` TINYINT NOT NULL,
  `status` TINYINT DEFAULT 1,
  `remark` VARCHAR(200) DEFAULT NULL,
  `processed_at` DATETIME NULL,
  `processed_by` BIGINT UNSIGNED DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_withdraws_withdraw_no` (`withdraw_no`),
  KEY `idx_withdraws_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `wechat_transfer_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `withdraw_id` BIGINT UNSIGNED NOT NULL,
  `partner_trade_no` VARCHAR(64) NOT NULL,
  `open_id` VARCHAR(50) NOT NULL,
  `amount` DECIMAL(10,2) NOT NULL,
  `description` VARCHAR(100) DEFAULT NULL,
  `status` TINYINT DEFAULT 1,
  `payment_no` VARCHAR(64) DEFAULT NULL,
  `payment_time` DATETIME NULL,
  `error_code` VARCHAR(20) DEFAULT NULL,
  `error_msg` VARCHAR(200) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_wtr_partner_trade_no` (`partner_trade_no`),
  KEY `idx_wtr_withdraw_id` (`withdraw_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 营销/优惠券
CREATE TABLE IF NOT EXISTS `coupons` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `code` VARCHAR(50) DEFAULT NULL,
  `title` VARCHAR(200) DEFAULT NULL,
  `type` TINYINT DEFAULT 1,
  `amount` DECIMAL(10,2) DEFAULT 0,
  `min_amount` DECIMAL(10,2) DEFAULT 0,
  `start_at` DATETIME NULL,
  `end_at` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_coupons_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `user_coupons` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `user_id` BIGINT UNSIGNED NOT NULL,
  `coupon_id` BIGINT UNSIGNED NOT NULL,
  `status` TINYINT DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `idx_user_coupons_user_id` (`user_id`),
  KEY `idx_user_coupons_coupon_id` (`coupon_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 配置/统计/日志
CREATE TABLE IF NOT EXISTS `system_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `config_key` VARCHAR(100) NOT NULL,
  `config_value` TEXT,
  `config_type` VARCHAR(20) DEFAULT 'string',
  `description` VARCHAR(200) DEFAULT NULL,
  `status` TINYINT DEFAULT 1,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_system_configs_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `banners` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `title` VARCHAR(100) DEFAULT NULL,
  `image_url` VARCHAR(500) NOT NULL,
  `link_type` TINYINT DEFAULT 1,
  `link_url` VARCHAR(500) DEFAULT NULL,
  `sort` INT DEFAULT 0,
  `status` TINYINT DEFAULT 1,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `order_statistics` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `date` DATE NOT NULL,
  `order_count` INT DEFAULT 0,
  `total_amount` DECIMAL(12,2) DEFAULT 0,
  `avg_amount` DECIMAL(10,2) DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_order_statistics_date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `product_statistics` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `date` DATE NOT NULL,
  `product_id` BIGINT UNSIGNED NOT NULL,
  `view_count` INT DEFAULT 0,
  `sales_count` INT DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_product_statistics_product_id` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `user_statistics` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `date` DATE NOT NULL,
  `new_user_count` INT DEFAULT 0,
  `active_user_count` INT DEFAULT 0,
  `order_user_count` INT DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_statistics_date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `access_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `user_id` BIGINT UNSIGNED DEFAULT NULL,
  `method` VARCHAR(10) NOT NULL,
  `path` VARCHAR(500) NOT NULL,
  `query` TEXT,
  `body` TEXT,
  `user_agent` VARCHAR(500),
  `ip` VARCHAR(50),
  `status_code` INT NOT NULL,
  `latency` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_access_logs_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `operation_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(32) NOT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED DEFAULT 0,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT 0,
  `deleted_at` DATETIME NULL,
  `is_deleted` TINYINT(1) DEFAULT 0,

  `user_id` BIGINT UNSIGNED NOT NULL,
  `module` VARCHAR(50) NOT NULL,
  `operation` VARCHAR(100) NOT NULL,
  `description` VARCHAR(500),
  `request_data` TEXT,
  `ip` VARCHAR(50),
  `user_agent` VARCHAR(500),
  PRIMARY KEY (`id`),
  KEY `idx_operation_logs_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

SET FOREIGN_KEY_CHECKS=1;

-- EOF
