package model

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// AssetGroup 表示用户拥有的一个素材组（逻辑分组），
// 用于把若干 Asset（图片 / 视频）按业务维度组织起来再用于视频生成。
// 本表是 new-api 侧的"事实源"——元数据 + 拥有权都以本表为准；
// 上游 CII/Seedance 仓库里对应的资源是缓存映射，只在
// UpstreamAssetGroupID 字段里保存上游 ID，便于失败时人工对账。
type AssetGroup struct {
	ID                   uint   `gorm:"primaryKey" json:"-"`
	PublicID             string `gorm:"type:varchar(64);uniqueIndex" json:"id"` // 对外公开 ID（group_xxx）
	UserID               int    `gorm:"index"           json:"-"`               // 拥有者
	ChannelID            int    `gorm:"index"           json:"-"`               // 创建时选定的上游渠道
	ChannelType          int    `gorm:"index"           json:"-"`               // 渠道类型
	UpstreamAssetGroupID string `gorm:"type:varchar(128);index" json:"-"`       // 上游 ID，缓存/同步用
	Name                 string `gorm:"type:varchar(255)" json:"name"`
	Description          string `gorm:"type:text"        json:"description"`
	GroupType            string `gorm:"type:varchar(32)" json:"group_type,omitempty"` // 默认 "AIGC"
	CreatedAt            int64  `json:"created_at"`
	UpdatedAt            int64  `json:"updated_at"`
	DeletedAt            int64  `gorm:"index"            json:"-"` // 软删时间戳（0=未删）
}

// Asset 是用户上传并用于视频生成的图片/视频素材。
// 逻辑挂在 AssetGroup 下，group_id 是 AssetGroup.ID（自增主键，不是 PublicID）。
type Asset struct {
	ID               uint          `gorm:"primaryKey" json:"-"`
	PublicID         string        `gorm:"type:varchar(64);uniqueIndex" json:"id"`   // 对外公开 ID（asset_xxx）
	UserID           int           `gorm:"index"           json:"-"`                 // 拥有者
	ChannelID        int           `gorm:"index"           json:"-"`                 // 创建时选定的上游渠道
	ChannelType      int           `gorm:"index"           json:"-"`                 // 渠道类型
	GroupID          uint          `gorm:"index"           json:"-"`                 // 关联 asset_groups.id
	UpstreamAssetID  string        `gorm:"type:varchar(128);index" json:"-"`         // 上游 ID
	AssetType        string        `gorm:"type:varchar(32);index" json:"asset_type"` // Image / Video
	Name             string        `gorm:"type:varchar(255)" json:"name"`
	SourceURL        string        `gorm:"type:text"         json:"source_url"`  // 本地 OSS URL（=上游 ImageUrl）
	Status           string        `gorm:"type:varchar(32);index" json:"status"` // Active / Disabled
	SizeBytes        int64         `json:"size_bytes,omitempty"`
	MimeType         string        `gorm:"type:varchar(64)" json:"mime_type,omitempty"`
	UpstreamMappings AssetMappings `gorm:"type:json" json:"-"` // 预留：多 provider 时的上游 ID 映射
	CreatedAt        int64         `json:"created_at"`
	UpdatedAt        int64         `json:"updated_at"`
	DeletedAt        int64         `gorm:"index"            json:"-"` // 软删
}

// AssetMappings 保存多 provider 维度的上游资源 ID 映射。
// 初版只接 doubao/volcengine，Asset.UpstreamAssetID 已经够用；
// 这里留一个 JSON 字段以便未来扩展到更多 provider 时不破坏表结构。
type AssetMappings map[string]string

// Scan / Value 让 AssetMappings 可以当 GORM JSON 字段存到三库一致的 TEXT / JSON 字段。
func (a *AssetMappings) Scan(val interface{}) error {
	bytesValue, ok := val.([]byte)
	if !ok {
		return nil
	}
	if len(bytesValue) == 0 {
		return nil
	}
	return common.Unmarshal(bytesValue, a)
}

func (a AssetMappings) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	return common.Marshal(a)
}

// GenerateAssetGroupID 生成对外公开的素材组 ID（group_<32位随机字符>）。
// 与 task_xxx 同风格，便于从 ID 直观区分资源类型。
func GenerateAssetGroupID() string {
	return "group_" + randomIDKey()
}

// GenerateAssetID 生成对外公开的素材 ID（asset_<32位随机字符>）。
func GenerateAssetID() string {
	return "asset_" + randomIDKey()
}

func randomIDKey() string {
	key, err := common.GenerateRandomCharsKey(32)
	if err != nil {
		// 极端情况（熵源失败）用时间戳兜底；概率极低，避免阻塞主流程。
		return fmt.Sprintf("ts%d", time.Now().UnixNano())
	}
	return key
}

// Insert 新建一条记录。
func (g *AssetGroup) Insert() error {
	return DB.Create(g).Error
}

// Insert 新建一条记录。
func (a *Asset) Insert() error {
	return DB.Create(a).Error
}

// Update 全量保存（依赖主键）。
func (g *AssetGroup) Update() error {
	return DB.Save(g).Error
}

// Update 全量保存（依赖主键）。
func (a *Asset) Update() error {
	return DB.Save(a).Error
}

// GetAssetGroupByPublicID 通过对外 ID 拿素材组。
// ownerUserID<=0 表示不限制 owner；否则只返回属于该用户的组。
// 软删状态（DeletedAt>0）的组视作不存在。
func GetAssetGroupByPublicID(publicID string, ownerUserID int) (*AssetGroup, error) {
	if publicID == "" {
		return nil, errors.New("public id is required")
	}
	var g AssetGroup
	q := DB.Where("public_id = ?", publicID).Where("deleted_at = 0")
	if ownerUserID > 0 {
		q = q.Where("user_id = ?", ownerUserID)
	}
	if err := q.First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

// GetAssetByPublicID 通过对外 ID 拿素材。
// ownerUserID<=0 表示不限制 owner；否则只返回属于该用户的素材。
// 软删（DeletedAt>0）视为不存在。
func GetAssetByPublicID(publicID string, ownerUserID int) (*Asset, error) {
	if publicID == "" {
		return nil, errors.New("public id is required")
	}
	var a Asset
	q := DB.Where("public_id = ?", publicID).Where("deleted_at = 0")
	if ownerUserID > 0 {
		q = q.Where("user_id = ?", ownerUserID)
	}
	if err := q.First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// GetAssetGroupByUpstreamID 通过上游 ID 找素材组（不限制软删状态——
// 上游组只创建一次，重复注册前用它做幂等判断）。
// ownerUserID<=0 表示不限制 owner。
func GetAssetGroupByUpstreamID(upstreamID string, ownerUserID int) (*AssetGroup, error) {
	if upstreamID == "" {
		return nil, errors.New("upstream id is required")
	}
	var g AssetGroup
	q := DB.Where("upstream_asset_group_id = ?", upstreamID).Where("deleted_at = 0")
	if ownerUserID > 0 {
		q = q.Where("user_id = ?", ownerUserID)
	}
	if err := q.First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

// ListAssetGroups 列出某个用户下的素材组，按 id 倒序（最新在前）。
// pageNum/pageSize 由调用方负责钳制；本函数只做查询。
func ListAssetGroups(userID int, pageNum, pageSize int) ([]*AssetGroup, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var (
		items []*AssetGroup
		total int64
	)
	q := DB.Model(&AssetGroup{}).Where("user_id = ?", userID).Where("deleted_at = 0")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id desc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ListAssets 列出某个素材组下的素材，可选按 status 过滤。
// pageNum/pageSize 由调用方负责钳制。
func ListAssets(userID, groupID uint, statuses []string, pageNum, pageSize int) ([]*Asset, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var (
		items []*Asset
		total int64
	)
	q := DB.Model(&Asset{}).
		Where("user_id = ?", userID).
		Where("group_id = ?", groupID).
		Where("deleted_at = 0")
	if len(statuses) > 0 {
		q = q.Where("status in (?)", statuses)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id desc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// CountActiveAssetsInGroup 统计某个素材组下未删除的素材数量。
// 用于删素材组时校验"组内还有未删素材"——业务约束要求先清空再删组。
func CountActiveAssetsInGroup(userID, groupID uint) (int64, error) {
	var n int64
	err := DB.Model(&Asset{}).
		Where("user_id = ?", userID).
		Where("group_id = ?", groupID).
		Where("deleted_at = 0").
		Count(&n).Error
	return n, err
}

// SoftDeleteAssetGroup 把指定素材组软删（CAS：只删未删的）。
// 第二个返回值表示这次调用是否成功抢到删除（true=成功）。
// 已经被软删的素材组返回 (false, nil)，调用方应按幂等处理。
func SoftDeleteAssetGroup(userID int, publicID string) (bool, error) {
	res := DB.Model(&AssetGroup{}).
		Where("public_id = ?", publicID).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Updates(map[string]interface{}{
			"deleted_at": time.Now().Unix(),
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// SoftDeleteAsset 把指定素材软删（CAS：只删未删的）。
// 同 SoftDeleteAssetGroup 的语义，幂等。
func SoftDeleteAsset(userID int, publicID string) (bool, error) {
	res := DB.Model(&Asset{}).
		Where("public_id = ?", publicID).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Updates(map[string]interface{}{
			"deleted_at": time.Now().Unix(),
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
