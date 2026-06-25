package migration

import (
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	"gorm.io/gorm"
)

func init() {
	Register("001_init_tables", func(db *gorm.DB) error {
		return db.AutoMigrate(
			&model.Config{},
			&model.Group{},
			&model.Rule{},
			&model.Word{},
			&model.Policy{},
			&model.HitLog{},
			&model.DailyStat{},
			&model.SyncState{},
			&model.AuditLog{},
		)
	})
}
