-- 创建客服/投诉工单表，对应 PRD 3.2.9 Ticket 模型

-- +migrate Up
+CREATE TABLE IF NOT EXISTS tickets (
+    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
+    uid VARCHAR(32) NOT NULL UNIQUE,
+    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
+    created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
+    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
+    updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
+    deleted_at DATETIME NULL,
+    is_deleted TINYINT(1) NOT NULL DEFAULT 0,
+
+    type VARCHAR(30) NOT NULL,
+    source VARCHAR(30) NOT NULL,
+    user_id BIGINT UNSIGNED NULL,
+    order_id BIGINT UNSIGNED NULL,
+    store_id BIGINT UNSIGNED NULL,
+    title VARCHAR(200) NOT NULL,
+    content TEXT NULL,
+    attachments TEXT NULL,
+    status VARCHAR(20) NOT NULL DEFAULT 'new',
+    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
+    assignee_id BIGINT UNSIGNED NULL,
+    remark TEXT NULL,
+    reject_reason TEXT NULL,
+    resolved_at DATETIME NULL,
+    closed_at DATETIME NULL,
+
+    INDEX idx_tickets_status (status),
+    INDEX idx_tickets_type (type),
+    INDEX idx_tickets_source (source),
+    INDEX idx_tickets_user_id (user_id),
+    INDEX idx_tickets_order_id (order_id),
+    INDEX idx_tickets_store_id (store_id),
+    INDEX idx_tickets_assignee_id (assignee_id)
+) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='客服/投诉工单表';
+
+-- +migrate Down
+DROP TABLE IF EXISTS tickets;
