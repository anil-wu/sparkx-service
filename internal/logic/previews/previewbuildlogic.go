package previews

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type PreviewBuildLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPreviewBuildLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreviewBuildLogic {
	return &PreviewBuildLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type buildVersionFileEntry struct {
	Path          string `json:"path"`
	FileId        int64  `json:"fileId"`
	VersionId     int64  `json:"versionId"`
	VersionNumber int64  `json:"versionNumber"`
}

type buildVersionJson struct {
	Entry string                  `json:"entry"`
	Files []buildVersionFileEntry `json:"files"`
}

func normalizeBuildPath(raw string) string {
	p := strings.TrimSpace(raw)
	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimLeft(p, "/")
	return p
}

func injectBaseHref(html string, baseHref string) string {
	lower := strings.ToLower(html)
	if strings.Contains(lower, "<base") {
		return html
	}
	headIdx := strings.Index(lower, "<head")
	if headIdx < 0 {
		return html
	}
	gtIdx := strings.Index(lower[headIdx:], ">")
	if gtIdx < 0 {
		return html
	}
	insertAt := headIdx + gtIdx + 1
	return html[:insertAt] + `<base href="` + baseHref + `">` + html[insertAt:]
}

func (l *PreviewBuildLogic) requireBuildVersionAccess(buildVersionId int64) (*model.BuildVersions, error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	var bv model.BuildVersions
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", buildVersionId).First(&bv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ProjectMembers{}).
		Where("project_id = ? AND user_id = ?", bv.ProjectId, userId).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("project not found or permission denied")
	}
	return &bv, nil
}

func (l *PreviewBuildLogic) readObject(storageKey string) ([]byte, error) {
	if l.svcCtx.OSSBucket == nil {
		return nil, errors.New("OSS not configured")
	}
	reader, err := l.svcCtx.OSSBucket.GetObject(storageKey)
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()
	return io.ReadAll(reader)
}

func (l *PreviewBuildLogic) signGetURL(storageKey string) (string, error) {
	if l.svcCtx.OSSBucket == nil {
		return "", errors.New("OSS not configured")
	}
	return l.svcCtx.OSSBucket.SignURL(storageKey, "GET", int64(l.svcCtx.Config.OSS.ExpireSeconds))
}

func (l *PreviewBuildLogic) loadBuildVersionManifest(bv *model.BuildVersions) (*buildVersionJson, error) {
	if bv.BuildVersionFileId <= 0 || bv.BuildVersionFileVersionId <= 0 {
		return nil, model.ErrNotFound
	}

	var version model.FileVersions
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", bv.BuildVersionFileVersionId).First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	raw, err := l.readObject(version.StorageKey)
	if err != nil {
		return nil, err
	}

	var manifest buildVersionJson
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func (l *PreviewBuildLogic) findManifestEntry(manifest *buildVersionJson, requestedRelPath string) *buildVersionFileEntry {
	if manifest == nil {
		return nil
	}
	normalized := normalizeBuildPath(requestedRelPath)
	if normalized == "" {
		return nil
	}
	for i := range manifest.Files {
		if normalizeBuildPath(manifest.Files[i].Path) == normalized {
			return &manifest.Files[i]
		}
	}
	for i := range manifest.Files {
		p := normalizeBuildPath(manifest.Files[i].Path)
		if p == normalized {
			return &manifest.Files[i]
		}
		if strings.HasSuffix(p, "/"+normalized) || strings.HasSuffix(p, normalized) {
			return &manifest.Files[i]
		}
	}
	return nil
}

func (l *PreviewBuildLogic) resolveFileVersionStorageKey(fileId int64, versionId int64, versionNumber int64) (string, error) {
	if fileId <= 0 {
		return "", model.InputParamInvalid
	}
	var fv model.FileVersions
	if versionId > 0 {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", versionId).First(&fv).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", model.ErrNotFound
			}
			return "", err
		}
		return fv.StorageKey, nil
	}

	if versionNumber > 0 {
		if err := l.svcCtx.DB.WithContext(l.ctx).
			Where("file_id = ? AND version_number = ?", fileId, versionNumber).
			First(&fv).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", model.ErrNotFound
			}
			return "", err
		}
		return fv.StorageKey, nil
	}

	var file model.Files
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", fileId).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", model.ErrNotFound
		}
		return "", err
	}
	if file.CurrentVersionId == 0 {
		return "", model.ErrNotFound
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", file.CurrentVersionId).First(&fv).Error; err != nil {
		return "", err
	}
	return fv.StorageKey, nil
}

func (l *PreviewBuildLogic) GetEntryHTML(buildVersionId int64) (string, error) {
	bv, err := l.requireBuildVersionAccess(buildVersionId)
	if err != nil {
		return "", err
	}

	baseHref := fmt.Sprintf("/api/v1/previews/builds/%d/", buildVersionId)
	prefix := strings.TrimSpace(bv.PreviewStoragePrefix)
	if prefix != "" {
		entryPath := normalizeBuildPath(bv.EntryPath)
		if entryPath == "" {
			entryPath = "index.html"
		}
		storageKey := prefix + entryPath
		raw, err := l.readObject(storageKey)
		if err != nil {
			return "", err
		}
		return injectBaseHref(string(raw), baseHref), nil
	}

	manifest, err := l.loadBuildVersionManifest(bv)
	if err != nil {
		return "", err
	}

	entryPath := normalizeBuildPath(manifest.Entry)
	if entryPath == "" {
		entryPath = "index.html"
	}

	entry := l.findManifestEntry(manifest, entryPath)
	if entry == nil {
		if l.svcCtx.OSSBucket != nil {
			prefix, err := l.ensurePreviewStoragePrefix(bv)
			if err != nil {
				return "", err
			}
			raw, err := l.readObject(prefix + entryPath)
			if err != nil {
				return "", err
			}
			return injectBaseHref(string(raw), baseHref), nil
		}
		return "", model.ErrNotFound
	}

	if entry.FileId <= 0 && l.svcCtx.OSSBucket != nil {
		prefix, err := l.ensurePreviewStoragePrefix(bv)
		if err != nil {
			return "", err
		}
		raw, err := l.readObject(prefix + entryPath)
		if err != nil {
			return "", err
		}
		return injectBaseHref(string(raw), baseHref), nil
	}

	storageKey, err := l.resolveFileVersionStorageKey(entry.FileId, entry.VersionId, entry.VersionNumber)
	if err != nil {
		return "", err
	}

	raw, err := l.readObject(storageKey)
	if err != nil {
		return "", err
	}
	return injectBaseHref(string(raw), baseHref), nil
}

func (l *PreviewBuildLogic) GetAssetRedirectURL(buildVersionId int64, requestedPath string) (string, error) {
	bv, err := l.requireBuildVersionAccess(buildVersionId)
	if err != nil {
		return "", err
	}

	reqPath := normalizeBuildPath(requestedPath)
	if reqPath == "" {
		return "", model.InputParamInvalid
	}

	prefix := strings.TrimSpace(bv.PreviewStoragePrefix)
	if prefix != "" {
		url, err := l.signGetURL(prefix + reqPath)
		if err != nil {
			return "", err
		}
		return url, nil
	}

	manifest, err := l.loadBuildVersionManifest(bv)
	if err != nil {
		if l.svcCtx.OSSBucket != nil {
			prefix, err := l.ensurePreviewStoragePrefix(bv)
			if err != nil {
				return "", err
			}
			return l.signGetURL(prefix + reqPath)
		}
		return "", err
	}

	entry := l.findManifestEntry(manifest, reqPath)
	if entry == nil {
		if l.svcCtx.OSSBucket != nil {
			prefix, err := l.ensurePreviewStoragePrefix(bv)
			if err != nil {
				return "", err
			}
			return l.signGetURL(prefix + reqPath)
		}
		return "", model.ErrNotFound
	}

	if entry.FileId <= 0 && l.svcCtx.OSSBucket != nil {
		prefix, err := l.ensurePreviewStoragePrefix(bv)
		if err != nil {
			return "", err
		}
		return l.signGetURL(prefix + reqPath)
	}

	storageKey, err := l.resolveFileVersionStorageKey(entry.FileId, entry.VersionId, entry.VersionNumber)
	if err != nil {
		if l.svcCtx.OSSBucket != nil {
			prefix, err := l.ensurePreviewStoragePrefix(bv)
			if err != nil {
				return "", err
			}
			return l.signGetURL(prefix + reqPath)
		}
		return "", err
	}
	return l.signGetURL(storageKey)
}

func (l *PreviewBuildLogic) ResolveAssetStorageKey(buildVersionId int64, requestedPath string) (string, error) {
	bv, err := l.requireBuildVersionAccess(buildVersionId)
	if err != nil {
		return "", err
	}

	reqPath := normalizeBuildPath(requestedPath)
	if reqPath == "" {
		return "", model.InputParamInvalid
	}

	prefix := strings.TrimSpace(bv.PreviewStoragePrefix)
	if prefix != "" {
		return prefix + reqPath, nil
	}

	manifest, err := l.loadBuildVersionManifest(bv)
	if err != nil {
		if l.svcCtx.OSSBucket != nil {
			prefix, err := l.ensurePreviewStoragePrefix(bv)
			if err != nil {
				return "", err
			}
			return prefix + reqPath, nil
		}
		return "", err
	}

	entry := l.findManifestEntry(manifest, reqPath)
	if entry == nil {
		if l.svcCtx.OSSBucket != nil {
			prefix, err := l.ensurePreviewStoragePrefix(bv)
			if err != nil {
				return "", err
			}
			return prefix + reqPath, nil
		}
		return "", model.ErrNotFound
	}

	if entry.FileId <= 0 && l.svcCtx.OSSBucket != nil {
		prefix, err := l.ensurePreviewStoragePrefix(bv)
		if err != nil {
			return "", err
		}
		return prefix + reqPath, nil
	}

	storageKey, err := l.resolveFileVersionStorageKey(entry.FileId, entry.VersionId, entry.VersionNumber)
	if err != nil {
		if l.svcCtx.OSSBucket != nil {
			prefix, err := l.ensurePreviewStoragePrefix(bv)
			if err != nil {
				return "", err
			}
			return prefix + reqPath, nil
		}
		return "", err
	}
	return storageKey, nil
}

func (l *PreviewBuildLogic) OpenAssetObject(buildVersionId int64, requestedPath string) (io.ReadCloser, string, int64, error) {
	if l.svcCtx.OSSBucket == nil {
		return nil, "", 0, errors.New("OSS not configured")
	}

	reqPath := normalizeBuildPath(requestedPath)
	if reqPath == "" {
		return nil, "", 0, model.InputParamInvalid
	}

	storageKey, err := l.ResolveAssetStorageKey(buildVersionId, reqPath)
	if err != nil {
		return nil, "", 0, err
	}

	var contentType string
	var contentLength int64

	if meta, err := l.svcCtx.OSSBucket.GetObjectMeta(storageKey); err == nil && meta != nil {
		contentType = strings.TrimSpace(meta.Get("Content-Type"))
		if rawLen := strings.TrimSpace(meta.Get("Content-Length")); rawLen != "" {
			if n, parseErr := strconv.ParseInt(rawLen, 10, 64); parseErr == nil {
				contentLength = n
			}
		}
	}

	if contentType == "" {
		if idx := strings.LastIndex(reqPath, "."); idx >= 0 && idx+1 < len(reqPath) {
			contentType = getContentTypeByFormat(reqPath[idx+1:])
		}
	}

	reader, err := l.svcCtx.OSSBucket.GetObject(storageKey)
	if err != nil {
		return nil, "", 0, err
	}
	return reader, contentType, contentLength, nil
}

func (l *PreviewBuildLogic) ensurePreviewStoragePrefix(bv *model.BuildVersions) (string, error) {
	prefix := strings.TrimSpace(bv.PreviewStoragePrefix)
	if prefix != "" {
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		return prefix, nil
	}

	softwareId := bv.SoftwareManifestId
	prefix = fmt.Sprintf("previews/%d/%d/%d/", bv.ProjectId, softwareId, bv.Id)
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.BuildVersions{}).
		Where("id = ?", bv.Id).
		Update("preview_storage_prefix", prefix).Error; err != nil {
		bv.PreviewStoragePrefix = prefix
		return prefix, nil
	}
	bv.PreviewStoragePrefix = prefix
	return prefix, nil
}

func (l *PreviewBuildLogic) PreuploadBuildFile(buildVersionId int64, relativePath string, fileFormat string, sizeBytes int64, hash string, contentType string) (string, string, string, error) {
	if sizeBytes <= 0 {
		return "", "", "", errors.New("sizeBytes is required")
	}
	if strings.TrimSpace(hash) == "" {
		return "", "", "", errors.New("hash is required")
	}
	p, err := normalizeRelativePathForPreview(relativePath)
	if err != nil {
		return "", "", "", err
	}

	bv, err := l.requireBuildVersionAccess(buildVersionId)
	if err != nil {
		return "", "", "", err
	}

	prefix, err := l.ensurePreviewStoragePrefix(bv)
	if err != nil {
		return "", "", "", err
	}

	ct := strings.TrimSpace(contentType)
	if ct == "" {
		ct = getContentTypeByFormat(fileFormat)
	}

	storageKey := prefix + p
	uploadUrl, err := presignOSSPutURL(
		l.svcCtx.Config.OSS.Endpoint,
		l.svcCtx.Config.OSS.Bucket,
		l.svcCtx.Config.OSS.AccessKeyId,
		l.svcCtx.Config.OSS.AccessKeySecret,
		storageKey,
		ct,
		int64(l.svcCtx.Config.OSS.ExpireSeconds),
	)
	if err != nil {
		return "", "", "", err
	}
	return uploadUrl, ct, storageKey, nil
}

func normalizeRelativePathForPreview(raw string) (string, error) {
	p := normalizeBuildPath(raw)
	if p == "" {
		return "", errors.New("name is required")
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return "", errors.New("invalid path")
		}
	}
	return p, nil
}

func getContentTypeByFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "txt":
		return "text/plain"
	case "html", "htm":
		return "text/html"
	case "css":
		return "text/css"
	case "js", "mjs":
		return "application/javascript"
	case "wasm":
		return "application/wasm"
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "svg":
		return "image/svg+xml"
	case "ico":
		return "image/x-icon"
	case "json", "map":
		return "application/json"
	case "xml":
		return "application/xml"
	case "woff":
		return "font/woff"
	case "woff2":
		return "font/woff2"
	case "ttf":
		return "font/ttf"
	case "otf":
		return "font/otf"
	default:
		return "application/octet-stream"
	}
}

func normalizeOSSEndpointForSign(rawEndpoint, bucket string) (scheme string, host string, err error) {
	raw := strings.TrimSpace(rawEndpoint)
	if raw == "" {
		return "", "", errors.New("empty OSS endpoint")
	}

	u, parseErr := url.Parse(raw)
	if parseErr == nil && u.Host != "" {
		scheme = strings.TrimSpace(u.Scheme)
		host = strings.TrimSpace(u.Host)
	} else {
		u2, parseErr2 := url.Parse("https://" + raw)
		if parseErr2 != nil || u2.Host == "" {
			return "", "", errors.New("invalid OSS endpoint")
		}
		scheme = "https"
		host = strings.TrimSpace(u2.Host)
	}

	b := strings.TrimSpace(bucket)
	if b != "" {
		prefix := b + "."
		for strings.HasPrefix(host, prefix) {
			host = strings.TrimPrefix(host, prefix)
		}
	}

	if scheme == "" {
		scheme = "https"
	}
	if host == "" {
		return "", "", errors.New("invalid OSS endpoint host")
	}
	return scheme, host, nil
}

func escapeOSSObjectKeyPath(objectKey string) string {
	if objectKey == "" {
		return ""
	}
	parts := strings.Split(objectKey, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func presignOSSPutURL(endpoint, bucket, accessKeyId, accessKeySecret, objectKey, contentType string, expireSeconds int64) (string, error) {
	if bucket == "" || accessKeyId == "" || accessKeySecret == "" || endpoint == "" {
		return "", errors.New("OSS not configured")
	}

	scheme, host, err := normalizeOSSEndpointForSign(endpoint, bucket)
	if err != nil {
		return "", err
	}

	expires := timeNowUnix() + expireSeconds
	canonicalResource := "/" + bucket + "/" + objectKey
	stringToSign := "PUT\n\n" + contentType + "\n" + strconv.FormatInt(expires, 10) + "\n" + canonicalResource

	signature := hmacSha1Base64(accessKeySecret, stringToSign)
	escapedObjectKey := escapeOSSObjectKeyPath(objectKey)
	return scheme + "://" + bucket + "." + host + "/" + escapedObjectKey +
		"?Expires=" + strconv.FormatInt(expires, 10) +
		"&OSSAccessKeyId=" + url.QueryEscape(accessKeyId) +
		"&Signature=" + url.QueryEscape(signature), nil
}

func timeNowUnix() int64 {
	return time.Now().Unix()
}

func hmacSha1Base64(secret string, data string) string {
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
