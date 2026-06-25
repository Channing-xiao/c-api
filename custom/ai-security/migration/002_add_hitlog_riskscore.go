package migration

import (
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	"gorm.io/gorm"
)

func init() {
	// AutoMigrate 仅补齐缺失列（risk_score），不会破坏已有数据；
	// 跨 SQLite/MySQL/PostgreSQL 均为 ADD COLUMN 语义。
	Register("002_add_hitlog_riskscore", func(db *gorm.DB) error {
		return db.AutoMigrate(&model.HitLog{})
	})
}
