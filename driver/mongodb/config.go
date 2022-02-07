package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var client *ClientInit

// Config MongoDB连接配置
type Config struct {
	// URI 连接DSN 格式: protocol://username:password@host:port/auth_db
	URI string `yaml:"uri"`
	// DBName
	DBName string `yaml:"db-name"`
	// MinPoolSize 连接池最小 默认1个
	MinPoolSize uint64 `yaml:"min-pool-size"`
	// MaxPoolSize 连接池最大 默认32
	MaxPoolSize uint64 `yaml:"max-pool-size"`
	// ConnTimeout 连接超时时间 单位秒 默认10秒
	ConnTimeout uint64 `yaml:"conn-timeout"`
	// RetryWrites 可重试
	RetryWrites bool `yaml:"retry-writes"`
	// RegistryBuilder 注册bson文档的自定义解析器 详见当前目录 codec.go 其中定义了一系列的bson文档解析器
	RegistryBuilder *bsoncodec.RegistryBuilder
	// ReadPreference 读配置
	ReadPreference *readpref.ReadPref
}

func (c *Config) Init(ctx context.Context) error {
	var err error
	c.RegistryBuilder = RegisterTimestampCodec(nil)
	client, err = ConnInit(c)
	if err != nil {
		return err
	}

	return nil
}

func GetClient() *ClientInit {
	return client
}
