package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/anil-wu/spark-x/internal/config"
	"github.com/zeromicro/go-zero/core/conf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	configFile = flag.String("f", "etc/sparkx-api.yaml", "the config file")
	dsn        = flag.String("dsn", "", "mysql dsn (override config)")
)

type UsersTable struct {
	Id           uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Username     string    `gorm:"column:username;type:varchar(64);not null;default:''"`
	Email        string    `gorm:"column:email;type:varchar(128);not null;uniqueIndex:uk_users_email"`
	PasswordHash string    `gorm:"column:password_hash;type:char(32);not null"`
	Avatar       string    `gorm:"column:avatar;type:varchar(255);default:''"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (UsersTable) TableName() string { return "users" }

type UserIdentitiesTable struct {
	Id          uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	UserId      uint64    `gorm:"column:user_id;not null;index:idx_user_id"`
	Provider    string    `gorm:"column:provider;type:varchar(32);not null;uniqueIndex:uk_provider_uid,priority:1"`
	ProviderUid string    `gorm:"column:provider_uid;type:varchar(255);not null;uniqueIndex:uk_provider_uid,priority:2"`
	Email       string    `gorm:"column:email;type:varchar(128);not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (UserIdentitiesTable) TableName() string { return "user_identities" }

type ProjectsTable struct {
	Id          uint64         `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string         `gorm:"column:name;type:varchar(128);not null"`
	Description sql.NullString `gorm:"column:description;type:text"`
	CoverFileId uint64         `gorm:"column:cover_file_id;type:bigint;default:0"`
	OwnerId     uint64         `gorm:"column:owner_id;not null;index:idx_projects_owner_id"`
	Status      string         `gorm:"column:status;type:enum('active','archived');not null;default:'active'"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (ProjectsTable) TableName() string { return "projects" }

type ProjectMembersTable struct {
	Id        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectId uint64    `gorm:"column:project_id;not null;uniqueIndex:uk_project_user,priority:1"`
	UserId    uint64    `gorm:"column:user_id;not null;uniqueIndex:uk_project_user,priority:2;index:idx_project_members_user_id"`
	Role      string    `gorm:"column:role;type:enum('owner','admin','developer','viewer');not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ProjectMembersTable) TableName() string { return "project_members" }

type FilesTable struct {
	Id               uint64         `gorm:"column:id;primaryKey;autoIncrement"`
	Name             string         `gorm:"column:name;type:varchar(255);not null"`
	FileCategory     string         `gorm:"column:file_category;type:enum('text','image','video','audio','binary','archive');not null"`
	FileFormat       string         `gorm:"column:file_format;type:varchar(50);not null;default:''"`
	CurrentVersionId uint64         `gorm:"column:current_version_id;index:idx_files_current_version_id"`
	CreatedAt        time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index;softDelete"`
}

func (FilesTable) TableName() string { return "files" }

type ProjectFilesTable struct {
	Id        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectId uint64    `gorm:"column:project_id;not null;uniqueIndex:uk_project_file,priority:1"`
	FileId    uint64    `gorm:"column:file_id;not null;uniqueIndex:uk_project_file,priority:2;index:idx_project_files_file_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ProjectFilesTable) TableName() string { return "project_files" }

type FileVersionsTable struct {
	Id            uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	FileId        uint64    `gorm:"column:file_id;not null;uniqueIndex:uk_file_version,priority:1;index:idx_file_versions_file_id"`
	VersionNumber uint64    `gorm:"column:version_number;not null;uniqueIndex:uk_file_version,priority:2"`
	SizeBytes     uint64    `gorm:"column:size_bytes;not null"`
	Hash          string    `gorm:"column:hash;type:varchar(128);not null"`
	StorageKey    string    `gorm:"column:storage_key;type:varchar(512);not null"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
	CreatedBy     uint64    `gorm:"column:created_by;not null"`
}

func (FileVersionsTable) TableName() string { return "file_versions" }

type AdminsTable struct {
	Id           uint64       `gorm:"column:id;primaryKey;autoIncrement"`
	Username     string       `gorm:"column:username;type:varchar(64);not null;uniqueIndex:uk_admins_username"`
	PasswordHash string       `gorm:"column:password_hash;type:char(32);not null"`
	Role         string       `gorm:"column:role;type:enum('super_admin','admin');not null;default:'admin'"`
	Status       string       `gorm:"column:status;type:enum('active','disabled');not null;default:'active'"`
	LastLoginAt  sql.NullTime `gorm:"column:last_login_at"`
	CreatedAt    time.Time    `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time    `gorm:"column:updated_at;autoUpdateTime"`
}

func (AdminsTable) TableName() string { return "admins" }

type SoftwareTemplatesTable struct {
	Id            uint64         `gorm:"column:id;primaryKey;autoIncrement"`
	Name          string         `gorm:"column:name;type:varchar(128);not null"`
	Description   sql.NullString `gorm:"column:description;type:text"`
	ArchiveFileId uint64         `gorm:"column:archive_file_id;type:bigint;default:0"`
	CreatedBy     uint64         `gorm:"column:created_by;not null;index:idx_software_templates_created_by"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (SoftwareTemplatesTable) TableName() string { return "software_templates" }

type SoftwaresTable struct {
	Id              uint64         `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectId       uint64         `gorm:"column:project_id;not null;index:idx_softwares_project_id"`
	Name            string         `gorm:"column:name;type:varchar(128);not null"`
	Description     sql.NullString `gorm:"column:description;type:text"`
	TemplateId      uint64         `gorm:"column:template_id;type:bigint;not null;default:0;index:idx_softwares_template_id"`
	TechnologyStack string         `gorm:"column:technology_stack;type:varchar(128);not null;default:''"`
	Status          string         `gorm:"column:status;type:enum('active','archived');not null;default:'active'"`
	CreatedBy       uint64         `gorm:"column:created_by;not null"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (SoftwaresTable) TableName() string { return "softwares" }

type SoftwareManifestsTable struct {
	Id                    uint64         `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectId             uint64         `gorm:"column:project_id;not null;index:idx_software_manifests_project_id"`
	SoftwareId            uint64         `gorm:"column:software_id;not null;index:idx_software_manifests_software_id"`
	ManifestFileId        uint64         `gorm:"column:manifest_file_id;not null;index:idx_software_manifests_manifest_file_id"`
	ManifestFileVersionId uint64         `gorm:"column:manifest_file_version_id;type:bigint;not null;default:0"`
	VersionDescription    sql.NullString `gorm:"column:version_description;type:text"`
	CreatedBy             uint64         `gorm:"column:created_by;not null"`
	CreatedAt             time.Time      `gorm:"column:created_at;autoCreateTime"`
}

func (SoftwareManifestsTable) TableName() string { return "software_manifests" }

type BuildVersionsTable struct {
	Id                        uint64         `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectId                 uint64         `gorm:"column:project_id;not null;index:idx_build_versions_project_id"`
	SoftwareManifestId        uint64         `gorm:"column:software_manifest_id;not null;index:idx_build_versions_software_manifest_id"`
	Description               sql.NullString `gorm:"column:description;type:text"`
	BuildVersionFileId        uint64         `gorm:"column:build_version_file_id;type:bigint;not null;default:0"`
	BuildVersionFileVersionId uint64         `gorm:"column:build_version_file_version_id;type:bigint;not null;default:0"`
	CreatedAt                 time.Time      `gorm:"column:created_at;autoCreateTime;index:idx_build_versions_created_at"`
	CreatedBy                 uint64         `gorm:"column:created_by;not null"`
}

func (BuildVersionsTable) TableName() string { return "build_versions" }

type LlmProvidersTable struct {
	Id          uint64         `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_llm_providers_name"`
	BaseUrl     string         `gorm:"column:base_url;type:varchar(512);not null;default:''"`
	ApiKey      sql.NullString `gorm:"column:api_key;type:text"`
	Description sql.NullString `gorm:"column:description;type:text"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (LlmProvidersTable) TableName() string { return "llm_providers" }

type LlmModelsTable struct {
	Id               uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	ProviderId       uint64    `gorm:"column:provider_id;not null;uniqueIndex:uk_llm_models_provider_model,priority:1;index:idx_llm_models_provider_id"`
	ModelName        string    `gorm:"column:model_name;type:varchar(255);not null;uniqueIndex:uk_llm_models_provider_model,priority:2"`
	ModelType        string    `gorm:"column:model_type;type:enum('llm','vlm','embedding');not null;default:'llm';uniqueIndex:uk_llm_models_provider_model,priority:3"`
	MaxInputTokens   int       `gorm:"column:max_input_tokens;not null;default:0"`
	MaxOutputTokens  int       `gorm:"column:max_output_tokens;not null;default:0"`
	SupportStream    bool      `gorm:"column:support_stream;not null;default:false"`
	SupportJson      bool      `gorm:"column:support_json;not null;default:false"`
	PriceInputPer1k  float64   `gorm:"column:price_input_per_1k;type:decimal(10,6);not null;default:0"`
	PriceOutputPer1k float64   `gorm:"column:price_output_per_1k;type:decimal(10,6);not null;default:0"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (LlmModelsTable) TableName() string { return "llm_models" }

type LlmUsageLogsTable struct {
	Id           uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	LlmModelId   uint64    `gorm:"column:llm_model_id;not null;index:idx_llm_usage_logs_model_id"`
	ProjectId    uint64    `gorm:"column:project_id;not null;index:idx_llm_usage_logs_project_id"`
	InputTokens  int       `gorm:"column:input_tokens;not null;default:0"`
	OutputTokens int       `gorm:"column:output_tokens;not null;default:0"`
	CacheHit     bool      `gorm:"column:cache_hit;not null;default:false"`
	CostUsd      float64   `gorm:"column:cost_usd;type:decimal(10,6);not null;default:0"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;index:idx_llm_usage_logs_created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (LlmUsageLogsTable) TableName() string { return "llm_usage_logs" }

type AgentsTable struct {
	Id          uint64         `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_agents_name_type,priority:1"`
	Description sql.NullString `gorm:"column:description;type:text"`
	Instruction sql.NullString `gorm:"column:instruction;type:text"`
	AgentType   string         `gorm:"column:agent_type;type:enum('code','asset','design','test','build','ops','project');not null;default:'code';uniqueIndex:uk_agents_name_type,priority:2"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (AgentsTable) TableName() string { return "agents" }

type AgentLlmBindingsTable struct {
	Id         uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	AgentId    uint64    `gorm:"column:agent_id;not null;uniqueIndex:uk_agent_llm_bindings_agent_id"`
	LlmModelId uint64    `gorm:"column:llm_model_id;not null;index:idx_agent_llm_bindings_model_id"`
	Priority   int       `gorm:"column:priority;not null;default:0"`
	IsActive   bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AgentLlmBindingsTable) TableName() string { return "agent_llm_bindings" }

func cleanupDuplicateAgentBindings(db *gorm.DB) error {
	if !db.Migrator().HasTable("agent_llm_bindings") {
		return nil
	}

	type dupRow struct {
		AgentId uint64 `gorm:"column:agent_id"`
		Cnt     int64  `gorm:"column:cnt"`
	}

	var dups []dupRow
	if err := db.
		Table("agent_llm_bindings").
		Select("agent_id, COUNT(*) AS cnt").
		Group("agent_id").
		Having("COUNT(*) > 1").
		Scan(&dups).Error; err != nil {
		return err
	}

	for _, d := range dups {
		var rows []AgentLlmBindingsTable
		if err := db.
			Where("agent_id = ?", d.AgentId).
			Order("is_active DESC, priority DESC, id DESC").
			Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) <= 1 {
			continue
		}

		idsToDelete := make([]uint64, 0, len(rows)-1)
		for _, r := range rows[1:] {
			idsToDelete = append(idsToDelete, r.Id)
		}
		if len(idsToDelete) > 0 {
			if err := db.Where("id IN ?", idsToDelete).Delete(&AgentLlmBindingsTable{}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func enforceSingleBindingPerAgent(db *gorm.DB) error {
	if err := cleanupDuplicateAgentBindings(db); err != nil {
		return err
	}

	if db.Migrator().HasIndex(&AgentLlmBindingsTable{}, "uk_agent_llm_bindings_agent_model") {
		_ = db.Migrator().DropIndex(&AgentLlmBindingsTable{}, "uk_agent_llm_bindings_agent_model")
	}
	if db.Migrator().HasIndex(&AgentLlmBindingsTable{}, "idx_agent_llm_bindings_agent_id") {
		_ = db.Migrator().DropIndex(&AgentLlmBindingsTable{}, "idx_agent_llm_bindings_agent_id")
	}

	if !db.Migrator().HasIndex(&AgentLlmBindingsTable{}, "uk_agent_llm_bindings_agent_id") {
		if err := db.Migrator().CreateIndex(&AgentLlmBindingsTable{}, "uk_agent_llm_bindings_agent_id"); err != nil {
			return err
		}
	}
	return nil
}

func ensureAgentsInstructionColumn(db *gorm.DB) error {
	if !db.Migrator().HasTable("agents") {
		return nil
	}

	var hasInstruction int64
	if err := db.
		Raw("SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?", "agents", "instruction").
		Scan(&hasInstruction).Error; err != nil {
		return err
	}
	if hasInstruction > 0 {
		return nil
	}

	var hasCommand int64
	if err := db.
		Raw("SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?", "agents", "command").
		Scan(&hasCommand).Error; err != nil {
		return err
	}
	if hasCommand == 0 {
		return nil
	}

	return db.Exec("ALTER TABLE `agents` CHANGE COLUMN `command` `instruction` TEXT").Error
}

func ensureAgentsAgentTypeEnum(db *gorm.DB) error {
	if !db.Migrator().HasTable("agents") {
		return nil
	}

	var columnType string
	if err := db.
		Raw(
			"SELECT COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ? LIMIT 1",
			"agents",
			"agent_type",
		).
		Scan(&columnType).Error; err != nil {
		return err
	}

	if strings.Contains(columnType, "'project'") {
		return nil
	}

	return db.Exec(
		"ALTER TABLE `agents` MODIFY COLUMN `agent_type` ENUM('code','asset','design','test','build','ops','project') NOT NULL DEFAULT 'code'",
	).Error
}

func migrateWithRetry(targetDSN string) error {
	var lastErr error
	for attempt := 1; attempt <= 5; attempt++ {
		db, err := gorm.Open(mysql.Open(targetDSN), &gorm.Config{})
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.SetMaxOpenConns(10)
			sqlDB.SetMaxIdleConns(10)
			sqlDB.SetConnMaxLifetime(10 * time.Minute)
			sqlDB.SetConnMaxIdleTime(5 * time.Minute)
		}

		if err := cleanupDuplicateAgentBindings(db); err != nil {
			lastErr = err
		} else {
			if err := ensureAgentsInstructionColumn(db); err != nil {
				lastErr = err
				if sqlDB != nil {
					_ = sqlDB.Close()
				}
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			if err := ensureAgentsAgentTypeEnum(db); err != nil {
				lastErr = err
				if sqlDB != nil {
					_ = sqlDB.Close()
				}
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			err = db.AutoMigrate(
				&UsersTable{},
				&UserIdentitiesTable{},
				&ProjectsTable{},
				&ProjectMembersTable{},
				&FilesTable{},
				&ProjectFilesTable{},
				&FileVersionsTable{},
				&AdminsTable{},
				&SoftwareTemplatesTable{},
				&SoftwaresTable{},
				&SoftwareManifestsTable{},
				&BuildVersionsTable{},
				&LlmProvidersTable{},
				&LlmModelsTable{},
				&LlmUsageLogsTable{},
				&AgentsTable{},
				&AgentLlmBindingsTable{},
			)
			if err == nil {
				if err := enforceSingleBindingPerAgent(db); err == nil {
					requiredTables := []string{"softwares", "software_manifests", "build_versions"}
					for _, t := range requiredTables {
						if !db.Migrator().HasTable(t) {
							return fmt.Errorf("missing table: %s", t)
						}
					}
					return nil
				} else {
					lastErr = err
				}
			} else {
				lastErr = err
			}
		}

		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	return lastErr
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	targetDSN := c.MySQL.DSN
	if *dsn != "" {
		targetDSN = *dsn
	}
	if targetDSN == "" {
		log.Fatal("MySQL DSN is empty")
	}

	fmt.Println("Starting migration...")
	err := migrateWithRetry(targetDSN)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration completed successfully!")
}
