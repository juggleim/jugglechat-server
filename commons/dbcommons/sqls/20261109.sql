CREATE TABLE IF NOT EXISTS password_sms_challenges (
  id char(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL PRIMARY KEY,
  phone varchar(32) NOT NULL,
  phone_version bigint NOT NULL DEFAULT 0,
  digest char(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  nonce char(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  expires_at bigint NOT NULL DEFAULT 0,
  failures int NOT NULL DEFAULT 0,
  delivered boolean NOT NULL DEFAULT false
) ENGINE=InnoDB;
SET @password_sms_ddl = (SELECT IF(COUNT(*)=0,'ALTER TABLE users ADD COLUMN password_sms_phone_version BIGINT NOT NULL DEFAULT 0','SELECT 1') FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='users' AND column_name='password_sms_phone_version');
PREPARE password_sms_stmt FROM @password_sms_ddl;
EXECUTE password_sms_stmt;
DEALLOCATE PREPARE password_sms_stmt;
-- Phone version trigger is installed by the explicit Go migration step.
CREATE TABLE IF NOT EXISTS password_sms_quotas (
  id char(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL PRIMARY KEY,
  last_send bigint NOT NULL DEFAULT 0,
  attempt_nonce char(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  hour_start bigint NOT NULL DEFAULT 0,
  hour_count int NOT NULL DEFAULT 0,
  day_start bigint NOT NULL DEFAULT 0,
  day_count int NOT NULL DEFAULT 0
) ENGINE=InnoDB;
