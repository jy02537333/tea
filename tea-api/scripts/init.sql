-- 茶心阁小程序数据库初始化脚本

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `tea_shop` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `tea_shop`;

-- 设置时区
SET time_zone = '+08:00';

-- 插入默认系统配置
INSERT INTO `system_configs` (`uid`, `config_key`, `config_value`, `config_type`, `description`, `status`, `created_at`, `updated_at`) VALUES
('sys_cfg_001', 'app_name', '茶心阁', 'string', '应用名称', 1, NOW(), NOW()),
('sys_cfg_002', 'app_version', '1.0.0', 'string', '应用版本', 1, NOW(), NOW()),
('sys_cfg_003', 'delivery_fee', '5.00', 'float', '配送费', 1, NOW(), NOW()),
('sys_cfg_004', 'free_delivery_amount', '30.00', 'float', '免配送费金额', 1, NOW(), NOW()),
('sys_cfg_005', 'withdraw_min_amount', '100.00', 'float', '最小提现金额', 1, NOW(), NOW()),
('sys_cfg_006', 'withdraw_fee_rate', '0.006', 'float', '提现手续费率', 1, NOW(), NOW());

-- 插入默认角色
INSERT INTO `roles` (`uid`, `name`, `display_name`, `description`, `status`, `created_at`, `updated_at`) VALUES
('role_admin', 'admin', '超级管理员', '拥有所有权限的超级管理员', 1, NOW(), NOW()),
('role_manager', 'manager', '店长', '店铺管理员', 1, NOW(), NOW()),
('role_staff', 'staff', '店员', '普通店员', 1, NOW(), NOW()),
('role_user', 'user', '用户', '普通用户', 1, NOW(), NOW());

-- 插入默认权限
INSERT INTO `permissions` (`uid`, `name`, `display_name`, `description`, `module`, `action`, `resource`, `created_at`, `updated_at`) VALUES
-- 用户管理
('perm_user_list', 'user.list', '查看用户列表', '查看所有用户信息', 'user', 'list', 'users', NOW(), NOW()),
('perm_user_create', 'user.create', '创建用户', '创建新用户', 'user', 'create', 'users', NOW(), NOW()),
('perm_user_update', 'user.update', '更新用户', '更新用户信息', 'user', 'update', 'users', NOW(), NOW()),
('perm_user_delete', 'user.delete', '删除用户', '删除用户', 'user', 'delete', 'users', NOW(), NOW()),

-- 商品管理
('perm_product_list', 'product.list', '查看商品列表', '查看所有商品信息', 'product', 'list', 'products', NOW(), NOW()),
('perm_product_create', 'product.create', '创建商品', '创建新商品', 'product', 'create', 'products', NOW(), NOW()),
('perm_product_update', 'product.update', '更新商品', '更新商品信息', 'product', 'update', 'products', NOW(), NOW()),
('perm_product_delete', 'product.delete', '删除商品', '删除商品', 'product', 'delete', 'products', NOW(), NOW()),

-- 订单管理
('perm_order_list', 'order.list', '查看订单列表', '查看所有订单信息', 'order', 'list', 'orders', NOW(), NOW()),
('perm_order_update', 'order.update', '更新订单', '更新订单状态', 'order', 'update', 'orders', NOW(), NOW()),
('perm_order_refund', 'order.refund', '订单退款', '处理订单退款', 'order', 'refund', 'orders', NOW(), NOW()),

-- 系统管理
('perm_system_config', 'system.config', '系统配置', '管理系统配置', 'system', 'config', 'configs', NOW(), NOW()),
('perm_system_log', 'system.log', '查看日志', '查看系统日志', 'system', 'log', 'logs', NOW(), NOW());

-- 插入默认角色权限关联
INSERT INTO `role_permissions` (`uid`, `role_id`, `permission_id`, `created_at`, `updated_at`)
SELECT 
    CONCAT('rp_', r.id, '_', p.id) as uid,
    r.id as role_id,
    p.id as permission_id,
    NOW() as created_at,
    NOW() as updated_at
FROM `roles` r, `permissions` p
WHERE r.name = 'admin'; -- 管理员拥有所有权限

-- 插入商品分类
INSERT INTO `categories` (`uid`, `name`, `description`, `image`, `sort`, `status`, `parent_id`, `created_at`, `updated_at`) VALUES
('cat_tea', '茶叶', '各类优质茶叶', '', 1, 1, 0, NOW(), NOW()),
('cat_drink', '茶饮', '现调茶饮', '', 2, 1, 0, NOW(), NOW()),
('cat_snack', '小食', '茶点小食', '', 3, 1, 0, NOW(), NOW()),
('cat_gift', '礼盒', '茶叶礼盒装', '', 4, 1, 0, NOW(), NOW());

-- 插入示例商品
INSERT INTO `products` (`uid`, `category_id`, `name`, `description`, `images`, `price`, `original_price`, `stock`, `sales`, `status`, `sort`, `is_hot`, `is_new`, `is_recommend`, `created_at`, `updated_at`) VALUES
('prod_001', 1, '西湖龙井', '正宗西湖龙井，清香淡雅', '[]', 168.00, 188.00, 100, 0, 1, 1, 1, 1, 1, NOW(), NOW()),
('prod_002', 1, '铁观音', '安溪铁观音，韵味悠长', '[]', 128.00, 158.00, 50, 0, 1, 2, 1, 0, 1, NOW(), NOW()),
('prod_003', 2, '奶茶', '香浓奶茶', '[]', 15.00, 18.00, 999, 0, 1, 1, 0, 1, 0, NOW(), NOW()),
('prod_004', 2, '柠檬蜂蜜茶', '清香柠檬配蜂蜜', '[]', 12.00, 15.00, 999, 0, 1, 2, 0, 1, 0, NOW(), NOW());

-- 插入轮播图
INSERT INTO `banners` (`uid`, `title`, `image_url`, `link_type`, `link_url`, `sort`, `status`, `created_at`, `updated_at`) VALUES
('banner_001', '欢迎来到茶心阁', 'https://example.com/banner1.jpg', 1, '', 1, 1, NOW(), NOW()),
('banner_002', '新品上市', 'https://example.com/banner2.jpg', 2, '/product/001', 2, 1, NOW(), NOW()),
('banner_003', '限时优惠', 'https://example.com/banner3.jpg', 2, '/activity/001', 3, 1, NOW(), NOW());