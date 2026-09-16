package model

import "errors"

// VisualValidateSession 记录一次"真人审核"认证会话。
//
// 业务流（dev-docs/cii-api.md 真人审核）：
//  1. 调用方 POST /v1/visual-validate/sessions → new-api 调上游 sessions，
//     拿到 BytedToken + H5Link，并把上游 CallbackURL 指到 new-api 自己的
//     /v1/visual-validate/callback 端点（带 session public id + sign）。
//  2. 终端用户打开 H5Link 完成人脸认证，浏览器跳转 CallbackURL 并追加
//     bytedToken / resultCode 等 query 参数。
//  3. new-api 回调端点校验 sign 后落库结果；resultCode=10000 时顺势调
//     上游 results 换取 GroupId，并把该上游素材组自动注册为本地 AssetGroup。
//  4. 若调用方注册了下游 callback_url，best-effort POST 通知；调用方也可
//     直接轮询 GET /v1/visual-validate/sessions/:id。
//
// BytedToken 30 分钟有效、仅能认证一次，落库仅用于对账，不重复使用。
type VisualValidateSession struct {
	ID            uint   `gorm:"primaryKey" json:"-"`
	PublicID      string `gorm:"type:varchar(64);uniqueIndex" json:"id"` // 对外公开 ID（vv_<32位随机>）
	SignToken     string `gorm:"type:varchar(64)" json:"-"`              // 回调端点防伪造校验 token
	UserID        int    `gorm:"index"           json:"-"`
	ChannelID     int    `gorm:"index"           json:"-"`
	ChannelType   int    `gorm:"index"           json:"-"`
	UpstreamToken string `gorm:"type:varchar(256);index" json:"byted_token"` // 上游 BytedToken
	H5Link        string `gorm:"type:text"        json:"h5_link"`
	CallbackURL   string `gorm:"type:text"        json:"callback_url"` // new-api 生成的上游回调地址（含 session + sign）
	// DownstreamCallbackURL 是调用方自己注册的回调地址；认证完成后
	// best-effort POST 一次，失败不影响主流程。
	DownstreamCallbackURL string `gorm:"type:text"        json:"downstream_callback_url,omitempty"`
	// Status: pending（已创建未认证）/ succeeded / failed / expired
	Status string `gorm:"type:varchar(32);index" json:"status"`
	// ResultCode 保存回调里的 resultCode 原值（"10000"=通过）。
	ResultCode string `gorm:"type:varchar(32)" json:"result_code,omitempty"`
	// UpstreamGroupID 认证成功后上游 results 返回的素材组 ID。
	UpstreamGroupID string `gorm:"type:varchar(128)" json:"upstream_group_id,omitempty"`
	// LocalAssetGroupPublicID 上游 GroupId 对应的本地素材组公开 ID。
	LocalAssetGroupPublicID string `gorm:"type:varchar(64)" json:"local_asset_group_id,omitempty"`
	// NotifyStatus: 空串=无需通知；notified=已通知下游；failed=通知失败（可人工重推）。
	NotifyStatus string `gorm:"type:varchar(32)" json:"notify_status,omitempty"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// VisualValidate 会话状态常量。
const (
	VisualValidateStatusPending   = "pending"
	VisualValidateStatusSucceeded = "succeeded"
	VisualValidateStatusFailed    = "failed"
	VisualValidateStatusExpired   = "expired"

	VisualValidateNotifyNotified = "notified"
	VisualValidateNotifyFailed   = "failed"
)

// GenerateVisualValidateSessionID 生成对外公开的会话 ID（vv_<32位随机字符>）。
func GenerateVisualValidateSessionID() string {
	return "vv_" + randomIDKey()
}

// Insert 新建一条记录。
func (s *VisualValidateSession) Insert() error {
	return DB.Create(s).Error
}

// Update 全量保存（依赖主键）。
func (s *VisualValidateSession) Update() error {
	return DB.Save(s).Error
}

// GetVisualValidateSessionByPublicID 通过对外 ID 拿会话。
// ownerUserID<=0 表示不限制 owner（回调端点用）；否则只返回该用户的会话。
func GetVisualValidateSessionByPublicID(publicID string, ownerUserID int) (*VisualValidateSession, error) {
	if publicID == "" {
		return nil, ErrVvSessionNotFound
	}
	var s VisualValidateSession
	q := DB.Where("public_id = ?", publicID)
	if ownerUserID > 0 {
		q = q.Where("user_id = ?", ownerUserID)
	}
	if err := q.First(&s).Error; err != nil {
		return nil, ErrVvSessionNotFound
	}
	return &s, nil
}

// ListVisualValidateSessions 列出某个用户的认证会话，按 id 倒序。
// pageNum/pageSize 由调用方负责钳制。
func ListVisualValidateSessions(userID int, pageNum, pageSize int) ([]*VisualValidateSession, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var (
		items []*VisualValidateSession
		total int64
	)
	q := DB.Model(&VisualValidateSession{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id desc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ErrVvSessionNotFound 会话不存在的统一错误，便于调用方判断。
var ErrVvSessionNotFound = errors.New("visual validate session not found")
