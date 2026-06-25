package model

import "github.com/QuantumNous/new-api/model"

// Word 敏感词（Keyword 规则的可选展开存储）
type Word struct {
	ID      int64  `gorm:"primaryKey;autoIncrement;column:id"`
	GroupID int64  `gorm:"column:group_id;type:bigint;not null;index:idx_aisec_word_group"`
	Word    string `gorm:"column:word;size:255;not null"`
	Type    int    `gorm:"column:type;type:int;default:1"`
	Status  int    `gorm:"column:status;type:int;default:1"`
	CreatedAt int64 `gorm:"column:created_at;type:bigint;default:0"`
}

func (Word) TableName() string { return "aisec_words" }

// ListWordsByGroupID 获取分组下的敏感词
func ListWordsByGroupID(groupID int64) ([]*Word, error) {
	var words []*Word
	err := model.DB.Where("group_id = ? AND status = ?", groupID, 1).Find(&words).Error
	return words, err
}
