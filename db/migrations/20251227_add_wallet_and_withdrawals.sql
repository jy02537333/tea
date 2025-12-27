-- Explicit SQL migration: wallets, wallet_transactions, withdrawal_requests
-- Safe to run multiple times due to IF NOT EXISTS

-- 钱包与流水
CREATE TABLE IF NOT EXISTS `wallets` (
  `user_id` BIGINT UNSIGNED NOT NULL,
  `balance` BIGINT NOT NULL DEFAULT 0 COMMENT '可用余额（分）',
  `frozen` BIGINT NOT NULL DEFAULT 0 COMMENT '冻结金额（分）',
  PRIMARY KEY (`user_id`),
  CONSTRAINT `fk_wallets_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户钱包表（余额/冻结)';

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

-- 提现申请
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
