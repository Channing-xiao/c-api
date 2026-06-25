package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/custom/ai-security/model"
)

const (
	cachePrefixRules       = "aisec:rules:group:%d"
	cachePrefixRulesGroups = "aisec:rules:groups:%s"
	cachePrefixPolicies    = "aisec:policies:user:%d"
	cachePrefixConfig      = "aisec:config:%s"
	cacheExpiration        = 5 * time.Minute
)

type cacheItem[T any] struct {
	data      T
	expiredAt int64
}

var (
	ruleCache    map[string]cacheItem[[]*model.Rule]
	policyCache  map[string]cacheItem[[]*model.Policy]
	configCache  map[string]cacheItem[string]
	cacheMutex   sync.RWMutex
)

// InitCache 初始化缓存
func InitCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	ruleCache = make(map[string]cacheItem[[]*model.Rule])
	policyCache = make(map[string]cacheItem[[]*model.Policy])
	configCache = make(map[string]cacheItem[string])
}

// GetCachedRules 从缓存获取规则
func GetCachedRules(groupID int64) ([]*model.Rule, bool) {
	key := fmt.Sprintf(cachePrefixRules, groupID)
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	item, ok := ruleCache[key]
	if !ok || item.expiredAt < time.Now().Unix() {
		return nil, false
	}
	return item.data, true
}

// SetCachedRules 设置规则缓存
func SetCachedRules(groupID int64, rules []*model.Rule) {
	key := fmt.Sprintf(cachePrefixRules, groupID)
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	ruleCache[key] = cacheItem[[]*model.Rule]{
		data:      rules,
		expiredAt: time.Now().Add(cacheExpiration).Unix(),
	}
}

// GetCachedRulesByGroups 从缓存获取多个分组的规则
func GetCachedRulesByGroups(groupIDs []int64) ([]*model.Rule, bool) {
	key := groupRulesCacheKey(groupIDs)
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	item, ok := ruleCache[key]
	if !ok || item.expiredAt < time.Now().Unix() {
		return nil, false
	}
	return item.data, true
}

// SetCachedRulesByGroups 设置多个分组规则缓存
func SetCachedRulesByGroups(groupIDs []int64, rules []*model.Rule) {
	key := groupRulesCacheKey(groupIDs)
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	ruleCache[key] = cacheItem[[]*model.Rule]{
		data:      rules,
		expiredAt: time.Now().Add(cacheExpiration).Unix(),
	}
}

func groupRulesCacheKey(groupIDs []int64) string {
	ids := make([]string, len(groupIDs))
	for i, id := range groupIDs {
		ids[i] = strconv.FormatInt(id, 10)
	}
	sort.Strings(ids)
	return fmt.Sprintf(cachePrefixRulesGroups, strings.Join(ids, ","))
}

// InvalidateRuleCache 失效规则缓存
func InvalidateRuleCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	ruleCache = make(map[string]cacheItem[[]*model.Rule])
}

// GetCachedPolicies 从缓存获取策略
func GetCachedPolicies(userID int) ([]*model.Policy, bool) {
	key := fmt.Sprintf(cachePrefixPolicies, userID)
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	item, ok := policyCache[key]
	if !ok || item.expiredAt < time.Now().Unix() {
		return nil, false
	}
	return item.data, true
}

// SetCachedPolicies 设置策略缓存
func SetCachedPolicies(userID int, policies []*model.Policy) {
	key := fmt.Sprintf(cachePrefixPolicies, userID)
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	policyCache[key] = cacheItem[[]*model.Policy]{
		data:      policies,
		expiredAt: time.Now().Add(cacheExpiration).Unix(),
	}
}

// InvalidatePolicyCache 失效策略缓存
func InvalidatePolicyCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	policyCache = make(map[string]cacheItem[[]*model.Policy])
}

// GetCachedConfig 从缓存获取配置
func GetCachedConfig(key string) (string, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	item, ok := configCache[key]
	if !ok || item.expiredAt < time.Now().Unix() {
		return "", false
	}
	return item.data, true
}

// SetCachedConfig 设置配置缓存
func SetCachedConfig(key, value string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	configCache[key] = cacheItem[string]{
		data:      value,
		expiredAt: time.Now().Add(cacheExpiration).Unix(),
	}
}

// InvalidateConfigCache 失效配置缓存
func InvalidateConfigCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	configCache = make(map[string]cacheItem[string])
}

// InvalidateAllCache 失效所有缓存
func InvalidateAllCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	ruleCache = make(map[string]cacheItem[[]*model.Rule])
	policyCache = make(map[string]cacheItem[[]*model.Policy])
	configCache = make(map[string]cacheItem[string])
}
