package config

import (
	"fmt"
	"kafka-consumer/application/logger"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ConsumerTopicInfo struct {
	TopicBomBartenderPrinter string `yaml:"topic_bom_bartender_printer"`
}

type KafkaConfig struct {
	BootstrapServers string `yaml:"bootstrap_servers"`
	GroupID          string `yaml:"group_id"`
	AutoOffsetReset  string `yaml:"auto_offset_reset"`
}

type BartenderPrinterAPIConfig struct {
	IsCallAPI bool   `yaml:"is_call_api"`
	Method    string `yaml:"method"`
	URL       string `yaml:"url"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
}

type BartenderTrackingScriptAPI struct {
	IsCallAPI bool   `yaml:"is_call_api"`
	Method    string `yaml:"method"`
	URL       string `yaml:"url"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
}

type Config struct {
	Kafka                      KafkaConfig                `yaml:"kafka"`
	ConsumerTopicInfo          ConsumerTopicInfo          `yaml:"consumer_topic_info"`
	BartenderPrinterAPI        BartenderPrinterAPIConfig  `yaml:"bartender_printer_api"`
	BartenderTrackingScriptAPI BartenderTrackingScriptAPI `yaml:"bartender_tracking_status"`
	FileSharePath              string                     `yaml:"file_share_path"`
	Logger                     logger.ConfigLogger        `yaml:"logger"`
}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}(f)

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func GetConfigByEnv() (*Config, string, error) {
	//env := os.Getenv("ENV")
	env := "production"

	fmt.Printf("Environment: %s\n", env)

	cfgFile := "config.yml"
	if env == "production" {
		cfgFile = "config_prod.yml"
	} else if env == "qa" {
		cfgFile = "config_qa.yml"
	}
	//configPath := filepath.Join(getConfigPath(), "config", cfgFile)

	configPath := filepath.Join("config/", cfgFile)
	fmt.Printf("Config path: %s\n", configPath)

	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, "", err
	}
	return cfg, env, nil
}

func getConfigPath() string {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		return ""
	}
	exeDir := filepath.Dir(exePath)

	return exeDir
}
