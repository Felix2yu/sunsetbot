package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type RequestConfig struct {
	BaseURL string `yaml:"base_url"`
}

type PushConfig struct {
	Enable     bool   `yaml:"enable"`
	NtfyServer string `yaml:"ntfy_server"`
	NtfyTopic  string `yaml:"ntfy_topic"`
	NtfyToken  string `yaml:"ntfy_token"`
}

type TaskConfig struct {
	Enable bool     `yaml:"enable"`
	Time   []string `yaml:"time"`
	Model  []string `yaml:"model"`
}

type ScheduleConfig struct {
	City             string     `yaml:"city"`
	SendTestOnStart  bool       `yaml:"send_test_on_start"`
	PushError        bool       `yaml:"push_error"`
	Morning          TaskConfig `yaml:"morning"`
	Evening          TaskConfig `yaml:"evening"`
}

type Config struct {
	Request  RequestConfig  `yaml:"request"`
	Push     PushConfig     `yaml:"push"`
	Schedule ScheduleConfig `yaml:"schedule"`
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		exe, err := os.Executable()
		if err != nil {
			exe = "."
		}
		configPath = filepath.Join(filepath.Dir(exe), "config.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			configPath = "config.yaml"
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("配置文件格式错误: %w", err)
	}

	if cfg.Push.NtfyServer == "" {
		cfg.Push.NtfyServer = "https://ntfy.sh"
	}

	return &cfg, nil
}
