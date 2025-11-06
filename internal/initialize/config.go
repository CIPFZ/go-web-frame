package initialize

import (
	"fmt"

	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// InitConfig 加载配置文件
func InitConfig(configPath string, serviceCtx *svc.ServiceContext) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	// 热更新监听
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("config file changed: %s\n", e.Name)
		if err := v.Unmarshal(&serviceCtx.Config); err != nil {
			fmt.Printf("failed to reload config: %v\n", err)
		} else {
			fmt.Println("config reloaded successfully")
		}
	})

	if err := v.Unmarshal(&serviceCtx.Config); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}
	return v, nil
}
