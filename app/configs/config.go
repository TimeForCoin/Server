package configs

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config 应用配置
type Config struct {
	HTTP  HttpConfig  `yaml:"http"`
	Db    DBConfig    `yaml:"db"`
	Redis RedisConfig `yaml:"redis"`
}

// HttpConfig 服务器配置
type HttpConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Dev  bool   `yaml:"dev"`
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
}

// GetConf 从文件读取配置信息
func (c *Config) GetConf(path string) *Config {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Can't read config file", err.Error())
	}

	err = yaml.Unmarshal(yamlFile, c)

	if err != nil {
		fmt.Println("Can't marshal config file", err.Error())
	}
	return c
}
