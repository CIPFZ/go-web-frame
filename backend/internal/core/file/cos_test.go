package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	cos "github.com/tencentyun/cos-go-sdk-v5"

	"github.com/CIPFZ/gowebframe/internal/core/config"
)

var testCOS = config.COSConfig{SecretID: "id", SecretKey: "key", Bucket: "files-1250000000", Region: "ap-guangzhou", KeyPrefix: "uploads"}

func TestCOSURLs(t *testing.T) {
	d, err := NewCOSDriver(testCOS)
	require.NoError(t, err)
	require.Equal(t, "https://files-1250000000.cos.ap-guangzhou.myqcloud.com/reports/2026/Q1%20summary.pdf", d.PublicURL("reports/2026/Q1 summary.pdf"))
	require.Equal(t, "https://files-1250000000.cos-internal.ap-guangzhou.tencentcos.cn/exports/data.csv", d.InternalURL("/exports/data.csv"))
	require.Equal(t, "uploads/", d.keyPrefix)
}

func TestCOSUploadPolicyIsLimitedToOneKey(t *testing.T) {
	d, err := NewCOSDriver(testCOS)
	require.NoError(t, err)
	policy := d.uploadPolicy("exports/2026/data.csv")
	require.Len(t, policy.Statement, 1)
	require.Equal(t, []string{"qcs::cos:ap-guangzhou:uid/1250000000:files-1250000000/exports/2026/data.csv"}, policy.Statement[0].Resource)
	require.Contains(t, policy.Statement[0].Action, "name/cos:UploadPart")
	require.NotContains(t, policy.Statement[0].Action, "name/cos:DeleteObject")
}

func TestCOSRejectsWildcardCredentialKey(t *testing.T) {
	d, err := NewCOSDriver(testCOS)
	require.NoError(t, err)
	for _, key := range []string{"", "/", "exports/*", "a?b"} {
		_, err := d.IssueUploadCredential(context.Background(), key, time.Hour)
		require.Error(t, err, key)
	}
}

// TestCOSIntegration 访问真实 COS，需显式开启：
// FILE_COS_IT=1 FILE_COS_SECRET_ID=... FILE_COS_SECRET_KEY=... FILE_COS_BUCKET=... FILE_COS_REGION=...
func TestCOSIntegration(t *testing.T) {
	if os.Getenv("FILE_COS_IT") != "1" {
		t.Skip("set FILE_COS_IT=1 to run against real COS")
	}
	cfg := config.COSConfig{
		SecretID:  os.Getenv("FILE_COS_SECRET_ID"),
		SecretKey: os.Getenv("FILE_COS_SECRET_KEY"),
		Bucket:    os.Getenv("FILE_COS_BUCKET"),
		Region:    os.Getenv("FILE_COS_REGION"),
		KeyPrefix: "_it",
	}
	ctx := context.Background()
	d, err := NewCOSDriver(cfg)
	require.NoError(t, err)
	run := fmt.Sprintf("%d", time.Now().UnixNano())

	// 后端上传：上传 → 公网可读 → HEAD → 删除
	content := []byte("file-it-" + run)
	pubURL, key, err := d.Upload(ctx, multipartHeader(t, "it.txt", content), run+"/it.txt")
	require.NoError(t, err)
	require.Equal(t, "_it/"+run+"/it.txt", key)
	resp, err := http.Get(pubURL)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, content, body)
	info, err := d.Stat(ctx, key)
	require.NoError(t, err)
	require.EqualValues(t, len(content), info.SizeBytes)
	require.NoError(t, d.Delete(ctx, key))
	_, err = d.Stat(ctx, key)
	require.True(t, errors.Is(err, ErrObjectNotFound))

	// 临时凭据：只能写入签发的 key
	allowed := "_it/" + run + "/sts-allowed.bin"
	cred, err := d.IssueUploadCredential(ctx, allowed, 15*time.Minute)
	require.NoError(t, err)
	require.True(t, cred.ExpiresAt.After(time.Now()))
	tmp := tempClient(t, cfg, cred)
	_, err = tmp.Object.Put(ctx, allowed, bytes.NewReader([]byte("ok")), nil)
	require.NoError(t, err)
	_, err = tmp.Object.Put(ctx, "_it/"+run+"/sts-denied.bin", bytes.NewReader([]byte("no")), nil)
	require.Error(t, err, "credential must not write other keys")
	_, err = tmp.Object.Delete(ctx, allowed)
	require.Error(t, err, "credential must not delete")
	require.NoError(t, d.Delete(ctx, allowed))
}

func tempClient(t *testing.T, cfg config.COSConfig, cred UploadCredential) *cos.Client {
	t.Helper()
	u, err := url.Parse(publicBaseURL(cfg))
	require.NoError(t, err)
	return cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{Transport: &cos.AuthorizationTransport{
		SecretID: cred.TmpSecretID, SecretKey: cred.TmpSecretKey, SessionToken: cred.SessionToken,
	}})
}

func multipartHeader(t *testing.T, name string, content []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", name)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	_, header, err := req.FormFile("file")
	require.NoError(t, err)
	t.Cleanup(func() { _ = req.MultipartForm.RemoveAll() })
	require.True(t, strings.HasSuffix(header.Filename, name))
	return header
}
