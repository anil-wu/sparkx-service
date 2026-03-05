package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type S3Store struct {
	scheme        string
	host          string
	region        string
	accessKeyId   string
	accessSecret  string
	bucket        string
	expireSeconds int64
	httpClient    *http.Client
}

func NewS3Store(endpoint string, useSSL bool, region string, accessKeyId string, accessKeySecret string, bucket string, expireSeconds int64) (*S3Store, error) {
	scheme, host, err := normalizeS3Endpoint(endpoint, useSSL)
	if err != nil {
		return nil, err
	}
	return &S3Store{
		scheme:        scheme,
		host:          host,
		region:        strings.TrimSpace(region),
		accessKeyId:   strings.TrimSpace(accessKeyId),
		accessSecret:  accessKeySecret,
		bucket:        strings.TrimSpace(bucket),
		expireSeconds: expireSeconds,
		httpClient:    http.DefaultClient,
	}, nil
}

func normalizeS3Endpoint(rawEndpoint string, useSSL bool) (scheme string, host string, err error) {
	raw := strings.TrimSpace(rawEndpoint)
	if raw == "" {
		return "", "", errors.New("empty S3 endpoint")
	}
	u, parseErr := url.Parse(raw)
	if parseErr == nil && u.Host != "" {
		scheme = strings.TrimSpace(u.Scheme)
		host = strings.TrimSpace(u.Host)
	} else {
		u2, parseErr2 := url.Parse("http://" + raw)
		if parseErr2 != nil || u2.Host == "" {
			return "", "", errors.New("invalid S3 endpoint")
		}
		host = strings.TrimSpace(u2.Host)
	}
	host = strings.TrimSuffix(host, "/")
	if scheme == "" {
		if useSSL {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return scheme, host, nil
}

func escapeS3ObjectKeyPath(objectKey string) string {
	if objectKey == "" {
		return ""
	}
	parts := strings.Split(objectKey, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func awsQueryEscape(s string) string {
	escaped := url.QueryEscape(s)
	escaped = strings.ReplaceAll(escaped, "+", "%20")
	escaped = strings.ReplaceAll(escaped, "%7E", "~")
	return escaped
}

func sha256Hex(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func hmacSha256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(data))
	return mac.Sum(nil)
}

func (s *S3Store) presignURL(ctx context.Context, method string, objectKey string, contentType string, expiry time.Duration) (string, error) {
	_ = ctx
	if s.host == "" || s.bucket == "" || s.accessKeyId == "" || strings.TrimSpace(s.accessSecret) == "" {
		return "", errors.New("S3 not configured")
	}
	region := strings.TrimSpace(s.region)
	if region == "" {
		region = "us-east-1"
	}

	e := expiry
	if e <= 0 {
		e = time.Duration(s.expireSeconds) * time.Second
	}
	expiresSeconds := int64(e.Seconds())
	if expiresSeconds <= 0 {
		expiresSeconds = 1800
	}
	if expiresSeconds > 604800 {
		expiresSeconds = 604800
	}

	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	escapedKey := escapeS3ObjectKeyPath(objectKey)
	canonicalURI := "/" + awsQueryEscape(s.bucket) + "/" + escapedKey

	signedHeaders := "host"
	canonicalHeaders := "host:" + s.host + "\n"
	ct := strings.TrimSpace(contentType)
	if strings.EqualFold(method, "PUT") && ct != "" {
		signedHeaders = "content-type;host"
		canonicalHeaders = "content-type:" + ct + "\n" + canonicalHeaders
	}

	credentialScope := dateStamp + "/" + region + "/s3/aws4_request"
	credential := s.accessKeyId + "/" + credentialScope

	query := map[string]string{
		"X-Amz-Algorithm":     "AWS4-HMAC-SHA256",
		"X-Amz-Credential":    credential,
		"X-Amz-Date":          amzDate,
		"X-Amz-Expires":       strconv.FormatInt(expiresSeconds, 10),
		"X-Amz-SignedHeaders": signedHeaders,
	}

	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	canonicalQueryParts := make([]string, 0, len(keys))
	for _, k := range keys {
		canonicalQueryParts = append(canonicalQueryParts, awsQueryEscape(k)+"="+awsQueryEscape(query[k]))
	}
	canonicalQueryString := strings.Join(canonicalQueryParts, "&")

	canonicalRequest := strings.ToUpper(method) + "\n" +
		canonicalURI + "\n" +
		canonicalQueryString + "\n" +
		canonicalHeaders + "\n" +
		signedHeaders + "\n" +
		"UNSIGNED-PAYLOAD"

	stringToSign := "AWS4-HMAC-SHA256\n" +
		amzDate + "\n" +
		credentialScope + "\n" +
		sha256Hex(canonicalRequest)

	kDate := hmacSha256([]byte("AWS4"+s.accessSecret), dateStamp)
	kRegion := hmacSha256(kDate, region)
	kService := hmacSha256(kRegion, "s3")
	kSigning := hmacSha256(kService, "aws4_request")
	signature := hex.EncodeToString(hmacSha256(kSigning, stringToSign))

	finalQuery := canonicalQueryString + "&X-Amz-Signature=" + signature
	return s.scheme + "://" + s.host + canonicalURI + "?" + finalQuery, nil
}

func (s *S3Store) PresignPutObject(ctx context.Context, objectKey string, contentType string, expiry time.Duration) (string, error) {
	return s.presignURL(ctx, "PUT", objectKey, contentType, expiry)
}

func (s *S3Store) PresignGetObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	return s.presignURL(ctx, "GET", objectKey, "", expiry)
}

func (s *S3Store) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	u, err := s.presignURL(ctx, "GET", objectKey, "", time.Duration(s.expireSeconds)*time.Second)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = resp.Body.Close() }()
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("s3 get failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return resp.Body, nil
}

func (s *S3Store) DeleteObject(ctx context.Context, objectKey string) error {
	u, err := s.presignURL(ctx, "DELETE", objectKey, "", time.Duration(s.expireSeconds)*time.Second)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == 204 || resp.StatusCode == 200 || resp.StatusCode == 202 {
		return nil
	}
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("s3 delete failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(b)))
}

func (s *S3Store) StatObject(ctx context.Context, objectKey string) (*ObjectStat, error) {
	u, err := s.presignURL(ctx, "HEAD", objectKey, "", time.Duration(s.expireSeconds)*time.Second)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("s3 head failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	var sizeBytes int64
	if rawLen := strings.TrimSpace(resp.Header.Get("Content-Length")); rawLen != "" {
		if n, parseErr := strconv.ParseInt(rawLen, 10, 64); parseErr == nil {
			sizeBytes = n
		}
	}
	return &ObjectStat{
		ContentType: strings.TrimSpace(resp.Header.Get("Content-Type")),
		SizeBytes:   sizeBytes,
	}, nil
}
