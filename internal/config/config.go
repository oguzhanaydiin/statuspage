package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Target struct {
	Name         string `yaml:"name"`
	URL          string `yaml:"url"`
	ExpectStatus int    `yaml:"expect_status"`
}

type Config struct {
	Addr       string
	Interval   time.Duration
	Timeout    time.Duration
	WebhookURL string
	Targets    []Target
}

type fileConfig struct {
	Addr       string   `yaml:"addr"`
	Interval   string   `yaml:"interval"`
	Timeout    string   `yaml:"timeout"`
	WebhookURL string   `yaml:"webhook_url"`
	Targets    []Target `yaml:"targets"`
}

func Load(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var parsed fileConfig
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		return Config{}, fmt.Errorf("parse yaml: %w", err)
	}

	interval, err := time.ParseDuration(parsed.Interval)
	if err != nil {
		return Config{}, fmt.Errorf("interval: %w", err)
	}
	timeout, err := time.ParseDuration(parsed.Timeout)
	if err != nil {
		return Config{}, fmt.Errorf("timeout: %w", err)
	}

	cfg := Config{
		Addr:       parsed.Addr,
		Interval:   interval,
		Timeout:    timeout,
		WebhookURL: parsed.WebhookURL,
		Targets:    parsed.Targets,
	}
	if cfg.Targets == nil {
		cfg.Targets = []Target{}
	}

	if v := os.Getenv("ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := os.Getenv("WEBHOOK_URL"); v != "" {
		cfg.WebhookURL = v
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	seen := make(map[string]struct{}, len(c.Targets))
	for i, t := range c.Targets {
		name := strings.TrimSpace(t.Name)
		if name == "" {
			return fmt.Errorf("target %d: empty name", i)
		}
		if err := httpURL(t.URL); err != nil {
			return fmt.Errorf("target %q: %w", name, err)
		}
		if _, dup := seen[name]; dup {
			return fmt.Errorf("duplicate target name %q", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func httpURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url %q is not http or https", raw)
	}
	if u.Host == "" {
		return fmt.Errorf("url %q is not http or https", raw)
	}
	return nil
}
