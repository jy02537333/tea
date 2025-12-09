-- 茶心阁 小程序 MySQL 建表脚本（示例）
-- 编码与存储引擎
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 用户表
CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(100) DEFAULT NULL COMMENT '用户名/登录名（可为空）',
  `phone` VARCHAR(32) DEFAULT NULL COMMENT '用户手机号（唯一）',
  `email` VARCHAR(255) DEFAULT NULL COMMENT '邮箱地址',
  `password_hash` VARCHAR(255) DEFAULT NULL COMMENT '密码哈希',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '账户状态：1=激活,0=禁用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表（包含消费者、合伙人、管理员等）';

CREATE TABLE IF NOT EXISTS `user_profiles` (
  `user_id` BIGINT UNSIGNED NOT NULL,
  `nickname` VARCHAR(64) DEFAULT NULL COMMENT '昵称',
  `avatar` VARCHAR(512) DEFAULT NULL COMMENT '头像 URL',
  `gender` TINYINT DEFAULT NULL COMMENT '性别：0未知/1男/2女',
  `birthday` DATE DEFAULT NULL COMMENT '生日',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`user_id`),
  CONSTRAINT `fk_user_profiles_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户扩展信息表（昵称/头像等）';

-- 权限与角色
CREATE TABLE IF NOT EXISTS `roles` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(64) NOT NULL COMMENT '角色名称（如 super_admin/store_admin/staff）',
  `description` TEXT COMMENT '角色说明',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='平台角色表（权限角色）';

CREATE TABLE IF NOT EXISTS `user_roles` (
  `user_id` BIGINT UNSIGNED NOT NULL,
  `role_id` INT UNSIGNED NOT NULL,
  `assigned_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '分配时间',
  PRIMARY KEY (`user_id`,`role_id`),
  CONSTRAINT `fk_user_roles_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_roles_role` FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户与角色关联表';

-- 门店
CREATE TABLE IF NOT EXISTS `stores` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(200) NOT NULL COMMENT '门店名称',
  `address` VARCHAR(500) DEFAULT NULL COMMENT '详细地址',
  `lat` DECIMAL(10,7) DEFAULT NULL COMMENT '纬度',
  `lng` DECIMAL(10,7) DEFAULT NULL COMMENT '经度',
  `phone` VARCHAR(64) DEFAULT NULL COMMENT '联系电话',
  `open_hours` VARCHAR(200) DEFAULT NULL COMMENT '营业时间描述',
  `status` TINYINT DEFAULT 1 COMMENT '门店状态：1=正常,0=关闭',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='门店信息表';

CREATE TABLE IF NOT EXISTS `store_staff` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `store_id` BIGINT UNSIGNED NOT NULL,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `role` VARCHAR(64) DEFAULT NULL COMMENT '店员角色（如店长/收银）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
  PRIMARY KEY (`id`),
  KEY `idx_store_staff_store` (`store_id`),
  CONSTRAINT `fk_store_staff_store` FOREIGN KEY (`store_id`) REFERENCES `stores`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_store_staff_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='门店员工表';

-- 商品与分类
CREATE TABLE IF NOT EXISTS `categories` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `parent_id` INT UNSIGNED DEFAULT NULL,
  `name` VARCHAR(128) NOT NULL COMMENT '分类名称',
  `sort_order` INT DEFAULT 0 COMMENT '排序权重',
  PRIMARY KEY (`id`),
  KEY `idx_categories_parent` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品分类表（支持二级）';

CREATE TABLE IF NOT EXISTS `products` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `sku` VARCHAR(128) DEFAULT NULL COMMENT '商品中心编码（可选）',
  `title` VARCHAR(256) NOT NULL COMMENT '商品标题/名称',
  `description` TEXT COMMENT '商品描述',
  `type` VARCHAR(32) DEFAULT 'mall' COMMENT '商品类型：mall=商城,store=门店,points=积分',
  `status` TINYINT DEFAULT 1 COMMENT '上下架状态：1=上架',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  FULLTEXT KEY `ft_products_title` (`title`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品主表（SPU）';

CREATE TABLE IF NOT EXISTS `product_images` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `product_id` BIGINT UNSIGNED NOT NULL,
  `url` VARCHAR(1024) NOT NULL COMMENT '图片 URL',
  `sort` INT DEFAULT 0 COMMENT '图片排序',
  PRIMARY KEY (`id`),
  KEY `idx_product_images_product` (`product_id`),
  CONSTRAINT `fk_product_images_product` FOREIGN KEY (`product_id`) REFERENCES `products`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品图片表';

CREATE TABLE IF NOT EXISTS `product_skus` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `product_id` BIGINT UNSIGNED NOT NULL,
  `sku_code` VARCHAR(128) DEFAULT NULL COMMENT 'SKU 编码',
  `spec` VARCHAR(255) DEFAULT NULL COMMENT '规格描述（如：100g/礼盒）',
  `price` INT NOT NULL COMMENT '销售价（分）',
  `market_price` INT DEFAULT 0 COMMENT '市场价（分）',
  `cost_price` INT DEFAULT 0 COMMENT '成本价（分）',
  `stock` INT DEFAULT 0 COMMENT '库存数量',
  `weight` INT DEFAULT 0 COMMENT '重量（克）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_product_skus_product` (`product_id`),
  CONSTRAINT `fk_product_skus_product` FOREIGN KEY (`product_id`) REFERENCES `products`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品 SKU 表（规格与价格）';

-- 门店与商品映射（门店特供与不同门店价格/上下架）
CREATE TABLE IF NOT EXISTS `store_products` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `store_id` BIGINT UNSIGNED NOT NULL,
  `product_id` BIGINT UNSIGNED NOT NULL,
  `sku_id` BIGINT UNSIGNED DEFAULT NULL,
  `price` INT DEFAULT NULL COMMENT '门店价（分），若为空则使用 SKU 价格',
  `is_delivery` TINYINT DEFAULT 1 COMMENT '是否支持外卖（1=是）',
  `is_dinein` TINYINT DEFAULT 1 COMMENT '是否支持堂食（1=是）',
  `status` TINYINT DEFAULT 1 COMMENT '门店商品状态：1=可售',
  PRIMARY KEY (`id`),
  KEY `idx_store_products_store` (`store_id`),
  CONSTRAINT `fk_store_products_store` FOREIGN KEY (`store_id`) REFERENCES `stores`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_store_products_product` FOREIGN KEY (`product_id`) REFERENCES `products`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='门店商品映射表（门店特供/价格/上下架）';

-- 库存（按 sku & store）
CREATE TABLE IF NOT EXISTS `inventories` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `sku_id` BIGINT UNSIGNED NOT NULL,
  `store_id` BIGINT UNSIGNED DEFAULT NULL,
  `available` INT NOT NULL DEFAULT 0 COMMENT '可售库存',
  `reserved` INT NOT NULL DEFAULT 0 COMMENT '已预占库存',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_inventories_sku` (`sku_id`),
  CONSTRAINT `fk_inventories_sku` FOREIGN KEY (`sku_id`) REFERENCES `product_skus`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存表（按 SKU 与门店分）';

-- 购物车
CREATE TABLE IF NOT EXISTS `cart_items` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `store_id` BIGINT UNSIGNED DEFAULT NULL,
  `product_id` BIGINT UNSIGNED NOT NULL,
  `sku_id` BIGINT UNSIGNED DEFAULT NULL,
  `quantity` INT NOT NULL DEFAULT 1 COMMENT '数量',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入购物车时间',
  PRIMARY KEY (`id`),
  KEY `idx_cart_user` (`user_id`),
  CONSTRAINT `fk_cart_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='购物车项表';

-- 地址
CREATE TABLE IF NOT EXISTS `addresses` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `contact_name` VARCHAR(128) DEFAULT NULL COMMENT '收货人姓名',
  `contact_phone` VARCHAR(64) DEFAULT NULL COMMENT '收货人电话',
  `province` VARCHAR(128) DEFAULT NULL COMMENT '省',
  `city` VARCHAR(128) DEFAULT NULL COMMENT '市',
  `district` VARCHAR(128) DEFAULT NULL COMMENT '区/县',
  `detail` VARCHAR(512) DEFAULT NULL COMMENT '详细地址',
  `is_default` TINYINT DEFAULT 0 COMMENT '是否默认地址',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_address_user` (`user_id`),
  CONSTRAINT `fk_addresses_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户收货地址表';

-- 订单与明细
CREATE TABLE IF NOT EXISTS `orders` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_no` VARCHAR(64) NOT NULL,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `store_id` BIGINT UNSIGNED DEFAULT NULL,
  `type` VARCHAR(32) NOT NULL COMMENT '订单类型：mall=商城,store=门店,delivery=外卖,points=积分',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '订单状态',
  `total_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '订单总金额（分）',
  `pay_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '实际支付金额（分）',
  `coupon_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '使用的优惠券ID',
  `address_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '收货地址ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_orders_order_no` (`order_no`),
  KEY `idx_orders_user` (`user_id`),
  CONSTRAINT `fk_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';

CREATE TABLE IF NOT EXISTS `order_items` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT UNSIGNED NOT NULL,
  `product_id` BIGINT UNSIGNED NOT NULL,
  `sku_id` BIGINT UNSIGNED DEFAULT NULL,
  `quantity` INT NOT NULL DEFAULT 1 COMMENT '购买数量',
  `unit_price` INT NOT NULL COMMENT '下单时单价（分）',
  `total_price` BIGINT NOT NULL COMMENT '小计金额（分）',
  PRIMARY KEY (`id`),
  KEY `idx_order_items_order` (`order_id`),
  CONSTRAINT `fk_order_items_order` FOREIGN KEY (`order_id`) REFERENCES `orders`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单明细表';

-- 支付流水（渠道回调记录）
CREATE TABLE IF NOT EXISTS `payments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT UNSIGNED NOT NULL,
  `channel` VARCHAR(32) NOT NULL COMMENT '支付渠道（wechat/alipay）',
  `channel_trade_no` VARCHAR(128) DEFAULT NULL COMMENT '第三方支付单号',
  `amount` BIGINT NOT NULL COMMENT '支付金额（分）',
  `status` VARCHAR(32) DEFAULT 'initiated' COMMENT '支付状态',
  `payload` JSON DEFAULT NULL COMMENT '原始回调负载',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_payments_order` (`order_id`),
  CONSTRAINT `fk_payments_order` FOREIGN KEY (`order_id`) REFERENCES `orders`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付流水表（记录渠道回调）';

-- 优惠券模板与实例
CREATE TABLE IF NOT EXISTS `coupons_templates` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `type` VARCHAR(32) DEFAULT 'amount' COMMENT '券类型：amount=减金额,percent=折扣',
  `value` INT NOT NULL COMMENT '数值：amount为分，percent为百分比整数',
  `min_order_amount` INT DEFAULT 0 COMMENT '使用门槛（分）',
  `total_quantity` INT DEFAULT 0 COMMENT '总发行量（0为不限）',
  `valid_from` DATETIME DEFAULT NULL COMMENT '生效时间',
  `valid_to` DATETIME DEFAULT NULL COMMENT '失效时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券模板表';

CREATE TABLE IF NOT EXISTS `coupons` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `template_id` BIGINT UNSIGNED NOT NULL,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `code` VARCHAR(128) DEFAULT NULL COMMENT '券码（可选）',
  `status` VARCHAR(32) DEFAULT 'unused' COMMENT '券状态：unused/used/expired',
  `expires_at` DATETIME DEFAULT NULL COMMENT '过期时间',
  `claimed_at` DATETIME DEFAULT NULL COMMENT '领取时间',
  PRIMARY KEY (`id`),
  KEY `idx_coupons_user` (`user_id`),
  CONSTRAINT `fk_coupons_template` FOREIGN KEY (`template_id`) REFERENCES `coupons_templates`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_coupons_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户优惠券实例表';

-- 钱包与流水
CREATE TABLE IF NOT EXISTS `wallets` (
  `user_id` BIGINT UNSIGNED NOT NULL,
  `balance` BIGINT NOT NULL DEFAULT 0 COMMENT '可用余额（分）',
  `frozen` BIGINT NOT NULL DEFAULT 0 COMMENT '冻结金额（分）',
  PRIMARY KEY (`user_id`),
  CONSTRAINT `fk_wallets_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户钱包表（余额/冻结）';

CREATE TABLE IF NOT EXISTS `wallet_transactions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `type` VARCHAR(64) NOT NULL COMMENT '流水类型：recharge/consume/refund/commission/withdraw',
  `amount` BIGINT NOT NULL COMMENT '变动金额（分）',
  `balance_after` BIGINT DEFAULT NULL COMMENT '变动后余额（分）',
  `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_wallet_tx_user` (`user_id`),
  CONSTRAINT `fk_wallet_tx_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='钱包流水表';

-- 积分流水
CREATE TABLE IF NOT EXISTS `points_transactions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `change` INT NOT NULL COMMENT '积分变动（正为增加，负为减少）',
  `reason` VARCHAR(255) DEFAULT NULL COMMENT '变动原因',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_points_user` (`user_id`),
  CONSTRAINT `fk_points_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分流水表';

-- 会员与合伙人礼包
CREATE TABLE IF NOT EXISTS `membership_packages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `price` BIGINT NOT NULL COMMENT '价格（分）',
  `tea_coin_award` BIGINT DEFAULT 0 COMMENT '赠送茶币（分）',
  `discount_rate` DECIMAL(5,2) DEFAULT 1.00 COMMENT '消费折扣（0.95 表示95折）',
  `purchase_discount_rate` DECIMAL(5,2) DEFAULT 1.00 COMMENT '拿货/购买折扣（合伙人拿货折扣）',
  `direct_commission_rate` DECIMAL(5,2) DEFAULT 0.00 COMMENT '直推佣金比例（如 0.30 表示30%）',
  `team_commission_rate` DECIMAL(5,2) DEFAULT 0.00 COMMENT '团队管理奖比例（如 0.10 表示10%）',
  `upgrade_reward_rate` DECIMAL(5,2) DEFAULT 0.00 COMMENT '升级差价奖励比例',
  `type` VARCHAR(64) DEFAULT 'membership' COMMENT '套餐类型：membership=会员,partner_package=合伙人礼包',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会员/合伙人礼包配置表';

CREATE TABLE IF NOT EXISTS `user_memberships` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `package_id` BIGINT UNSIGNED NOT NULL,
  `status` VARCHAR(32) DEFAULT 'active' COMMENT '状态：active/expired/cancelled',
  `started_at` DATETIME DEFAULT NULL COMMENT '生效时间',
  `expires_at` DATETIME DEFAULT NULL COMMENT '到期时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_members_user` (`user_id`),
  CONSTRAINT `fk_user_members_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
  CONSTRAINT `fk_user_members_package` FOREIGN KEY (`package_id`) REFERENCES `membership_packages`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户会员或合伙人购买记录表';

-- 推荐关系与佣金
CREATE TABLE IF NOT EXISTS `referrals` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `referred_user_id` BIGINT UNSIGNED NOT NULL,
  `referrer_user_id` BIGINT UNSIGNED NOT NULL,
  `source` VARCHAR(64) DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '关联建立时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_referrals_referred` (`referred_user_id`),
  CONSTRAINT `fk_referrals_referred_user` FOREIGN KEY (`referred_user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_referrals_referrer_user` FOREIGN KEY (`referrer_user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='推荐关系表（记录上级与下级）';

CREATE TABLE IF NOT EXISTS `commissions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `order_id` BIGINT UNSIGNED DEFAULT NULL,
  `order_item_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '如按明细计算则关联 order_items.id',
  `package_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联的 membership_packages 或 partner package',
  `level_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '合伙人等级，关联 partner_levels.id（若适用）',
  `commission_type` VARCHAR(32) NOT NULL DEFAULT 'direct' COMMENT '佣金类型：direct|indirect|upgrade',
  `source_user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '触发佣金的下单/被推荐用户ID，或被升级用户ID',
  `rate` DECIMAL(7,4) DEFAULT NULL COMMENT '使用的佣金比例（例如 0.3000 表示 30%）',
  `calculation_basis` BIGINT DEFAULT NULL COMMENT '计算基数（分），例如：商品售价 - 优惠 - 运费',
  `gross_amount` BIGINT NOT NULL COMMENT '毛佣金额（分），等于 calculation_basis * rate（未扣手续费）',
  `fee` BIGINT DEFAULT 0 COMMENT '平台扣除手续费（分），如提现手续费在结算时处理',
  `net_amount` BIGINT NOT NULL COMMENT '净佣金额（分）= gross_amount - fee',
  `status` VARCHAR(32) DEFAULT 'frozen' COMMENT '佣金状态：frozen/available/paid/reversed',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `available_at` DATETIME DEFAULT NULL COMMENT '可提现/可用时间',
  PRIMARY KEY (`id`),
  KEY `idx_commissions_user` (`user_id`),
  KEY `idx_commissions_status_available` (`status`,`available_at`),
  CONSTRAINT `fk_commissions_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='佣金记录表（冻结/可用/已发放）';

CREATE TABLE IF NOT EXISTS `commission_transactions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `commission_id` BIGINT UNSIGNED NOT NULL,
  `type` VARCHAR(64) NOT NULL COMMENT '交易类型：release|withdraw|adjust|upgrade_reward|fee',
  `amount` BIGINT NOT NULL COMMENT '变动金额（分）',
  `balance_after` BIGINT DEFAULT NULL COMMENT '变动后佣金余额（分）',
  `operator_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '操作者（管理员）',
  `external_txn_id` VARCHAR(128) DEFAULT NULL COMMENT '外部打款/渠道流水号',
  `note` VARCHAR(255) DEFAULT NULL COMMENT '备注/凭证',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_commission_tx_commission` FOREIGN KEY (`commission_id`) REFERENCES `commissions`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='佣金流水表';

-- 平台财务/提现设置（便于运营配置：提现门槛、手续费比例、发票阈值等）
CREATE TABLE IF NOT EXISTS `platform_financial_settings` (
  `k` VARCHAR(128) NOT NULL,
  `v` VARCHAR(255) DEFAULT NULL,
  `description` VARCHAR(255) DEFAULT NULL,
  PRIMARY KEY (`k`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='平台财务配置：key-value 存储';

-- 示例设置（金额以分为单位；比例为小数）
INSERT IGNORE INTO `platform_financial_settings` (`k`,`v`,`description`) VALUES
  ('withdrawal_min_cents','10000','单次提现最小金额（分），例如 10000 = 100 元'),
  ('withdrawal_fee_rate','0.01','提现手续费比例，例如 0.01 = 1%'),
  ('withdrawal_auto_pay_threshold_cents','0','自动打款阈值（分），低于该值需人工审核，0=无'),
  ('commission_payout_min_cents','10000','佣金自动结算到余额的最低金额（分），例如 10000 = 100 元'),
  ('monthly_invoice_threshold_cents','500000','每月累计提现超过该金额需提供营销发票（分），例如 500000 = 5000 元');

-- 打印任务队列
CREATE TABLE IF NOT EXISTS `print_jobs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `store_id` BIGINT UNSIGNED NOT NULL,
  `order_id` BIGINT UNSIGNED DEFAULT NULL,
  `template_type` VARCHAR(32) DEFAULT 'receipt',
  `payload` JSON DEFAULT NULL COMMENT '打印数据（JSON）',
  `status` VARCHAR(32) DEFAULT 'pending' COMMENT '任务状态：pending/processing/done/failed',
  `retries` INT DEFAULT 0 COMMENT '重试次数',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_print_jobs_store` FOREIGN KEY (`store_id`) REFERENCES `stores`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='门店打印任务队列表';

-- 通知/消息
CREATE TABLE IF NOT EXISTS `notifications` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED DEFAULT NULL,
  `type` VARCHAR(64) DEFAULT NULL COMMENT '通知类型（order/commission/activity 等）',
  `title` VARCHAR(255) DEFAULT NULL COMMENT '通知标题',
  `content` TEXT COMMENT '通知内容',
  `payload` JSON DEFAULT NULL COMMENT '额外负载数据',
  `status` VARCHAR(32) DEFAULT 'unread' COMMENT '状态：unread/read/archived',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_notifications_user` (`user_id`),
  CONSTRAINT `fk_notifications_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户通知表（小程序/公众号推送记录）';

-- 活动与报名
CREATE TABLE IF NOT EXISTS `activities` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `title` VARCHAR(255) NOT NULL COMMENT '活动标题',
  `start_at` DATETIME DEFAULT NULL COMMENT '开始时间',
  `end_at` DATETIME DEFAULT NULL COMMENT '结束时间',
  `location` VARCHAR(255) DEFAULT NULL COMMENT '活动地址或链接',
  `details` TEXT COMMENT '活动详情',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='活动表（线下/线上活动）';

CREATE TABLE IF NOT EXISTS `activity_registrations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `activity_id` BIGINT UNSIGNED NOT NULL,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `contact_name` VARCHAR(128) DEFAULT NULL COMMENT '报名人姓名',
  `contact_phone` VARCHAR(64) DEFAULT NULL COMMENT '报名人电话',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '报名时间',
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_activity_reg_activity` FOREIGN KEY (`activity_id`) REFERENCES `activities`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_activity_reg_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='活动报名表';

-- 运费模版
CREATE TABLE IF NOT EXISTS `shipping_templates` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL COMMENT '运费模版名称',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='运费模板表';

CREATE TABLE IF NOT EXISTS `shipping_rules` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `template_id` BIGINT UNSIGNED NOT NULL,
  `region` VARCHAR(128) DEFAULT NULL COMMENT '适用区域',
  `base_amount` INT DEFAULT 0 COMMENT '基础运费（分）',
  `per_unit_amount` INT DEFAULT 0 COMMENT '单位额外费用（分/单位）',
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_shipping_rules_template` FOREIGN KEY (`template_id`) REFERENCES `shipping_templates`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='运费规则表（按区域）';

-- 审计日志
CREATE TABLE IF NOT EXISTS `admin_audit_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_user_id` BIGINT UNSIGNED NOT NULL,
  `action` VARCHAR(128) NOT NULL COMMENT '操作行为（如: create_product）',
  `target` VARCHAR(128) DEFAULT NULL COMMENT '操作目标（如: product:123）',
  `payload` JSON DEFAULT NULL COMMENT '操作负载/快照',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_admin_audit_admin` FOREIGN KEY (`admin_user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员操作审计日志表';

-- 权限与权限映射（细粒度权限控制）
CREATE TABLE IF NOT EXISTS `permissions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(128) NOT NULL COMMENT '权限代码（唯一，如 product:create）',
  `name` VARCHAR(128) NOT NULL COMMENT '权限名称',
  `description` TEXT COMMENT '权限描述',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_permissions_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限项表（细粒度）';

CREATE TABLE IF NOT EXISTS `role_permissions` (
  `role_id` INT UNSIGNED NOT NULL,
  `permission_id` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`role_id`,`permission_id`),
  CONSTRAINT `fk_role_permissions_role` FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_role_permissions_permission` FOREIGN KEY (`permission_id`) REFERENCES `permissions`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色与权限映射表';

-- 合伙人等级与策略表（可独立于礼包存在，用于管理等级策略）
CREATE TABLE IF NOT EXISTS `partner_levels` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL COMMENT '等级名称（初级/中级/高级）',
  `purchase_discount_rate` DECIMAL(5,2) DEFAULT 1.00 COMMENT '拿货折扣（如 0.70 表示7折）',
  `direct_commission_rate` DECIMAL(5,2) DEFAULT 0.00 COMMENT '直推佣金比例（如 0.30 表示30%）',
  `team_commission_rate` DECIMAL(5,2) DEFAULT 0.00 COMMENT '团队佣金比例（如 0.10 表示10%）',
  `upgrade_reward_rate` DECIMAL(5,2) DEFAULT 0.00 COMMENT '升级差价奖励比例',
  `note` VARCHAR(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='合伙人等级策略表';

-- 用户银行卡/收款账户
CREATE TABLE IF NOT EXISTS `user_bank_accounts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `account_type` VARCHAR(32) DEFAULT 'bank' COMMENT 'bank|alipay|wechat',
  `account_name` VARCHAR(128) NOT NULL COMMENT '账户开户名',
  `account_no` VARCHAR(128) NOT NULL COMMENT '账号/收款号（加密/脱敏存储）',
  `bank_name` VARCHAR(128) DEFAULT NULL COMMENT '银行名称/支付渠道名',
  `is_default` TINYINT DEFAULT 0 COMMENT '是否默认提现账户',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_user_bank_accounts_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户提现收款账户表（银行卡/支付宝/微信）';

-- 提现申请/请求（用于记录提现流程与发票要求）
CREATE TABLE IF NOT EXISTS `withdrawal_requests` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `amount` BIGINT NOT NULL COMMENT '申请提现金额（分）',
  `fee` BIGINT DEFAULT 0 COMMENT '手续费（分）',
  `status` VARCHAR(32) DEFAULT 'pending' COMMENT 'pending/approved/paid/rejected',
  `bank_account_id` BIGINT UNSIGNED DEFAULT NULL,
  `invoice_required` TINYINT DEFAULT 0 COMMENT '是否需要发票',
  `invoice_no` VARCHAR(128) DEFAULT NULL,
  `remark` VARCHAR(255) DEFAULT NULL,
  `requested_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `processed_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_withdrawal_requests_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_withdrawal_requests_bank` FOREIGN KEY (`bank_account_id`) REFERENCES `user_bank_accounts`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提现申请表';

-- 推荐闭包表（支持多层分销查询与快速聚合）
CREATE TABLE IF NOT EXISTS `referrals_closure` (
  `ancestor_user_id` BIGINT UNSIGNED NOT NULL,
  `descendant_user_id` BIGINT UNSIGNED NOT NULL,
  `depth` INT NOT NULL DEFAULT 0 COMMENT '距离层级，0 表示同一用户',
  PRIMARY KEY (`ancestor_user_id`,`descendant_user_id`),
  CONSTRAINT `fk_referrals_closure_ancestor` FOREIGN KEY (`ancestor_user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_referrals_closure_descendant` FOREIGN KEY (`descendant_user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='推荐闭包表（支撑 N 层分销计算）';

-- ========== 示例数据：会员/合伙人等级配置（依据需求文档 2.2） ===========
-- 注意：金额与茶币均以分为单位；比例字段为小数（例如 0.95 表示 95 折，0.30 表示 30%）
INSERT IGNORE INTO `membership_packages` (`id`,`name`,`price`,`tea_coin_award`,`discount_rate`,`purchase_discount_rate`,`direct_commission_rate`,`team_commission_rate`,`upgrade_reward_rate`,`type`) VALUES
  (1, '普通用户', 0, 0, 1.00, 1.00, 0.10, 0.00, 0.00, 'membership'),
  (2, '客随主泡会员', 36500, 0, 0.95, 1.00, 0.10, 0.00, 0.00, 'membership'),
  (3, '黄金会员', 199900, 50000, 0.90, 1.00, 0.10, 0.00, 0.00, 'membership'),
  (4, '初级合伙人', 680000, 680000, 0.70, 0.70, 0.30, 0.10, 0.00, 'partner_package'),
  (5, '中级合伙人', 980000, 980000, 0.60, 0.60, 0.40, 0.10, 0.00, 'partner_package'),
  (6, '高级合伙人', 1680000, 1680000, 0.55, 0.55, 0.45, 0.10, 0.00, 'partner_package'),
  (7, '门店管理员', 0, 0, 0.45, 0.45, 0.55, 0.00, 0.00, 'membership');

-- 将合伙人等级同步写入 `partner_levels`（供策略/运营配置使用）
INSERT IGNORE INTO `partner_levels` (`id`,`name`,`purchase_discount_rate`,`direct_commission_rate`,`team_commission_rate`,`upgrade_reward_rate`,`note`) VALUES
  (1, '初级合伙人', 0.70, 0.30, 0.10, 0.00, '与需求一致：拿货7折，直推30%，团队10%'),
  (2, '中级合伙人', 0.60, 0.40, 0.10, 0.00, '与需求一致：拿货6折，直推40%，团队10%'),
  (3, '高级合伙人', 0.55, 0.45, 0.10, 0.00, '与需求一致：拿货5.5折，直推45%，团队10%'),
  (4, '门店管理员', 0.45, 0.55, 0.00, 0.00, '门店管理员特殊等级：拿货4.5折，直推55%');

-- 说明：
-- 1) `membership_packages` 中同时保存“消费折扣（discount_rate）”与“拿货折扣（purchase_discount_rate）”，方便在结算时区分消费者购买折扣与合伙人/门店拿货成本折扣。
-- 2) `tea_coin_award` 以分为单位，示例中将文档里的“500/6800/9800/16800”按人民币元转换为分（*100）。
-- 3) 如需不同运营活动或差异化配置，可继续在 `membership_packages` 新增行或由运营在后台编辑。

SET FOREIGN_KEY_CHECKS = 1;
