package model

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

// ModelTag 是模型广场标签的可维护字典。
// 字典只提供候选项与展示顺序；模型自身的 tags 字段仍是模型广场
// 标签筛选与卡片标签展示的数据源，二者通过标签文本关联。
type ModelTag struct {
	Id          int            `json:"id"`
	Name        string         `json:"name" gorm:"size:64;not null"`
	Sort        int            `json:"sort" gorm:"not null"`
	CreatedTime int64          `json:"created_time" gorm:"bigint"`
	UpdatedTime int64          `json:"updated_time" gorm:"bigint"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// ModelTagNameMaxRunes 与 Name 列的 varchar(64) 对齐，按字符数而不是字节数校验。
const ModelTagNameMaxRunes = 64

// IsValidModelTagName 校验标签能否安全地存入逗号分隔的 models.tags 字段。
func IsValidModelTagName(name string) bool {
	if name == "" || utf8.RuneCountInString(name) > ModelTagNameMaxRunes {
		return false
	}
	// 这些字符会被 models.tags 的解析逻辑当成分隔符或空格切分。
	return !strings.ContainsAny(name, ",;| \t\n\r")
}

// SplitModelTags 同时兼容管理端（逗号）与广场端（逗号/分号/竖线/空白）的解析。
func SplitModelTags(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', ';', '|', '、', ' ', '\t', '\n', '\r':
			return true
		default:
			return false
		}
	})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if tag := strings.TrimSpace(part); tag != "" {
			result = append(result, tag)
		}
	}
	return result
}

func GetAllModelTags() ([]ModelTag, error) {
	var tags []ModelTag
	err := DB.Model(&ModelTag{}).Order("sort ASC, id ASC").Find(&tags).Error
	return tags, err
}

// GetModelTagNames 供公开接口 /api/pricing 下发字典顺序；字典表尚未迁移时容错返回空。
func GetModelTagNames() []string {
	tags, err := GetAllModelTags()
	if err != nil {
		common.SysLog(fmt.Sprintf("load model tag dictionary failed: %v", err))
		return []string{}
	}
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	return names
}

func IsModelTagDuplicated(id int, name string) (bool, error) {
	if name == "" {
		return false, nil
	}
	var count int64
	err := DB.Model(&ModelTag{}).
		Where("id <> ? AND LOWER(name) = LOWER(?)", id, name).
		Count(&count).Error
	return count > 0, err
}

func InsertModelTag(name string, sort int) (*ModelTag, error) {
	now := common.GetTimestamp()
	tag := &ModelTag{
		Name:        name,
		Sort:        sort,
		CreatedTime: now,
		UpdatedTime: now,
	}
	if err := DB.Create(tag).Error; err != nil {
		return nil, err
	}
	return tag, nil
}

func UpdateModelTagById(id int, name string, sort int) error {
	return DB.Model(&ModelTag{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":         name,
		"sort":         sort,
		"updated_time": common.GetTimestamp(),
	}).Error
}

func DeleteModelTagById(id int) error {
	return DB.Delete(&ModelTag{}, id).Error
}

// ImportModelTagsFromModels 把历史模型上已有的标签补齐进字典，
// 便于升级后下拉框立即可用，不改动任何模型数据。
func ImportModelTagsFromModels() (int, error) {
	var models []Model
	if err := DB.Model(&Model{}).Select("tags").Find(&models).Error; err != nil {
		return 0, err
	}

	existing, err := GetAllModelTags()
	if err != nil {
		return 0, err
	}
	seen := make(map[string]struct{}, len(existing))
	maxSort := 0
	for _, tag := range existing {
		seen[strings.ToLower(tag.Name)] = struct{}{}
		if tag.Sort > maxSort {
			maxSort = tag.Sort
		}
	}

	added := 0
	for _, item := range models {
		for _, name := range SplitModelTags(item.Tags) {
			if !IsValidModelTagName(name) {
				continue
			}
			key := strings.ToLower(name)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			maxSort++
			if _, err := InsertModelTag(name, maxSort); err != nil {
				return added, err
			}
			added++
		}
	}
	return added, nil
}
