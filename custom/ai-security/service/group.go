package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	newapimodel "github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

// GroupRequest 分组请求
type GroupRequest struct {
	Name        string `json:"name" binding:"required,max=128"`
	Description string `json:"description" binding:"max=255"`
	ParentID    int64  `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
	Status      int    `json:"status" binding:"oneof=0 1"`
}

// GetGroupByID 获取分组
func GetGroupByID(id int64) (*model.Group, error) {
	return model.GetGroupByID(id)
}

// ListGroups 获取分组列表
func ListGroups(page, pageSize, status int, parentID int64, name string) ([]*model.Group, int64, error) {
	var groups []*model.Group
	var total int64

	db := newapimodel.DB.Model(&model.Group{})
	if status >= 0 {
		db = db.Where("status = ?", status)
	}
	if parentID >= 0 {
		db = db.Where("parent_id = ?", parentID)
	}
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("sort_order ASC, id ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&groups).Error; err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// CreateGroup 创建分组
func CreateGroup(req *GroupRequest) (*model.Group, error) {
	depth := 0
	path := ""

	if req.ParentID > 0 {
		parent, err := model.GetGroupByID(req.ParentID)
		if err != nil {
			return nil, errors.New("父分组不存在")
		}
		maxDepth := model.GetIntConfig(constant.ConfigKeyMaxGroupDepth, constant.DefaultMaxGroupDepth)
		if parent.Depth >= maxDepth-1 {
			return nil, fmt.Errorf("分组嵌套深度不能超过%d层", maxDepth)
		}
		depth = parent.Depth + 1
		path = parent.Path
	}

	group := &model.Group{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		Depth:       depth,
		Path:        path,
		Status:      req.Status,
		SortOrder:   req.SortOrder,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	if err := newapimodel.DB.Create(group).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, errors.New("分组名称已存在")
		}
		return nil, err
	}

	if path == "" {
		group.Path = fmt.Sprintf("/%d", group.ID)
	} else {
		group.Path = fmt.Sprintf("%s/%d", path, group.ID)
	}
	_ = newapimodel.DB.Model(group).Update("path", group.Path)

	InvalidateRuleCache()
	return group, nil
}

// UpdateGroup 更新分组
func UpdateGroup(id int64, req *GroupRequest) error {
	group, err := model.GetGroupByID(id)
	if err != nil {
		return errors.New("分组不存在")
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"sort_order":  req.SortOrder,
		"status":      req.Status,
		"updated_at":  time.Now().Unix(),
	}

	if err := newapimodel.DB.Model(group).Updates(updates).Error; err != nil {
		return err
	}

	InvalidateRuleCache()
	return nil
}

// UpdateGroupStatus 更新分组状态
func UpdateGroupStatus(id int64, status int) error {
	group, err := model.GetGroupByID(id)
	if err != nil {
		return errors.New("分组不存在")
	}

	group.Status = status
	group.UpdatedAt = time.Now().Unix()
	if err := newapimodel.DB.Save(group).Error; err != nil {
		return err
	}

	InvalidateRuleCache()
	return nil
}

// DeleteGroup 删除分组
func DeleteGroup(id int64) error {
	group, err := model.GetGroupByID(id)
	if err != nil {
		return errors.New("分组不存在")
	}

	return newapimodel.DB.Transaction(func(tx *gorm.DB) error {
		// 删除子分组规则
		if err := tx.Where("group_id IN (?)", tx.Model(&model.Group{}).Select("id").Where("path LIKE ?", group.Path+"/%")).Delete(&model.Rule{}).Error; err != nil {
			return err
		}
		// 删除当前分组规则
		if err := tx.Where("group_id = ?", id).Delete(&model.Rule{}).Error; err != nil {
			return err
		}
		// 删除子分组
		if err := tx.Where("path LIKE ?", group.Path+"/%").Delete(&model.Group{}).Error; err != nil {
			return err
		}
		// 删除当前分组
		if err := tx.Delete(group).Error; err != nil {
			return err
		}
		return nil
	})
}

// CopyGroup 复制分组
func CopyGroup(id int64) (*model.Group, error) {
	src, err := model.GetGroupByID(id)
	if err != nil {
		return nil, errors.New("源分组不存在")
	}

	newGroup := &model.Group{
		Name:        src.Name + "(副本)",
		Description: src.Description,
		ParentID:    src.ParentID,
		Depth:       src.Depth,
		Path:        src.Path,
		Status:      constant.StatusEnabled,
		SortOrder:   src.SortOrder,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	if err := newapimodel.DB.Create(newGroup).Error; err != nil {
		return nil, err
	}

	if src.ParentID == 0 {
		newGroup.Path = fmt.Sprintf("/%d", newGroup.ID)
	} else {
		parentPath := strings.TrimSuffix(src.Path, fmt.Sprintf("/%d", src.ID))
		newGroup.Path = fmt.Sprintf("%s/%d", parentPath, newGroup.ID)
	}
	_ = newapimodel.DB.Model(newGroup).Update("path", newGroup.Path)

	// 复制规则
	var rules []*model.Rule
	newapimodel.DB.Where("group_id = ?", id).Find(&rules)
	for _, rule := range rules {
		newRule := &model.Rule{
			GroupID:     newGroup.ID,
			Name:        rule.Name,
			Type:        rule.Type,
			Content:     rule.Content,
			ExtraConfig: rule.ExtraConfig,
			Action:      rule.Action,
			Priority:    rule.Priority,
			RiskScore:   rule.RiskScore,
			Status:      rule.Status,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}
		newapimodel.DB.Create(newRule)
	}

	InvalidateRuleCache()
	return newGroup, nil
}

// GetGroupIDsWithChildren 获取分组及其所有子孙分组 ID
func GetGroupIDsWithChildren(groupID int64) ([]int64, error) {
	group, err := model.GetGroupByID(groupID)
	if err != nil {
		return nil, err
	}

	var ids []int64 = []int64{groupID}
	var children []*model.Group
	if err := newapimodel.DB.Where("path LIKE ?", group.Path+"/%").Find(&children).Error; err != nil {
		return nil, err
	}
	for _, child := range children {
		ids = append(ids, child.ID)
	}
	return ids, nil
}
