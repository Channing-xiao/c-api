package model

import "github.com/QuantumNous/new-api/model"

// Group 规则分组
type Group struct {
	ID          int64  `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name        string `json:"name" gorm:"column:name;size:128;not null"`
	Description string `json:"description" gorm:"column:description;size:255;default:''"`
	ParentID    int64  `json:"parent_id" gorm:"column:parent_id;type:bigint;default:0;index:idx_aisec_group_parent"`
	Depth       int    `json:"depth" gorm:"column:depth;type:int;default:0"`
	Path        string `json:"path" gorm:"column:path;size:500;default:'';index:idx_aisec_group_path"`
	Status      int    `json:"status" gorm:"column:status;type:int;default:1;index:idx_aisec_group_status"`
	SortOrder   int    `json:"sort_order" gorm:"column:sort_order;type:int;default:0"`
	CreatedAt   int64  `json:"created_at" gorm:"column:created_at;type:bigint;default:0"`
	UpdatedAt   int64  `json:"updated_at" gorm:"column:updated_at;type:bigint;default:0"`
}

func (Group) TableName() string { return "aisec_groups" }

// GetGroupByID 根据 ID 获取分组
func GetGroupByID(id int64) (*Group, error) {
	var group Group
	err := model.DB.Where("id = ?", id).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// ListActiveGroups 获取所有启用分组
func ListActiveGroups() ([]*Group, error) {
	var groups []*Group
	err := model.DB.Where("status = ?", 1).Order("sort_order ASC, id ASC").Find(&groups).Error
	return groups, err
}
