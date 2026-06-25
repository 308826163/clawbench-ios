package model

import (
	"log/slog"
	"regexp"
	"strings"
	"sync"
)

// ───────────────────────────────────────────────────────────
// 模型 ID 映射表
// 用途：统一 ACP 和 CLI 模式的模型 ID 命名空间
// ───────────────────────────────────────────────────────────

// ModelIDMapping 静态映射表
// key: ACP 返回的原始 ID
// value: 统一 ID（用于数据库存储和前端显示）
var ModelIDMapping = map[string]string{
	// ── mimo 系列（ACP 返回的 ID）──
	// 注意：ACP 返回的 ID 是简短的别名，不是完整的模型名称
	"sonnet":            "mimo",
	"opus":              "mimo-pro",
	"haiku":             "mimo-lite",
	"default":           "mimo",

	// ── 完整模型名称映射（如果 ACP 返回完整名称）──
	"mimo-v2.5[1M]":     "mimo",
	"mimo-v2.5-pro[1M]": "mimo-pro",
	"mimo-v2.5":         "mimo",
	"mimo-v2.4[1M]":     "mimo",
	"mimo-v2.4":         "mimo",

	// ── claude 系列 ──
	"claude-sonnet-4-6": "claude-sonnet-4-6",
	"claude-opus-4-6":   "claude-opus-4-6",
	"claude-haiku-4-5":  "claude-haiku-4-5",

	// ── glm 系列 ──
	"glm-5.1":           "glm-5.1",
	"glm-4-flash":       "glm-4-flash",
	"glm-4-air":         "glm-4-air",

	// ── 其他模型 ──
	"gpt-4o":            "gpt-4o",
	"gpt-4-turbo":       "gpt-4-turbo",
}

// ───────────────────────────────────────────────────────────
// 未映射模型记录（用于自动检测）
// ───────────────────────────────────────────────────────────

var (
	// unmappedModels 存储未映射的模型 ID
	// key: "acpID|modelName"
	// value: true
	unmappedModels   = make(map[string]bool)
	unmappedModelsMu sync.RWMutex
)

// ───────────────────────────────────────────────────────────
// 核心函数
// ───────────────────────────────────────────────────────────

// GetUnifiedModelID 获取统一模型 ID
// 如果映射表中存在，返回统一 ID；否则返回原始 ID
func GetUnifiedModelID(acpID string) string {
	if acpID == "" {
		return ""
	}

	// 只对 ACP 模型进行映射，CLI 模型直接返回
	if !IsACPModel(acpID) {
		return acpID
	}

	// 精确匹配
	if unified, ok := ModelIDMapping[acpID]; ok {
		return unified
	}

	// 尝试忽略大小写匹配
	lowerID := strings.ToLower(acpID)
	for key, value := range ModelIDMapping {
		if strings.ToLower(key) == lowerID {
			return value
		}
	}

	// 未映射，返回原始 ID
	return acpID
}

// GetACPModelID 获取 ACP 原始 ID（反向查找）
// 如果映射表中存在，返回 ACP ID；否则返回统一 ID
func GetACPModelID(unifiedID string) string {
	if unifiedID == "" {
		return ""
	}

	// 精确匹配
	for acpID, unified := range ModelIDMapping {
		if unified == unifiedID {
			return acpID
		}
	}

	// 未映射，返回统一 ID
	return unifiedID
}

// IsModelMapped 检查模型 ID 是否已映射
func IsModelMapped(acpID string) bool {
	_, ok := ModelIDMapping[acpID]
	return ok
}

// IsACPModel 检查是否是 ACP 模型
// 通过检查 ID 格式来判断
func IsACPModel(modelID string) bool {
	if modelID == "" {
		return false
	}

	// 规则 1: 包含 [xxx] 后缀
	if strings.Contains(modelID, "[") {
		return true
	}

	// 规则 2: 包含版本号 (v1, v2.5, -beta, -latest)
	if matched, _ := regexp.MatchString(`-v\d+|-\w+-(beta|latest|rc)`, modelID); matched {
		return true
	}

	return false
}

// ───────────────────────────────────────────────────────────
// 自动检测逻辑
// ───────────────────────────────────────────────────────────

// RecordUnmappedModel 记录未映射的模型 ID
func RecordUnmappedModel(acpID string, modelName string) {
	unmappedModelsMu.Lock()
	defer unmappedModelsMu.Unlock()

	key := acpID + "|" + modelName
	if !unmappedModels[key] {
		unmappedModels[key] = true
		// 记录日志
		slog.Warn("model_mapping: unmapped model detected",
			"acp_id", acpID,
			"model_name", modelName,
			"action", "please add to ModelIDMapping in internal/model/model_mapping.go")
	}
}

// GetUnmappedModels 获取所有未映射的模型
func GetUnmappedModels() map[string]string {
	unmappedModelsMu.RLock()
	defer unmappedModelsMu.RUnlock()

	result := make(map[string]string)
	for key := range unmappedModels {
		parts := strings.SplitN(key, "|", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// ClearUnmappedModels 清空未映射记录（定期清理）
func ClearUnmappedModels() {
	unmappedModelsMu.Lock()
	defer unmappedModelsMu.Unlock()
	unmappedModels = make(map[string]bool)
}
