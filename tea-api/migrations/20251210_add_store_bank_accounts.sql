-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS `store_bank_accounts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `store_id` BIGINT UNSIGNED NOT NULL,
  `account_type` VARCHAR(32) DEFAULT 'bank' COMMENT 'bank|alipay|wechat',
  `account_name` VARCHAR(128) NOT NULL COMMENT '账户开户名',
  `account_no` VARCHAR(128) NOT NULL COMMENT '账号/收款号（加密/脱敏存储）',
  `bank_name` VARCHAR(128) DEFAULT NULL COMMENT '银行名称/支付渠道名',
  `is_default` TINYINT DEFAULT 0 COMMENT '是否默认提现账户',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_store_bank_accounts_store` (`store_id`),
  CONSTRAINT `fk_store_bank_accounts_store` FOREIGN KEY (`store_id`) REFERENCES `stores`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='门店提现收款账户表（银行卡/支付宝/微信）';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS `store_bank_accounts`;

-- +goose StatementEnd
