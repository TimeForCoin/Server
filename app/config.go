package app

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// ServerConfig 服务端配置
type ServerConfig struct {
	Port string `yaml:"port"`
}

// AppConfig 配置
type Config struct {
	Server ServerConfig `yaml:"http"`
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
