package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
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
	ProjectId        uint64         `gorm:"column:project_id;not null;uniqueIndex:uk_project_name,priority:1"`
	Name             string         `gorm:"column:name;type:varchar(255);not null;uniqueIndex:uk_project_name,priority:2"`
	FileCategory     string         `gorm:"column:file_category;type:enum('text','image','video','audio','binary','archive');not null"`
	FileFormat       string         `gorm:"column:file_format;type:varchar(50);not null;default:''"`
	CurrentVersionId uint64         `gorm:"column:current_version_id;index:idx_files_current_version_id"`
	CreatedAt        time.Time      `gorm:"column:created_at;autoCreateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index;softDelete"`
}

func (FilesTable) TableName() string { return "files" }

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

	db, err := gorm.Open(mysql.Open(targetDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(
		&UsersTable{},
		&UserIdentitiesTable{},
		&ProjectsTable{},
		&ProjectMembersTable{},
		&FilesTable{},
		&FileVersionsTable{},
	)

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration completed successfully!")
}
