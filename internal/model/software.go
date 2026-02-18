package model

import (
	"database/sql"
	"time"
)

type Softwares struct {
	Id              uint64         `db:"id" gorm:"column:id;primaryKey"`
	ProjectId       uint64         `db:"project_id" gorm:"column:project_id"`
	Name            string         `db:"name" gorm:"column:name"`
	Description     sql.NullString `db:"description" gorm:"column:description"`
	TemplateId      uint64         `db:"template_id" gorm:"column:template_id"`
	TechnologyStack string         `db:"technology_stack" gorm:"column:technology_stack"`
	Status          string         `db:"status" gorm:"column:status"`
	CreatedBy       uint64         `db:"created_by" gorm:"column:created_by"`
	CreatedAt       time.Time      `db:"created_at" gorm:"column:created_at"`
	UpdatedAt       time.Time      `db:"updated_at" gorm:"column:updated_at"`
}

func (Softwares) TableName() string { return "softwares" }

type SoftwareManifests struct {
	Id                    uint64         `db:"id" gorm:"column:id;primaryKey"`
	ProjectId             uint64         `db:"project_id" gorm:"column:project_id"`
	SoftwareId            uint64         `db:"software_id" gorm:"column:software_id"`
	ManifestFileId        uint64         `db:"manifest_file_id" gorm:"column:manifest_file_id"`
	ManifestFileVersionId uint64         `db:"manifest_file_version_id" gorm:"column:manifest_file_version_id"`
	VersionNumber         uint32         `db:"version_number" gorm:"column:version_number"`
	VersionDescription    sql.NullString `db:"version_description" gorm:"column:version_description"`
	CreatedAt             time.Time      `db:"created_at" gorm:"column:created_at"`
	CreatedBy             uint64         `db:"created_by" gorm:"column:created_by"`
}

func (SoftwareManifests) TableName() string { return "software_manifests" }

type BuildVersions struct {
	Id                        uint64         `db:"id" gorm:"column:id;primaryKey"`
	ProjectId                 uint64         `db:"project_id" gorm:"column:project_id"`
	SoftwareManifestId        uint64         `db:"software_manifest_id" gorm:"column:software_manifest_id"`
	VersionNumber             uint32         `db:"version_number" gorm:"column:version_number"`
	Description               sql.NullString `db:"description" gorm:"column:description"`
	BuildVersionFileId        uint64         `db:"build_version_file_id" gorm:"column:build_version_file_id"`
	BuildVersionFileVersionId uint64         `db:"build_version_file_version_id" gorm:"column:build_version_file_version_id"`
	CreatedAt                 time.Time      `db:"created_at" gorm:"column:created_at"`
	CreatedBy                 uint64         `db:"created_by" gorm:"column:created_by"`
}

func (BuildVersions) TableName() string { return "build_versions" }

type Releases struct {
	Id                           uint64         `db:"id" gorm:"column:id;primaryKey"`
	ProjectId                    uint64         `db:"project_id" gorm:"column:project_id"`
	BuildVersionId               uint64         `db:"build_version_id" gorm:"column:build_version_id"`
	ReleaseManifestFileId        uint64         `db:"release_manifest_file_id" gorm:"column:release_manifest_file_id"`
	ReleaseManifestFileVersionId uint64         `db:"release_manifest_file_version_id" gorm:"column:release_manifest_file_version_id"`
	Name                         string         `db:"name" gorm:"column:name"`
	Channel                      string         `db:"channel" gorm:"column:channel"`
	Platform                     string         `db:"platform" gorm:"column:platform"`
	Status                       string         `db:"status" gorm:"column:status"`
	VersionTag                   sql.NullString `db:"version_tag" gorm:"column:version_tag"`
	ReleaseNotes                 sql.NullString `db:"release_notes" gorm:"column:release_notes"`
	CreatedAt                    time.Time      `db:"created_at" gorm:"column:created_at"`
	PublishedAt                  sql.NullTime   `db:"published_at" gorm:"column:published_at"`
	CreatedBy                    uint64         `db:"created_by" gorm:"column:created_by"`
}

func (Releases) TableName() string { return "releases" }
