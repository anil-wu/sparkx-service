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
	Id                   uint64         `db:"id" gorm:"column:id;primaryKey"`
	ProjectId             uint64        `db:"project_id" gorm:"column:project_id"`
	SoftwareId            uint64        `db:"software_id" gorm:"column:software_id"`
	ManifestFileId        uint64        `db:"manifest_file_id" gorm:"column:manifest_file_id"`
	ManifestFileVersionId uint64        `db:"manifest_file_version_id" gorm:"column:manifest_file_version_id"`
	VersionDescription    sql.NullString `db:"version_description" gorm:"column:version_description"`
	CreatedAt             time.Time     `db:"created_at" gorm:"column:created_at"`
	CreatedBy             uint64        `db:"created_by" gorm:"column:created_by"`
}

func (SoftwareManifests) TableName() string { return "software_manifests" }

