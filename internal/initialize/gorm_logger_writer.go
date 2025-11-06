package initialize

import (
	"fmt"

	"github.com/CIPFZ/gowebframe/internal/config"

	"gorm.io/gorm/logger"
)

type Writer struct {
	config config.Mysql
	writer logger.Writer
}

func NewWriter(config config.Mysql) *Writer {
	return &Writer{config: config}
}

// Printf 格式化打印日志
func (c *Writer) Printf(message string, data ...any) {

	// 当有日志时候均需要输出到控制台
	fmt.Printf(message, data...)

	//// 当开启了zap的情况，会打印到日志记录
	//if c.config.LogZap {
	//	switch c.config.LogLevel() {
	//	case logger.Silent:
	//		global.G.Logger.Debug(fmt.Sprintf(message, data...))
	//	case logger.Error:
	//		global.G.Logger.Error(fmt.Sprintf(message, data...))
	//	case logger.Warn:
	//		global.G.Logger.Warn(fmt.Sprintf(message, data...))
	//	case logger.Info:
	//		global.G.Logger.Info(fmt.Sprintf(message, data...))
	//	default:
	//		global.G.Logger.Info(fmt.Sprintf(message, data...))
	//	}
	//	return
	//}
}
