CREATE TABLE IF NOT EXISTS `accounts` (
  `id` int NOT NULL AUTO_INCREMENT,
  `account` varchar(45) DEFAULT NULL,
  `password` varchar(45) DEFAULT NULL,
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `state` tinyint DEFAULT '0',
  `role_type` tinyint DEFAULT 0,
  `parent_account` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_account` (`account`),
  KEY `idx_parent` (`parent_account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `accountapprels` (
  `id` int NOT NULL AUTO_INCREMENT,
  `app_key` varchar(20) DEFAULT '',
  `account` varchar(20) DEFAULT '',
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_account` (`app_key`,`account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `apps` (
  `id` int NOT NULL AUTO_INCREMENT,
  `app_key` varchar(45) NOT NULL,
  `app_secret` varchar(45) NOT NULL,
  `app_secure_key` varchar(45) NOT NULL,
  `app_status` tinyint DEFAULT '0',
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `app_type` tinyint DEFAULT '0',
  `app_name` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_appkey` (`app_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `appexts` (
  `id` int NOT NULL AUTO_INCREMENT,
  `app_key` varchar(50) DEFAULT NULL,
  `app_item_key` varchar(50) DEFAULT NULL,
  `app_item_value` varchar(2048) DEFAULT NULL,
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_APPKEY_APPITEMKEY` (`app_key`,`app_item_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `globalconfs` (
  `id` int NOT NULL AUTO_INCREMENT,
  `conf_key` varchar(50) DEFAULT NULL,
  `conf_value` varchar(2000) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_key` (`conf_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `fileconfs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `app_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `channel` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `conf` json DEFAULT NULL,
  `enable` tinyint(1) DEFAULT '0',
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_time` datetime(3) DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `app_key` (`app_key`,`channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `friendrels` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(32) DEFAULT NULL,
  `friend_id` varchar(32) DEFAULT NULL,
  `order_tag` varchar(20) NULL DEFAULT '',
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_friend` (`app_key`,`user_id`,`friend_id`),
  KEY `idx_order` (`app_key`, `user_id`, `order_tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `groupadmins` (
  `id` int NOT NULL AUTO_INCREMENT,
  `group_id` varchar(64) DEFAULT NULL,
  `admin_id` varchar(64) DEFAULT NULL,
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_admin` (`app_key`,`group_id`,`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `groupinfos` (
  `id` int NOT NULL AUTO_INCREMENT,
  `group_id` varchar(64) DEFAULT NULL,
  `group_name` varchar(64) DEFAULT NULL,
  `group_portrait` varchar(200) DEFAULT NULL,
  `creator_id` varchar(64) DEFAULT NULL,
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) DEFAULT NULL,
  `is_mute` tinyint DEFAULT '0',
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_appkey_groupid` (`app_key`,`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `groupinfoexts` (
  `id` int NOT NULL AUTO_INCREMENT,
  `group_id` varchar(32) DEFAULT NULL,
  `item_key` varchar(50) DEFAULT NULL,
  `item_value` varchar(100) DEFAULT NULL,
  `item_type` tinyint DEFAULT '0',
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_appkey_groupid` (`app_key`,`group_id`,`item_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `groupmembers` (
  `id` int NOT NULL AUTO_INCREMENT,
  `group_id` varchar(64) DEFAULT NULL,
  `member_id` varchar(64) DEFAULT NULL,
  `member_type` tinyint DEFAULT '0',
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `app_key` varchar(45) DEFAULT NULL,
  `is_mute` tinyint DEFAULT '0',
  `is_allow` tinyint DEFAULT '0',
  `mute_end_at` bigint DEFAULT '0',
  `grp_display_name` varchar(100) DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_appkey_grpid_memid` (`app_key`,`group_id`,`member_id`),
  KEY `idx_memberid` (`app_key`,`member_id`,`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_type` tinyint DEFAULT '0',
  `user_id` varchar(32) NOT NULL,
  `nickname` varchar(50) DEFAULT NULL,
  `user_portrait` varchar(200) DEFAULT NULL,
  `pinyin` varchar(50) DEFAULT NULL,
  `phone` varchar(50) DEFAULT NULL,
  `email` varchar(100) DEFAULT NULL,
  `login_account` varchar(50) DEFAULT NULL,
  `login_pass` varchar(50) DEFAULT NULL,
  `status` tinyint default '0',
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) DEFAULT NULL,
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_userid` (`app_key`,`user_id`),
  UNIQUE KEY `uniq_phone` (`app_key`,`phone`),
  UNIQUE KEY `uniq_email` (`app_key`,`email`),
  UNIQUE KEY `uniq_account` (`app_key`,`login_account`),
  KEY `idx_userid` (`app_key`,`user_type`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `userexts` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(32) DEFAULT NULL,
  `item_key` varchar(50) DEFAULT NULL,
  `item_value` varchar(2000) DEFAULT NULL,
  `item_type` tinyint DEFAULT '0',
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_item_key` (`app_key`,`user_id`,`item_key`),
  KEY `idx_item_key` (`app_key`,`item_key`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `friendapplications` (
  `id` int NOT NULL AUTO_INCREMENT,
  `recipient_id` varchar(32) DEFAULT NULL,
  `sponsor_id` varchar(32) DEFAULT NULL,
  `apply_time` bigint DEFAULT NULL,
  `status` tinyint DEFAULT '0',
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_apply` (`app_key`,`recipient_id`,`sponsor_id`),
  KEY `idx_recipient` (`app_key`,`recipient_id`,`apply_time`),
  KEY `idx_sponsor` (`app_key`,`sponsor_id`,`apply_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `grpapplications` (
  `id` int NOT NULL AUTO_INCREMENT,
  `group_id` varchar(32) DEFAULT NULL,
  `apply_type` tinyint DEFAULT '0',
  `sponsor_id` varchar(32) DEFAULT NULL,
  `recipient_id` varchar(32) DEFAULT NULL,
  `inviter_id` varchar(32) DEFAULT NULL,
  `operator_id` varchar(32) DEFAULT NULL,
  `apply_time` bigint DEFAULT '0',
  `status` tinyint DEFAULT '0',
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_apply` (`app_key`,`group_id`,`apply_type`,`sponsor_id`,`recipient_id`),
  KEY `idx_sponsor` (`app_key`,`apply_type`,`sponsor_id`,`apply_time`),
  KEY `idx_group` (`app_key`,`apply_type`,`group_id`,`apply_time`),
  KEY `idx_recipient` (`app_key`,`apply_type`,`recipient_id`,`apply_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `qrcoderecords` (
  `id` int NOT NULL AUTO_INCREMENT,
  `code_id` varchar(50) DEFAULT NULL,
  `status` tinyint DEFAULT NULL,
  `created_time` bigint DEFAULT NULL,
  `user_id` varchar(32) DEFAULT NULL,
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_id` (`app_key`,`code_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `smsrecords` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `phone` varchar(50) DEFAULT NULL,
  `email` varchar(200) DEFAULT NULL,
  `code` varchar(10) DEFAULT NULL,
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_phone` (`app_key`,`phone`,`created_time`),
  KEY `idx_mail` (`app_key`,`email`,`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `ai_engines` (
  `id` int NOT NULL AUTO_INCREMENT,
  `engine_type` tinyint DEFAULT '0',
  `engine_conf` varchar(5000) DEFAULT NULL,
  `status` tinyint DEFAULT '0',
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_appkey` (`app_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `assistant_prompts` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(32) DEFAULT NULL,
  `prompts` varchar(2000) DEFAULT NULL,
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_id` (`app_key`,`id`),
  KEY `idx_user` (`app_key`,`user_id`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `botconfs` (
  `id` int NOT NULL AUTO_INCREMENT,
  `bot_id` varchar(32) NULL,
  `nickname` varchar(50) DEFAULT NULL,
  `bot_portrait` varchar(200) DEFAULT NULL,
  `description` varchar(500) DEFAULT NULL,
  `bot_type` tinyint DEFAULT '0',
  `bot_conf` varchar(2000) NULL,
  `status` tinyint DEFAULT '0',
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uniq_botid` (`app_key`, `bot_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `telebots` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(50) DEFAULT NULL,
  `bot_name` varchar(50) DEFAULT NULL,
  `bot_token` varchar(200) DEFAULT NULL,
  `status` tinyint DEFAULT NULL,
  `app_key` varchar(20) DEFAULT NULL,
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_bot` (`app_key`,`bot_token`),
  KEY `idx_user` (`app_key`,`user_id`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `telebotrels` (
  `id` int NOT NULL AUTO_INCREMENT,
  `tele_bot_id` varchar(50) NOT NULL,
  `user_id` varchar(50) NOT NULL,
  `bot_token` varchar(100) DEFAULT NULL,
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_botid` (`app_key`,`tele_bot_id`),
  KEY `idx_userid` (`app_key`,`user_id`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `converconfs` (
  `id` int NOT NULL AUTO_INCREMENT,
  `conver_id` varchar(100) DEFAULT '',
  `conver_type` tinyint DEFAULT '0',
  `sub_channel` varchar(32) DEFAULT '',
  `item_key` varchar(100) DEFAULT '',
  `item_value` varchar(2000) DEFAULT '',
  `item_type` tinyint DEFAULT '0',
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_key` (`app_key`,`conver_id`,`conver_type`,`sub_channel`,`item_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `userconverconfs`(
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(32) DEFAULT '',
  `conver_id` varchar(100) DEFAULT '',
  `conver_type` tinyint DEFAULT '0',
  `sub_channel` varchar(32) DEFAULT '',
  `item_key` varchar(100) DEFAULT '',
  `item_value` varchar(2000) DEFAULT '',
  `item_type` tinyint DEFAULT '0',
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_key` (`app_key`,`conver_id`,`conver_type`,`sub_channel`,`item_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `feedbacks` (
  `id` int NOT NULL AUTO_INCREMENT,
  `app_key` varchar(50) DEFAULT NULL,
  `user_id` varchar(32) DEFAULT NULL,
  `category` varchar(100) DEFAULT NULL,
  `content` mediumblob,
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_appkey` (`app_key`,`user_id`),
  KEY `idx_time` (`app_key`,`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `applications` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `app_id` VARCHAR(32) NULL,
  `app_name` VARCHAR(50) NULL,
  `app_icon` VARCHAR(500) NULL,
  `app_desc` VARCHAR(500) NULL,
  `app_url` VARCHAR(500) NULL,
  `created_time` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_time` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `app_order` INT NULL DEFAULT 0,
  `app_key` VARCHAR(20) NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uniq_id` (`app_key`, `app_id`),
  INDEX `idx_order` (`app_key`, `app_order`, `created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `banusers` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(32) NOT NULL,
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `end_time` bigint DEFAULT '0',
  `scope_key` varchar(20) NOT NULL DEFAULT 'default',
  `scope_value` varchar(1000) DEFAULT '',
  `ext` varchar(100) DEFAULT NULL,
  `app_key` varchar(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_appkey_userid` (`app_key`,`user_id`,`scope_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `blocks` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(32) DEFAULT NULL,
  `block_user_id` varchar(32) DEFAULT NULL,
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_appkey_userid` (`app_key`,`user_id`,`block_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `sensitivewords` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `app_key` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `word` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `word_type` tinyint(1) NOT NULL DEFAULT '1' COMMENT '12',
  `created_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_word` (`app_key`,`word`),
  KEY `idx_appkey` (`app_key`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT IGNORE INTO `globalconfs` (`conf_key`,`conf_value`)VALUES('jchatdb_version','20251215');