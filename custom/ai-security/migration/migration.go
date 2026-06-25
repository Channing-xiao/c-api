package migration

import (
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

// MigrationRecord 迁移记录
type MigrationRecord struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Version   string `gorm:"column:version;size:64;not null;uniqueIndex:idx_aisec_migration_version"`
	AppliedAt int64  `gorm:"column:applied_at;type:bigint;default:0"`
}

func (MigrationRecord) TableName() string { return "aisec_migrations" }

// MigrateFunc 迁移函数签名
type MigrateFunc func(db *gorm.DB) error

// Migration 迁移定义
type Migration struct {
	Version string
	Up      MigrateFunc
}

var migrations []Migration

// Register 注册迁移
func Register(version string, fn MigrateFunc) {
	migrations = append(migrations, Migration{Version: version, Up: fn})
}

// Run 执行所有未应用的迁移
func Run() error {
	// 确保迁移记录表存在
	if err := model.DB.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	for _, m := range migrations {
		applied, err := isApplied(m.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		if err := m.Up(model.DB); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.Version, err)
		}

		if err := markApplied(m.Version); err != nil {
			return fmt.Errorf("failed to mark migration %s applied: %w", m.Version, err)
		}
	}

	return nil
}

func isApplied(version string) (bool, error) {
	var count int64
	err := model.DB.Model(&MigrationRecord{}).Where("version = ?", version).Count(&count).Error
	return count > 0, err
}

func markApplied(version string) error {
	return model.DB.Create(&MigrationRecord{
		Version:   version,
		AppliedAt: time.Now().Unix(),
	}).Error
}
