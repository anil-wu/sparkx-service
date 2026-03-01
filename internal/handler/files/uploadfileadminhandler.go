package files

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anil-wu/spark-x/internal/logic/files"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type uploadFileAdminResp struct {
	FileId        int64  `json:"fileId"`
	VersionId     int64  `json:"versionId"`
	VersionNumber int64  `json:"versionNumber"`
	ContentType   string `json:"contentType"`
}

func UploadFileAdminHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(128 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		formFile, fileHeader, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, errors.New("file is required"))
			return
		}
		defer func() { _ = formFile.Close() }()

		projectId, err := parseInt64OrDefault(r.FormValue("projectId"), 0)
		if err != nil || projectId < 0 {
			httpx.ErrorCtx(r.Context(), w, errors.New("projectId is invalid"))
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			name = fileHeader.Filename
		}
		if name == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("name is required"))
			return
		}

		fileFormat := strings.TrimSpace(r.FormValue("fileFormat"))
		if fileFormat == "" {
			fileFormat = strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), ".")
		}
		if fileFormat == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("fileFormat is required"))
			return
		}

		fileCategory := strings.TrimSpace(r.FormValue("fileCategory"))
		if fileCategory == "" {
			fileCategory = guessFileCategory(fileFormat)
		}

		contentType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
		if contentType == "" {
			contentType = strings.TrimSpace(r.FormValue("contentType"))
		}

		tmp, err := os.CreateTemp("", "sparkx-upload-*"+filepath.Ext(name))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		tmpPath := tmp.Name()
		defer func() {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}()

		hasher := sha256.New()
		sizeBytes, err := copyAndHash(tmp, formFile, hasher)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		hashHex := hex.EncodeToString(hasher.Sum(nil))
		if _, err := tmp.Seek(0, 0); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		preUploadReq := &types.PreUploadReq{
			ProjectId:    projectId,
			Name:         name,
			FileCategory: fileCategory,
			FileFormat:   fileFormat,
			SizeBytes:    sizeBytes,
			Hash:         hashHex,
			ContentType:  contentType,
		}

		preResp, err := files.NewPreUploadFileAdminLogic(r.Context(), svcCtx).PreUploadFileAdmin(preUploadReq)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		putReq, err := http.NewRequestWithContext(r.Context(), http.MethodPut, preResp.UploadUrl, tmp)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		putReq.Header.Set("Content-Type", preResp.ContentType)
		putReq.ContentLength = sizeBytes

		resp, err := http.DefaultClient.Do(putReq)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			requestId := resp.Header.Get("x-oss-request-id")
			if requestId == "" {
				requestId = resp.Header.Get("x-amz-request-id")
			}
			msg := fmt.Sprintf("oss upload failed: %s", resp.Status)
			if requestId != "" {
				msg = msg + ", requestId=" + requestId
			}
			if len(body) > 0 {
				msg = msg + ", detail=" + strings.TrimSpace(string(body))
			}
			httpx.ErrorCtx(r.Context(), w, errors.New(msg))
			return
		}

		httpx.OkJsonCtx(r.Context(), w, uploadFileAdminResp{
			FileId:        preResp.FileId,
			VersionId:     preResp.VersionId,
			VersionNumber: preResp.VersionNumber,
			ContentType:   preResp.ContentType,
		})
	}
}

func parseInt64OrDefault(raw string, defaultValue int64) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultValue, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

func copyAndHash(dst *os.File, src io.Reader, hasher hash.Hash) (int64, error) {
	w := io.MultiWriter(dst, hasher)
	return io.Copy(w, src)
}

func guessFileCategory(ext string) string {
	switch strings.ToLower(strings.TrimPrefix(ext, ".")) {
	case "zip", "rar", "7z", "tar", "gz", "bz2":
		return "archive"
	case "png", "jpg", "jpeg", "gif", "bmp", "webp", "svg":
		return "image"
	case "mp4", "avi", "mov", "wmv", "flv", "mkv":
		return "video"
	case "mp3", "wav", "ogg", "flac", "aac":
		return "audio"
	case "txt", "md", "json", "xml", "html", "css", "js", "ts", "go", "py", "java", "c", "cpp", "h":
		return "text"
	default:
		return "binary"
	}
}
