CREATE TABLE IF NOT EXISTS user_devices (
 id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
 app_key VARCHAR(64) COLLATE utf8mb4_bin NOT NULL,
 user_id VARCHAR(64) COLLATE utf8mb4_bin NOT NULL,
 device_id VARCHAR(128) COLLATE utf8mb4_bin NOT NULL,
 device_name VARCHAR(128) NOT NULL DEFAULT '',
 platform VARCHAR(32) NOT NULL,
 os_version VARCHAR(64) NOT NULL DEFAULT '',
 app_version VARCHAR(64) NOT NULL DEFAULT '',
 last_login_time BIGINT NOT NULL,
 last_active_time BIGINT NULL,
 session_id VARCHAR(64) COLLATE utf8mb4_bin NOT NULL,
 state VARCHAR(16) NOT NULL,
 revoke_operation_id VARCHAR(64) NOT NULL DEFAULT '',
 banned_until BIGINT NOT NULL DEFAULT 0,
 deleted_at BIGINT NULL,
 created_time BIGINT NOT NULL,
 updated_time BIGINT NOT NULL,
 UNIQUE KEY uk_user_device (app_key,user_id,device_id),
 KEY idx_device_list (app_key,user_id,deleted_at,last_login_time DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS device_revoke_jobs (
 operation_id VARCHAR(64) COLLATE utf8mb4_bin PRIMARY KEY,
 app_key VARCHAR(64) COLLATE utf8mb4_bin NOT NULL,
 user_id VARCHAR(64) COLLATE utf8mb4_bin NOT NULL,
 device_id VARCHAR(128) COLLATE utf8mb4_bin NOT NULL,
 target_session_id VARCHAR(64) COLLATE utf8mb4_bin NOT NULL,
 banned_until BIGINT NOT NULL,
 done BOOLEAN NOT NULL DEFAULT FALSE,
 attempts INT NOT NULL DEFAULT 0,
 next_attempt BIGINT NOT NULL,
 last_error VARCHAR(256) NOT NULL DEFAULT '',
 KEY idx_revoke_pending (done,next_attempt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS device_login_qrs (
 code_id VARCHAR(64) COLLATE utf8mb4_bin NOT NULL,
 app_key VARCHAR(64) COLLATE utf8mb4_bin NOT NULL,
 user_id VARCHAR(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
 status VARCHAR(16) NOT NULL DEFAULT 'pending',
 mode VARCHAR(16) NOT NULL DEFAULT '',
 secret_hash VARCHAR(64) NOT NULL,
 device_id VARCHAR(128) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
 session_id VARCHAR(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
 created_time BIGINT NOT NULL,
 consumed_time BIGINT NOT NULL DEFAULT 0,
 result_cipher MEDIUMTEXT NOT NULL,
 PRIMARY KEY (app_key,code_id),
 KEY idx_qr_expiry (created_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
-- Preserve unexpired QR codes created by the previous release. They remain legacy-only.
INSERT IGNORE INTO device_login_qrs (app_key,code_id,user_id,status,mode,secret_hash,created_time,result_cipher)
SELECT app_key,code_id,COALESCE(user_id,''),CASE WHEN status=1 THEN 'confirmed' ELSE 'pending' END,'legacy','',created_time,'' FROM qrcoderecords WHERE created_time > UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000-600000;
