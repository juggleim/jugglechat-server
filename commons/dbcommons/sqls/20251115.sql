ALTER TABLE `users` ADD COLUMN `status` TINYINT NULL DEFAULT 0;

INSERT INTO `globalconfs` (`conf_key`,`conf_value`)VALUES('jchatdb_version','20251115');