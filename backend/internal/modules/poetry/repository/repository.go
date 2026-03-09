package repository

import (
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/model"
	"gorm.io/gorm"
)

// PoetryRepo 聚合了所有诗词相关的 DB 操作
// 这样在 Service 层只需要依赖这一个 Repo 即可
type PoetryRepo struct {
	db *gorm.DB
}

func NewPoetryRepo(db *gorm.DB) *PoetryRepo {
	// 注册自定义连接表模型
	// 这一步告诉 GORM：这个多对多关系使用 PoemTagRel 结构体
	// 从而避免 GORM 尝试去读写不存在的 id/created_at 字段
	err := db.SetupJoinTable(&model.PoemWork{}, "Tags", &model.PoemTagRel{})
	if err != nil {
		// 在初始化阶段报错通常应该 Panic 或者记录 Error
		// 这里视您的错误处理策略而定
		panic("failed to setup join table: " + err.Error())
	}
	return &PoetryRepo{db: db}
}
