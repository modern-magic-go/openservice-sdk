package openservice

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	BaseURL        string        `json:"baseURL" yaml:"baseURL"`
	MID            string        `json:"mid" yaml:"mid"`
	Secret         string        `json:"secret" yaml:"secret"`
	AESKey         string        `json:"aesKey" yaml:"aesKey"`
	AESIv          string        `json:"aesIv" yaml:"aesIv"`
	ProviderAppId  string        `json:"providerAppId" yaml:"providerAppId"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
}

func ValidateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return ErrMissingBaseURL
	}
	parsed, err := url.Parse(strings.TrimSpace(cfg.BaseURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("%w: %s", ErrInvalidBaseURL, strings.TrimSpace(cfg.BaseURL))
	}
	if strings.TrimSpace(cfg.MID) == "" {
		return ErrMissingMID
	}
	if strings.TrimSpace(cfg.Secret) == "" {
		return ErrMissingSecret
	}
	if cfg.Timeout <= 0 {
		return ErrInvalidTimeout
	}
	return nil
}

func normalizeConfig(cfg Config) Config {
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	cfg.MID = strings.TrimSpace(cfg.MID)
	cfg.Secret = strings.TrimSpace(cfg.Secret)
	cfg.AESKey = strings.TrimSpace(cfg.AESKey)
	cfg.AESIv = strings.TrimSpace(cfg.AESIv)
	cfg.ProviderAppId = strings.TrimSpace(cfg.ProviderAppId)
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}
	return cfg
}
