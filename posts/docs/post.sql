CREATE TABLE IF NOT EXISTS `posts` (
  `id` int NOT NULL AUTO_INCREMENT,
  `post_id` varchar(32) DEFAULT NULL,
  `title` varchar(200) DEFAULT NULL,
  `content` mediumblob,
  `content_brief` varchar(5000) DEFAULT NULL,
  `is_delete` tinyint DEFAULT '0',
  `user_id` varchar(32) DEFAULT NULL,
  `post_exset` mediumblob,
  `created_time` bigint DEFAULT '0',
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `status` tinyint DEFAULT '0',
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `uniq_id` (`app_key`,`post_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `postcomments` (
  `id` int NOT NULL AUTO_INCREMENT,
  `comment_id` varchar(32) DEFAULT NULL,
  `post_id` varchar(32) DEFAULT NULL,
  `parent_comment_id` varchar(32) DEFAULT NULL,
  `parent_user_id` varchar(32) DEFAULT NULL,
  `user_id` varchar(32) DEFAULT NULL,
  `text` varchar(5000) DEFAULT NULL,
  `created_time` bigint DEFAULT NULL,
  `updated_time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `is_delete` tinyint DEFAULT '0',
  `status` tinyint DEFAULT '0',
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_id` (`app_key`,`comment_id`),
  KEY `idx_post` (`app_key`,`post_id`,`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `postreactions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `bus_id` varchar(50) DEFAULT NULL,
  `bus_type` tinyint DEFAULT '0',
  `key` varchar(50) DEFAULT NULL,
  `value` varchar(100) DEFAULT NULL,
  `created_time` bigint DEFAULT '0',
  `user_id` varchar(50) DEFAULT NULL,
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_key` (`app_key`,`bus_id`,`bus_type`,`key`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `postfeeds` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` varchar(32) DEFAULT '',
  `post_id` varchar(32) DEFAULT '',
  `feed_time` bigint DEFAULT '0',
  `app_key` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_time` (`app_key`,`user_id`,`feed_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `postcommentfeeds` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` varchar(32) DEFAULT '',
  `comment_id` varchar(32) DEFAULT '',
  `post_id` varchar(32) DEFAULT '',
  `feed_time` bigint DEFAULT '0',
  `app_key` varchar(20) DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_time` (`app_key`,`user_id`,`feed_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO `globalconfs` (`conf_key`,`conf_value`)VALUES('jpostdb_version','20250606');