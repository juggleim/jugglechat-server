CREATE TABLE IF NOT EXISTS `appnavs` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `app_key` VARCHAR(50) NULL,
  `alias_no` VARCHAR(50) NULL,
  `api_url` VARCHAR(200) NULL,
  `ws_url` VARCHAR(200) NULL,
  `app_url` VARCHAR(200) NULL,
  `created_time` DATETIME(3) NULL,
  `updated_time` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `app_key_UNIQUE` (`app_key`),
  UNIQUE INDEX `alias_no_UNIQUE` (`alias_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;