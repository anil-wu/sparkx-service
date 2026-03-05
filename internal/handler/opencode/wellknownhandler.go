package opencode

import (
	"encoding/json"
	"net/http"
	"regexp"
	"crypto/subtle"
	"os"
	"strconv"
	"strings"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
)

type openCodeModelCfg struct {
	Name string `json:"name,omitempty"`
}

type openCodeProviderCfg struct {
	Npm     string                      `json:"npm,omitempty"`
	Options map[string]interface{}      `json:"options,omitempty"`
	Models  map[string]openCodeModelCfg `json:"models,omitempty"`
}

type openCodeConfig struct {
	Schema     string                         `json:"$schema,omitempty"`
	Model      string                         `json:"model,omitempty"`
	SmallModel string                         `json:"small_model,omitempty"`
	Provider   map[string]openCodeProviderCfg `json:"provider,omitempty"`
}

var providerKeySanitizer = regexp.MustCompile(`[^a-z0-9_-]+`)

func sanitizeProviderKey(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = providerKeySanitizer.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "provider"
	}
	return s
}

func allowProviderApiKey(r *http.Request) bool {
	expected := strings.TrimSpace(os.Getenv("API_SERVICE_API_KEY"))
	if expected == "" {
		expected = "123456"
	}
	provided := strings.TrimSpace(r.Header.Get("X-API-KEY"))
	if provided == "" {
		provided = strings.TrimSpace(r.Header.Get("API_SERVICE_API_KEY"))
	}
	if provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func WellKnownOpenCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		includeApiKey := allowProviderApiKey(r)
		cfg := openCodeConfig{
			Schema:   "https://opencode.ai/config.json",
			Provider: map[string]openCodeProviderCfg{},
		}

		if svcCtx == nil || svcCtx.DB == nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(cfg)
			return
		}

		var providers []model.LlmProviders
		if err := svcCtx.DB.WithContext(r.Context()).Order("`id` ASC").Find(&providers).Error; err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(cfg)
			return
		}

		var models []model.LlmModels
		if err := svcCtx.DB.WithContext(r.Context()).Order("`id` ASC").Find(&models).Error; err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(cfg)
			return
		}

		modelsByProvider := make(map[uint64][]model.LlmModels, len(providers))
		for _, m := range models {
			modelsByProvider[m.ProviderId] = append(modelsByProvider[m.ProviderId], m)
		}

		usedKeys := map[string]bool{}
		for _, p := range providers {
			key := sanitizeProviderKey(p.Name)
			if key == "provider" {
				key = key + "-" + strconv.FormatUint(p.Id, 10)
			}
			if usedKeys[key] {
				key = key + "-" + strconv.FormatUint(p.Id, 10)
			}
			usedKeys[key] = true

			providerCfg := openCodeProviderCfg{
				Npm:     "@ai-sdk/openai",
				Options: map[string]interface{}{},
				Models:  map[string]openCodeModelCfg{},
			}

			if strings.TrimSpace(p.BaseUrl) != "" {
				providerCfg.Options["baseURL"] = strings.TrimSpace(p.BaseUrl)
			}
			if includeApiKey && p.ApiKey.Valid && strings.TrimSpace(p.ApiKey.String) != "" {
				providerCfg.Options["apiKey"] = p.ApiKey.String
			}

			for _, m := range modelsByProvider[p.Id] {
				if strings.TrimSpace(m.ModelName) == "" {
					continue
				}
				name := strings.TrimSpace(m.ModelName)
				providerCfg.Models[name] = openCodeModelCfg{Name: name}
				if cfg.Model == "" && m.ModelType == "llm" {
					cfg.Model = key + "/" + name
				}
				if cfg.SmallModel == "" && m.ModelType == "llm" {
					cfg.SmallModel = key + "/" + name
				}
			}

			cfg.Provider[key] = providerCfg
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(cfg)
	}
}
