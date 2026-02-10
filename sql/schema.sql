-- users
CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(64) NOT NULL DEFAULT '',
  `email` VARCHAR(128) NOT NULL UNIQUE,
  `password_hash` CHAR(32) NOT NULL,
  `avatar` VARCHAR(255) NOT NULL DEFAULT '',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- user_identities
CREATE TABLE IF NOT EXISTS `user_identities` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `provider` VARCHAR(32) NOT NULL, -- google, apple
  `provider_uid` VARCHAR(255) NOT NULL, -- sub from google
  `email` VARCHAR(128) NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_provider_uid` (`provider`, `provider_uid`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- projects
CREATE TABLE IF NOT EXISTS `projects` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `description` TEXT,
  `cover_file_id` BIGINT NOT NULL DEFAULT 0,
  `owner_id` BIGINT UNSIGNED NOT NULL,
  `status` ENUM('active','archived') NOT NULL DEFAULT 'active',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_projects_owner_id` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- project_members
CREATE TABLE IF NOT EXISTS `project_members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `project_id` BIGINT UNSIGNED NOT NULL,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `role` ENUM('owner','admin','developer','viewer') NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_user` (`project_id`,`user_id`),
  KEY `idx_project_members_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- files
CREATE TABLE IF NOT EXISTS `files` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL,
  `file_category` ENUM('text','image','video','audio','binary','archive') NOT NULL,
  `file_format` VARCHAR(50) NOT NULL COMMENT '文件格式，如 png, jpg, mp4, mp3, txt 等',
  `current_version_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '当前版本ID，关联 file_versions.id',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_files_current_version_id` (`current_version_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- project_files
CREATE TABLE IF NOT EXISTS `project_files` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `project_id` BIGINT UNSIGNED NOT NULL,
  `file_id` BIGINT UNSIGNED NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_file` (`project_id`,`file_id`),
  KEY `idx_project_files_file_id` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- file_versions
CREATE TABLE IF NOT EXISTS `file_versions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `file_id` BIGINT UNSIGNED NOT NULL,
  `version_number` BIGINT UNSIGNED NOT NULL,
  `size_bytes` BIGINT UNSIGNED NOT NULL,
  `hash` VARCHAR(128) NOT NULL,
  `storage_key` VARCHAR(512) NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created_by` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_file_version` (`file_id`,`version_number`),
  KEY `idx_file_versions_file_id` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- admins
CREATE TABLE IF NOT EXISTS `admins` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(64) NOT NULL UNIQUE,
  `password_hash` CHAR(32) NOT NULL,
  `role` ENUM('super_admin','admin') NOT NULL DEFAULT 'admin',
  `status` ENUM('active','disabled') NOT NULL DEFAULT 'active',
  `last_login_at` TIMESTAMP NULL DEFAULT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admins_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- software_templates
CREATE TABLE IF NOT EXISTS `software_templates` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `description` TEXT,
  `archive_file_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `created_by` BIGINT UNSIGNED NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_software_templates_created_by` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- softwares
CREATE TABLE IF NOT EXISTS `softwares` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `project_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(128) NOT NULL,
  `description` TEXT,
  `template_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `technology_stack` VARCHAR(128) NOT NULL DEFAULT '',
  `status` ENUM('active','archived') NOT NULL DEFAULT 'active',
  `created_by` BIGINT UNSIGNED NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_softwares_project_id` (`project_id`),
  KEY `idx_softwares_template_id` (`template_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- software_manifests
CREATE TABLE IF NOT EXISTS `software_manifests` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `project_id` BIGINT UNSIGNED NOT NULL,
  `software_id` BIGINT UNSIGNED NOT NULL,
  `manifest_file_id` BIGINT UNSIGNED NOT NULL,
  `manifest_file_version_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `version_description` TEXT,
  `created_by` BIGINT UNSIGNED NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_software_manifests_project_id` (`project_id`),
  KEY `idx_software_manifests_software_id` (`software_id`),
  KEY `idx_software_manifests_manifest_file_id` (`manifest_file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- llm_providers
CREATE TABLE IF NOT EXISTS `llm_providers` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `base_url` VARCHAR(512) NOT NULL DEFAULT '',
  `api_key` TEXT,
  `description` TEXT,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_llm_providers_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- llm_models
CREATE TABLE IF NOT EXISTS `llm_models` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `provider_id` BIGINT UNSIGNED NOT NULL,
  `model_name` VARCHAR(255) NOT NULL,
  `model_type` ENUM('llm','vlm','embedding') NOT NULL DEFAULT 'llm',
  `max_input_tokens` INT NOT NULL DEFAULT 0,
  `max_output_tokens` INT NOT NULL DEFAULT 0,
  `support_stream` BOOLEAN NOT NULL DEFAULT FALSE,
  `support_json` BOOLEAN NOT NULL DEFAULT FALSE,
  `price_input_per_1k` DECIMAL(10,6) NOT NULL DEFAULT 0,
  `price_output_per_1k` DECIMAL(10,6) NOT NULL DEFAULT 0,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_llm_models_provider_model` (`provider_id`,`model_name`,`model_type`),
  KEY `idx_llm_models_provider_id` (`provider_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- llm_usage_logs
CREATE TABLE IF NOT EXISTS `llm_usage_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `llm_model_id` BIGINT UNSIGNED NOT NULL,
  `project_id` BIGINT UNSIGNED NOT NULL,
  `input_tokens` INT NOT NULL DEFAULT 0,
  `output_tokens` INT NOT NULL DEFAULT 0,
  `cache_hit` BOOLEAN NOT NULL DEFAULT FALSE,
  `cost_usd` DECIMAL(10,6) NOT NULL DEFAULT 0,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_llm_usage_logs_model_id` (`llm_model_id`),
  KEY `idx_llm_usage_logs_project_id` (`project_id`),
  KEY `idx_llm_usage_logs_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- agents
CREATE TABLE IF NOT EXISTS `agents` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `description` TEXT,
  `agent_type` ENUM('code','asset','design','test','build','ops') NOT NULL DEFAULT 'code',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_agents_name_type` (`name`,`agent_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- agent_llm_bindings
CREATE TABLE IF NOT EXISTS `agent_llm_bindings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `agent_id` BIGINT UNSIGNED NOT NULL,
  `llm_model_id` BIGINT UNSIGNED NOT NULL,
  `priority` INT NOT NULL DEFAULT 0,
  `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_agent_llm_bindings_agent_id` (`agent_id`),
  KEY `idx_agent_llm_bindings_model_id` (`llm_model_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
