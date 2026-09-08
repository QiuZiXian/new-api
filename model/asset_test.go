package model

import (
	"crypto/rand"
	"encoding/binary"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// uniqUserID 给每个测试一个独立的 user_id，避免与同一包内其它 _test.go
// 共享的 in-memory sqlite 发生行污染。task_cas_test.go 的 TestMain 不清理
// asset_groups / assets 表，所以走 user_id 隔离比"测试间 truncate"更稳。
func uniqUserID(t *testing.T) int {
	t.Helper()
	var b [8]byte
	_, _ = rand.Read(b[:])
	return int(binary.BigEndian.Uint64(b[:])%1_000_000_000) + 1
}

// uniqSuffix 给每个测试一个独立的 public_id 后缀（base36 计数器），
// 进一步防 row pollution。
var suffixCounter uint64

func uniqSuffix() string {
	return uintToBase36(atomic.AddUint64(&suffixCounter, 1))
}

func makeGroup(t *testing.T, userID, channelID, channelType int, publicID, upstreamID string) *AssetGroup {
	t.Helper()
	now := time.Now().Unix()
	g := &AssetGroup{
		PublicID:             publicID,
		UserID:               userID,
		ChannelID:            channelID,
		ChannelType:          channelType,
		UpstreamAssetGroupID: upstreamID,
		Name:                 "g-" + publicID,
		GroupType:            "AIGC",
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	require.NoError(t, g.Insert())
	return g
}

func makeAsset(t *testing.T, userID, channelID, channelType int, groupID uint, publicID, upstreamID string) *Asset {
	t.Helper()
	now := time.Now().Unix()
	a := &Asset{
		PublicID:        publicID,
		UserID:          userID,
		ChannelID:       channelID,
		ChannelType:     channelType,
		GroupID:         groupID,
		UpstreamAssetID: upstreamID,
		AssetType:       "Image",
		Name:            "a-" + publicID,
		SourceURL:       "https://example.com/" + publicID,
		Status:          "Active",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	require.NoError(t, a.Insert())
	return a
}

func TestGenerateAssetGroupID_Format(t *testing.T) {
	id := GenerateAssetGroupID()
	require.True(t, strings.HasPrefix(id, "group_"), "PublicID must start with group_")
	assert.Equal(t, 6+32, len(id), "group_<32-char>")
}

func TestGenerateAssetID_Format(t *testing.T) {
	id := GenerateAssetID()
	require.True(t, strings.HasPrefix(id, "asset_"), "PublicID must start with asset_")
	assert.Equal(t, 6+32, len(id), "asset_<32-char>")
}

func TestGetAssetGroupByPublicID_OwnerFilter(t *testing.T) {
	u1 := uniqUserID(t)
	u2 := uniqUserID(t)
	suf := uniqSuffix()
	pid1 := "group_owner1_" + suf
	pid2 := "group_owner2_" + suf
	g1 := makeGroup(t, u1, 10, 54, pid1, "upstream-1")
	_ = makeGroup(t, u2, 10, 54, pid2, "upstream-2")

	// owner=u1 能拿到自己的
	found, err := GetAssetGroupByPublicID(pid1, u1)
	require.NoError(t, err)
	assert.Equal(t, g1.PublicID, found.PublicID)

	// owner=u1 拿不到 u2 的
	_, err = GetAssetGroupByPublicID(pid2, u1)
	assert.Error(t, err)

	// owner=u2 能拿到自己的
	found, err = GetAssetGroupByPublicID(pid2, u2)
	require.NoError(t, err)
	assert.Equal(t, pid2, found.PublicID)
}

func TestSoftDeleteAssetGroup_IsIdempotent(t *testing.T) {
	u := uniqUserID(t)
	pid := "group_del_" + uniqSuffix()
	makeGroup(t, u, 10, 54, pid, "upstream-del")

	won, err := SoftDeleteAssetGroup(u, pid)
	require.NoError(t, err)
	assert.True(t, won, "first delete should win")

	won, err = SoftDeleteAssetGroup(u, pid)
	require.NoError(t, err)
	assert.False(t, won, "second delete is no-op")

	_, err = GetAssetGroupByPublicID(pid, u)
	assert.Error(t, err)
}

func TestCountActiveAssetsInGroup_RespectsSoftDelete(t *testing.T) {
	u := uniqUserID(t)
	suf := uniqSuffix()
	g := makeGroup(t, u, 10, 54, "group_cnt_"+suf, "upstream-cnt")
	a1 := makeAsset(t, u, 10, 54, g.ID, "asset_cnt_1_"+suf, "up-a-1")
	_ = makeAsset(t, u, 10, 54, g.ID, "asset_cnt_2_"+suf, "up-a-2")

	n, err := CountActiveAssetsInGroup(uint(u), g.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)

	won, err := SoftDeleteAsset(u, a1.PublicID)
	require.NoError(t, err)
	assert.True(t, won)

	n, err = CountActiveAssetsInGroup(uint(u), g.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestListAssets_StatusFilter(t *testing.T) {
	u := uniqUserID(t)
	suf := uniqSuffix()
	g := makeGroup(t, u, 10, 54, "group_lst_"+suf, "upstream-lst")
	makeAsset(t, u, 10, 54, g.ID, "asset_lst_1_"+suf, "up-a-1")
	a2 := makeAsset(t, u, 10, 54, g.ID, "asset_lst_2_"+suf, "up-a-2")

	items, total, err := ListAssets(uint(u), g.ID, nil, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)

	a2.Status = "Disabled"
	a2.UpdatedAt = time.Now().Unix()
	require.NoError(t, a2.Update())

	items, total, err = ListAssets(uint(u), g.ID, []string{"Active"}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, items, 1)
	assert.Equal(t, "Active", items[0].Status)
}

func TestListAssetGroups_OrderedAndPaged(t *testing.T) {
	u := uniqUserID(t)
	suf := uniqSuffix()
	makeGroup(t, u, 10, 54, "group_p_1_"+suf, "up-p-1")
	makeGroup(t, u, 10, 54, "group_p_2_"+suf, "up-p-2")
	makeGroup(t, u, 10, 54, "group_p_3_"+suf, "up-p-3")

	items, total, err := ListAssetGroups(u, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, items, 2, "page size=2")
	for i := 1; i < len(items); i++ {
		assert.Greater(t, items[i-1].ID, items[i].ID)
	}
}

func TestAssetMappings_JSONRoundTrip(t *testing.T) {
	m := AssetMappings{
		"doubao": "up-1",
		"volc":   "up-2",
	}
	val, err := m.Value()
	require.NoError(t, err)
	require.NotNil(t, val)

	var out AssetMappings
	require.NoError(t, out.Scan(val))
	assert.Equal(t, m, out)

	empty := AssetMappings{}
	val, err = empty.Value()
	require.NoError(t, err)
	assert.Nil(t, val, "empty mappings serialize to NULL so GORM does not write")
}

func uintToBase36(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [16]byte
	i := len(buf)
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	for n > 0 {
		i--
		buf[i] = alphabet[n%36]
		n /= 36
	}
	return string(buf[i:])
}
