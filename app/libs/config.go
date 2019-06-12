package libs

import (
	"io/ioutil"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Config 应用配置
type Config struct {
	Dev    bool         `yaml:"dev"`    // 开发模式
	HTTP   HTTPConfig   `yaml:"http"`   // HTTP 配置
	Db     DBConfig     `yaml:"db"`     // 数据库配置
	Redis  RedisConfig  `yaml:"redis"`  // Redis 配置
	Violet VioletConfig `yaml:"violet"` // Violet 配置
	Wechat WechatConfig `yaml:"wechat"` // 微信小程序 配置
	COS    COSConfig    `yaml:"cos"`
	Email  EmailConfig  `yaml:"email"`
}

// HTTPConfig 服务器配置
type HTTPConfig struct {
	Host    string        `yaml:"host"`    // 监听地址
	Port    string        `yaml:"port"`    // 监听端口
	Session SessionConfig `yaml:"session"` // Session 配置
}

// SessionConfig Session 配置
type SessionConfig struct {
	Key     string `yaml:"key"`     // Cookies 名字
	Expires int    `yaml:"expires"` // 过期天数
}

// DBConfig 数据库配置
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DBName   string `yaml:"db"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// RedisConfig 缓存配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	Session  string `yaml:"session"`
}

// VioletConfig Violet 配置
type VioletConfig struct {
	ClientID   string `yaml:"id"`
	ClientKey  string `yaml:"key"`
	ServerHost string `yaml:"host"`
	Callback   string `yaml:"callback"`
}

// WechatConfig 微信小程序配置
type WechatConfig struct {
	AppID     string `yaml:"id"`
	AppSecret string `yaml:"secret"`
}

// COSConfig 腾讯云 对象存储配置
type COSConfig struct {
	AppID     string `yaml:"id"`
	AppSecret string `yaml:"secret"`
	URL       string `yaml:"url"`
}

// EmailConfig 邮件服务配置
type EmailConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

var config *Config

// LoadConf 从文件读取配置信息
func (c *Config) LoadConf(path string) *Config {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Panic().Err(err).Msg("Can't read config file")
	}

	err = yaml.Unmarshal(yamlFile, c)

	if err != nil {
		log.Panic().Err(err).Msg("Can't marshal config file")
	}
	log.Info().Msg("Read config from " + path)
	config = c
	return c
}

// GetConf 获取全局配置
func GetConf() *Config {
	return config
}
