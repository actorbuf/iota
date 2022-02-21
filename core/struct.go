package core

import "github.com/gin-gonic/gin"

type StructField struct {
	Comment         string // 字段注释
	DbFieldName     string // 结构名
	StructFieldName string // 字段名
}

func (s *StructField) GetStruct() string {
	return s.StructFieldName
}

func (s *StructField) GetDb() string {
	return s.DbFieldName
}

type IndexType string

const (
	IndexTypeUnique IndexType = "unique" // 唯一索引
	IndexTypeNormal IndexType = "normal" // 普通索引
	IndexTypeTTLIdx IndexType = "ttl"    // ttl索引
)

type IndexField struct {
	Field string // 字段
	Sort  int    // 排序 1 升序 -1倒叙
}

type IndexInfo struct {
	Type               IndexType     // 索引类型
	Name               string        // 索引名称
	ExpireAfterSeconds int64         // 指定一个以秒为单位的数值，完成 TTL设定，设定集合的生存时间 零为按字段到期索引 非零为倒计时失效索引
	Fields             []*IndexField // 联合索引
}

// GroupRouterNode 组路由节点
type GroupRouterNode struct {
	API         string            // 路由路径
	Method      string            // 请求类型 POST/GET...
	Author      string            // 接口作者
	Describe    string            // 描述
	ReqName     string            // 请求体
	RespName    string            // 响应体
	Middlewares []gin.HandlerFunc // 单一路由中间件组
}

// GroupRouter 组路由聚合
type GroupRouter struct {
	RouterPrefix string                      // 路由前缀
	Apis         map[string]*GroupRouterNode // 路由节点
	Middlewares  []gin.HandlerFunc           // 路由组统一中间件
}

type FreqConfig struct {
	Minute int64
	Hour   int64
	Day    int64
}

// FreqMap 接口限频配置
type FreqMap map[string]FreqConfig

// Exist 是否存在接口限频
func (f *FreqMap) Exist(key string) bool {
	if f == nil || len(*f) == 0 {
		return false
	}
	_, ok := (*f)[key]
	return ok
}
