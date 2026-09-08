package doubao

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeAssetApiPath(t *testing.T) {
	validAssetGroups := "/api/v1/asset-groups"
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty falls back to default", "", defaultAssetGroupsApiPath},
		{"whitespace falls back to default", "   ", defaultAssetGroupsApiPath},
		{"asset-groups passes through", validAssetGroups, validAssetGroups},
		{"assets passes through", "/api/v1/assets", "/api/v1/assets"},
		{"missing leading slash gets prepended", "api/v1/asset-groups", validAssetGroups},
		{"trims surrounding whitespace", "  " + validAssetGroups + "  ", validAssetGroups},
		{"dot segment falls back", "/api/../v1/asset-groups", defaultAssetGroupsApiPath},
		{"leading dot segment falls back", "../etc/passwd", defaultAssetGroupsApiPath},
		{"backslash falls back", "/api\\v1\\asset-groups", defaultAssetGroupsApiPath},
		{"query string falls back", validAssetGroups + "?page=1", defaultAssetGroupsApiPath},
		{"fragment falls back", validAssetGroups + "#frag", defaultAssetGroupsApiPath},
		{"scheme falls back", "https://evil.example.com" + validAssetGroups, defaultAssetGroupsApiPath},
		{"colon falls back", "/api:v1/asset-groups", defaultAssetGroupsApiPath},
		{"space inside falls back", "/api/ v1/asset-groups", defaultAssetGroupsApiPath},
		{"too long falls back", "/" + strings.Repeat("a", 257), defaultAssetGroupsApiPath},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, normalizeAssetApiPath(tc.in, defaultAssetGroupsApiPath))
		})
	}
}

func TestJoinAssetID(t *testing.T) {
	t.Run("rejects empty id", func(t *testing.T) {
		_, err := joinAssetID("/api/v1/asset-groups", "  ")
		assert.Error(t, err)
	})

	t.Run("rejects id with slash", func(t *testing.T) {
		_, err := joinAssetID("/api/v1/asset-groups", "abc/def")
		assert.Error(t, err)
	})

	t.Run("rejects id with query", func(t *testing.T) {
		_, err := joinAssetID("/api/v1/asset-groups", "abc?x=1")
		assert.Error(t, err)
	})

	t.Run("rejects too long id", func(t *testing.T) {
		_, err := joinAssetID("/api/v1/asset-groups", strings.Repeat("a", 129))
		assert.Error(t, err)
	})

	t.Run("accepts alphanumeric+.-_", func(t *testing.T) {
		out, err := joinAssetID("/api/v1/asset-groups", "abc.123-XYZ_foo")
		require.NoError(t, err)
		assert.Equal(t, "/api/v1/asset-groups/abc.123-XYZ_foo", out)
	})
}

func TestAssetDoJSON_PassesAuthAndBody(t *testing.T) {
	var (
		gotMethod string
		gotAuth   string
		gotBody   string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		gotBody = string(buf[:n])
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"AssetGroupId":"grp_abc"}`))
	}))
	defer srv.Close()

	a := &TaskAdaptor{}
	body, status, err := a.CreateGroup(srv.URL, "sk-test", "/api/v1/asset-groups", []byte(`{"Name":"x"}`), "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, `{"AssetGroupId":"grp_abc"}`, string(body))
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "Bearer sk-test", gotAuth)
	assert.Equal(t, `{"Name":"x"}`, gotBody)
}

func TestAssetDoJSON_PathInjectionFallsBack(t *testing.T) {
	// 即使 path 注入失败，CreateGroup 也应该落到默认 path，并发到 /api/v1/asset-groups。
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	a := &TaskAdaptor{}
	_, _, err := a.CreateGroup(srv.URL, "k", "../etc/passwd", []byte(`{}`), "")
	require.NoError(t, err)
	assert.Equal(t, "/api/v1/asset-groups", gotPath)
}

func TestAssetDoDelete_404IsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`not found`))
	}))
	defer srv.Close()

	a := &TaskAdaptor{}
	err := a.DeleteGroup(srv.URL, "k", "/api/v1/asset-groups", "grp_1", "")
	assert.NoError(t, err, "404 on delete is idempotent success")
}

func TestAssetDoDelete_5xxIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`upstream broke`))
	}))
	defer srv.Close()

	a := &TaskAdaptor{}
	err := a.DeleteGroup(srv.URL, "k", "/api/v1/asset-groups", "grp_1", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "502")
}
