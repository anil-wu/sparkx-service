package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OSSStore struct {
	endpoint        string
	bucket          string
	accessKeyId     string
	accessKeySecret string
	expireSeconds   int64

	bucketClient *oss.Bucket
}

func NewOSSStore(bucketClient *oss.Bucket, endpoint, bucket, accessKeyId, accessKeySecret string, expireSeconds int64) *OSSStore {
	return &OSSStore{
		endpoint:        endpoint,
		bucket:          bucket,
		accessKeyId:     accessKeyId,
		accessKeySecret: accessKeySecret,
		expireSeconds:   expireSeconds,
		bucketClient:    bucketClient,
	}
}

func (s *OSSStore) PresignPutObject(ctx context.Context, objectKey string, contentType string, expiry time.Duration) (string, error) {
	_ = ctx
	expireSeconds := int64(expiry.Seconds())
	if expireSeconds <= 0 {
		expireSeconds = s.expireSeconds
	}
	return presignOSSPutURL(s.endpoint, s.bucket, s.accessKeyId, s.accessKeySecret, objectKey, contentType, expireSeconds)
}

func (s *OSSStore) PresignGetObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	_ = ctx
	if s.bucketClient == nil {
		return "", errors.New("OSS not configured")
	}
	expireSeconds := int64(expiry.Seconds())
	if expireSeconds <= 0 {
		expireSeconds = s.expireSeconds
	}
	return s.bucketClient.SignURL(objectKey, "GET", expireSeconds)
}

func (s *OSSStore) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	_ = ctx
	if s.bucketClient == nil {
		return nil, errors.New("OSS not configured")
	}
	return s.bucketClient.GetObject(objectKey)
}

func (s *OSSStore) DeleteObject(ctx context.Context, objectKey string) error {
	_ = ctx
	if s.bucketClient == nil {
		return errors.New("OSS not configured")
	}
	return s.bucketClient.DeleteObject(objectKey)
}

func (s *OSSStore) StatObject(ctx context.Context, objectKey string) (*ObjectStat, error) {
	_ = ctx
	if s.bucketClient == nil {
		return nil, errors.New("OSS not configured")
	}
	meta, err := s.bucketClient.GetObjectMeta(objectKey)
	if err != nil {
		return nil, err
	}
	if meta == nil {
		return nil, errors.New("object meta not found")
	}
	stat := &ObjectStat{
		ContentType: strings.TrimSpace(meta.Get("Content-Type")),
	}
	if rawLen := strings.TrimSpace(meta.Get("Content-Length")); rawLen != "" {
		if n, parseErr := strconv.ParseInt(rawLen, 10, 64); parseErr == nil {
			stat.SizeBytes = n
		}
	}
	return stat, nil
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

	expires := time.Now().Unix() + expireSeconds
	canonicalResource := "/" + bucket + "/" + objectKey
	stringToSign := "PUT\n\n" + contentType + "\n" + strconv.FormatInt(expires, 10) + "\n" + canonicalResource

	mac := hmac.New(sha1.New, []byte(accessKeySecret))
	_, _ = mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	escapedObjectKey := escapeOSSObjectKeyPath(objectKey)
	return scheme + "://" + bucket + "." + host + "/" + escapedObjectKey +
		"?Expires=" + strconv.FormatInt(expires, 10) +
		"&OSSAccessKeyId=" + url.QueryEscape(accessKeyId) +
		"&Signature=" + url.QueryEscape(signature), nil
}

