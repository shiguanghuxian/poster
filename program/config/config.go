package config

import (
	"errors"
	"os"

	"github.com/naoina/toml"
)

// 配置文件解析

var (
	// CFG 全局配置对象
	CFG *Config
)

// Config 配置文件对应对象
type Config struct {
	Debug   bool        `toml:"debug"`
	LogPath string      `toml:"log_path"`
	HTTP    *HTTPConfig `toml:"http"`
	GRPC    *GRPCConfig `toml:"grpc"`
}

// HTTPConfig http 监听配置
type HTTPConfig struct {
	Enable       bool   `toml:"enable"`
	Address      string `toml:"address"`
	Port         int    `toml:"port"`
	ReadTimeout  int    `toml:"read_timeout"`
	WriteTimeout int    `toml:"write_timeout"`
}

// GRPCConfig grpc监听配置
type GRPCConfig struct {
	Enable  bool   `toml:"enable"`
	Address string `toml:"address"`
	Port    int    `toml:"port"`
}

// LoadConfig 读取配置
func LoadConfig(cfgPath string) (*Config, error) {
	if cfgPath == "" {
		cfgPath = "./config/cfg.toml"
	}
	f, err := os.Open(cfgPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	CFG = new(Config)
	if err := toml.NewDecoder(f).Decode(CFG); err != nil {
		return nil, err
	}
	// 检查配置项是否全
	if CFG.HTTP == nil || CFG.GRPC == nil {
		return CFG, errors.New("Configure at least one of HTTP or grpc")
	}

	return CFG, nil
}
