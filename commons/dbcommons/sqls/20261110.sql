-- Also upgrades databases that already applied the initial password SMS migration.
SET @password_sms_ddl = (SELECT IF(COUNT(*)=0,'ALTER TABLE password_sms_quotas ADD COLUMN attempt_nonce CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT ''''','SELECT 1') FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='password_sms_quotas' AND column_name='attempt_nonce');
PREPARE password_sms_stmt FROM @password_sms_ddl;
EXECUTE password_sms_stmt;
DEALLOCATE PREPARE password_sms_stmt;
