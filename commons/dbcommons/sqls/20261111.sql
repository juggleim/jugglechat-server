CREATE TABLE IF NOT EXISTS converconf_scope_versions (
 scope_id char(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL PRIMARY KEY,
 app_key varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
 conver_id varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
 conver_type tinyint NOT NULL,
 sub_channel varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
 version bigint NOT NULL DEFAULT 0,
 UNIQUE KEY canonical_scope (app_key,conver_id,conver_type,sub_channel)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS converconf_outbox (
 event_id char(36) CHARACTER SET ascii COLLATE ascii_bin NOT NULL PRIMARY KEY,
 scope_id char(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 version bigint NOT NULL,
 app_key varchar(64) NOT NULL,
 actor_id varchar(64) NOT NULL,
 target_id varchar(64) NOT NULL,
 conver_type int NOT NULL,
 payload longtext NOT NULL,
 status varchar(16) NOT NULL DEFAULT 'pending',
 attempts bigint NOT NULL DEFAULT 0,
 next_attempt_at bigint NOT NULL DEFAULT 0,
 lease_until bigint NOT NULL DEFAULT 0,
 claim_token char(36) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
 diagnostic varchar(32) NOT NULL DEFAULT '',
 UNIQUE KEY scope_version (scope_id, version),
 KEY delivery (status, next_attempt_at, lease_until)
) ENGINE=InnoDB;
