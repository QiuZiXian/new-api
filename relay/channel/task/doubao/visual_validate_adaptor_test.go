package doubao

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNormalizeVisualValidateApiPath_Fallback 校验 visual-validate 路径
// 为空/非法时回退到默认值；合法值原样透传（含自动补前导斜杠）。
func TestNormalizeVisualValidateApiPath_Fallback(t *testing.T) {
	assert.Equal(t, defaultVisualValidateApiPath, normalizeAssetApiPath("", defaultVisualValidateApiPath))
	assert.Equal(t, "/api/v1/visual-validate", normalizeAssetApiPath("/api/v1/visual-validate", defaultVisualValidateApiPath))
	assert.Equal(t, "/api/v1/visual-validate", normalizeAssetApiPath("api/v1/visual-validate", defaultVisualValidateApiPath))
	assert.Equal(t, defaultVisualValidateApiPath, normalizeAssetApiPath("../etc/passwd", defaultVisualValidateApiPath))
	assert.Equal(t, defaultVisualValidateApiPath, normalizeAssetApiPath("https://evil.example.com", defaultVisualValidateApiPath))
}

// TestCreateVisualValidateSession_RequestShape 校验拉起认证请求的
// method / path / header / body 都符合 dev-docs 的约定。
func TestCreateVisualValidateSession_RequestShape(t *testing.T) {
	var gotPath, gotMethod, gotAuth string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		gotBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"BytedToken":"tok","H5Link":"https://h5","CallbackURL":"https://cb"}`))
	}))
	defer srv.Close()

	a := &TaskAdaptor{}
	body, status, err := a.CreateVisualValidateSession(srv.URL, "sk-test", "", []byte(`{"CallbackURL":"https://cb"}`), "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "/api/v1/visual-validate/sessions", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "Bearer sk-test", gotAuth)
	assert.Equal(t, `{"CallbackURL":"https://cb"}`, string(gotBody))
	assert.Equal(t, `{"BytedToken":"tok","H5Link":"https://h5","CallbackURL":"https://cb"}`, string(body))
}

// TestGetVisualValidateResult_RequestShape 校验获取结果请求的 method / path。
func TestGetVisualValidateResult_RequestShape(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"GroupId":"group-20260331145705-abc"}`))
	}))
	defer srv.Close()

	a := &TaskAdaptor{}
	_, _, err := a.GetVisualValidateResult(srv.URL, "k", "", []byte(`{"BytedToken":"tok"}`), "")
	require.NoError(t, err)
	assert.Equal(t, "/api/v1/visual-validate/results", gotPath)
}
