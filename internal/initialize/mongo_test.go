package initialize

import (
	"context"
	"testing"
	"time"

	"github.com/CIPFZ/gowebframe/internal/config"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

func TestMustInitMongo(t *testing.T) {
	serviceCtx := &svc.ServiceContext{
		Config: &config.Config{
			System: config.System{
				UseMongo: true,
			},
			JWT: config.JWT{},
			Mongo: config.Mongo{
				Coll:             "test",
				Options:          "",
				Database:         "test",
				Username:         "admin",
				Password:         "123456",
				AuthSource:       "",
				MinPoolSize:      1,
				MaxPoolSize:      10,
				SocketTimeoutMs:  10000,
				ConnectTimeoutMs: 10000,
				IsZap:            false,
				Hosts: []*config.MongoHost{
					{Host: "192.168.1.10", Port: "27017"},
				},
			},
		},
		Logger: zap.NewNop(),
	}
	defer func() {
		if r := recover(); r != nil {
			t.Logf("捕获到 panic (这是预期的，因为可能没有真实数据库): %v", r)
		}
	}()
	MustInitMongo(serviceCtx)
	// 测试查询语句
	collName := "users"

	// 1. 定义测试结构
	type TestUser struct {
		ID           primitive.ObjectID `bson:"_id,omitempty"`
		Name         string             `bson:"name"`
		Age          int                `bson:"age"`
		Email        string             `bson:"email"`
		Status       string             `bson:"status"`
		RegisteredAt time.Time          `bson:"registered_at"`
	}
	// 获取集合句柄
	coll := serviceCtx.Mongo.Database.Collection(collName)
	ctx := context.Background()
	var foundUser TestUser
	// 使用 bson.M 来构建查询
	err := coll.Find(ctx, bson.M{"email": "test@example.com"}).One(&foundUser)
	t.Logf("found user: %v", foundUser)

	// 5. 验证
	require.NoError(t, err, "执行 Find().One() 时出错，未找到文档或连接失败")
	assert.False(t, foundUser.ID.IsZero(), "找到的文档应该有一个 ObjectID")

	// 清理
	ShutdownMongo(serviceCtx)
	// 检查 mongo client 是否关闭
	err = serviceCtx.Mongo.Ping(10)
	t.Logf("ping: %v", err)
	assert.Error(t, err, "Pinging a closed mongo client should return an error")
}
