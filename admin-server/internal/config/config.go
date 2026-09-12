package config

import (
	"flag"
	"fmt"
)

// Config 仅包含进程启动参数（无配置文件）。
type Config struct {
	Listen string
	Dev    bool
}

func Parse() Config {
	cfg := Config{}
	flag.StringVar(&cfg.Listen, "listen", ":8080", "HTTP 监听地址")
	flag.BoolVar(&cfg.Dev, "dev", false, "开发模式：启用固定入口 /__dev__，供 Vite 代理联调")
	flag.Parse()
	return cfg
}

func (c Config) Validate() error {
	if c.Listen == "" {
		return fmt.Errorf("必须指定 listen 监听地址")
	}
	return nil
}
