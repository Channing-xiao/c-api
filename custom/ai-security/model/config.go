package model

import (
	"fmt"
	"strconv"
	"time"

	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

// Config 模块配置表
type Config struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;column:id"`
	ConfigKey   string `gorm:"column:config_key;size:128;not null;uniqueIndex:idx_aisec_config_key"`
	ConfigValue string `gorm:"column:config_value;type:text"`
	CreatedAt   int64  `gorm:"column:created_at;type:bigint;default:0"`
	UpdatedAt   int64  `gorm:"column:updated_at;type:bigint;default:0"`
}

func (Config) TableName() string { return "aisec_configs" }

// GetConfig 读取字符串配置
func GetConfig(key string) (string, error) {
	var config Config
	err := model.DB.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		return "", err
	}
	return config.ConfigValue, nil
}

// GetConfigWithDefault 读取配置，不存在返回默认值
func GetConfigWithDefault(key, defaultValue string) string {
	value, err := GetConfig(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetBoolConfig 读取布尔配置
func GetBoolConfig(key string, defaultValue bool) bool {
	value, err := GetConfig(key)
	if err != nil {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// GetIntConfig 读取整数配置
func GetIntConfig(key string, defaultValue int) int {
	value, err := GetConfig(key)
	if err != nil {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// GetStringConfig 读取字符串配置
func GetStringConfig(key, defaultValue string) string {
	return GetConfigWithDefault(key, defaultValue)
}

// SetConfig 设置配置
func SetConfig(key, value string) error {
	var config Config
	err := model.DB.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		// create
		config = Config{
			ConfigKey:   key,
			ConfigValue: value,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}
		return model.DB.Create(&config).Error
	}
	config.ConfigValue = value
	config.UpdatedAt = time.Now().Unix()
	return model.DB.Save(&config).Error
}

// IsEnabled 返回模块是否启用
func IsEnabled() bool {
	return GetBoolConfig(constant.ConfigKeyEnabled, true)
}

// InitDefaultConfigs 初始化默认配置
func InitDefaultConfigs() error {
	defaults := map[string]string{
		constant.ConfigKeyEnabled:           "true",
		constant.ConfigKeyAIModel:           constant.DefaultAIModel,
		constant.ConfigKeyAITimeout:         strconv.Itoa(constant.DefaultAITimeoutSeconds),
		constant.ConfigKeyLogRetention:      strconv.Itoa(constant.DefaultLogRetentionDays),
		constant.ConfigKeyAuditLogRetention: strconv.Itoa(constant.DefaultAuditRetentionDays),
		constant.ConfigKeyDefaultRiskScore:  strconv.Itoa(constant.DefaultRiskScore),
		constant.ConfigKeyMaxGroupDepth:     strconv.Itoa(constant.DefaultMaxGroupDepth),
		constant.ConfigKeyMaskStrategy:      "full",
		constant.ConfigKeyMaskPreserveChars: "0",
	}

	for key, value := range defaults {
		if _, err := GetConfig(key); err != nil {
			if err := SetConfig(key, value); err != nil {
				return fmt.Errorf("failed to set config %s: %w", key, err)
			}
		}
	}
	return nil
}

// ConfigMap 返回所有配置键值对
func ConfigMap() (map[string]string, error) {
	var configs []Config
	if err := model.DB.Find(&configs).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, c := range configs {
		result[c.ConfigKey] = c.ConfigValue
	}
	return result, nil
}

// UpdateConfigs 批量更新配置
func UpdateConfigs(updates map[string]string) error {
	return model.DB.Transaction(func(tx *gorm.DB) error {
		for key, value := range updates {
			var config Config
			err := tx.Where("config_key = ?", key).First(&config).Error
			if err != nil {
				config = Config{
					ConfigKey:   key,
					ConfigValue: value,
					CreatedAt:   time.Now().Unix(),
					UpdatedAt:   time.Now().Unix(),
				}
				if err := tx.Create(&config).Error; err != nil {
					return err
				}
			} else {
				config.ConfigValue = value
				config.UpdatedAt = time.Now().Unix()
				if err := tx.Save(&config).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
