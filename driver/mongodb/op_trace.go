package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

// OpType 操作类型
type OpType string

const (
	OpInsertOne         OpType = "InsertOne"
	OpInsertMany        OpType = "InsertMany"
	OpDeleteOne         OpType = "DeleteOne"
	OpDeleteMany        OpType = "DeleteMany"
	OpUpdateOne         OpType = "UpdateOne"
	OpUpdateByID        OpType = "UpdateByID"
	OpUpdateMany        OpType = "UpdateMany"
	OpReplaceOne        OpType = "ReplaceOne"
	OpAggregate         OpType = "Aggregate"
	OpCountDocuments    OpType = "CountDocuments"
	OpDistinct          OpType = "Distinct"
	OpFind              OpType = "Find"
	OpFindOne           OpType = "FindOne"
	OpFindOneAndDelete  OpType = "FindOneAndDelete"
	OpFindOneAndReplace OpType = "FindOneAndReplace"
	OpFindOneAndUpdate  OpType = "FindOneAndUpdate"
	OpBulkWrite         OpType = "BulkWrite"
)

// OpStep 操作步骤
type OpStep string

const (
	OpStepBefore = "before"
	OpStepAfter  = "after"
)

// OpTrace 记录操作的执行过程
type OpTrace struct {
	Ctx        context.Context
	Collection string
	Dbname     string
	Op         OpType // 操作类型
	OpStep     OpStep // 操作步骤
	// Opts 所有操作对应的 Opts
	Opts interface{}
	// Insert 对应的 Documents， insert 的话只使用头一个
	InsertDocuments []interface{}
	// Filter 查询相关的过滤条件
	Filter interface{}
	// Update 更新相关的更新参数
	Update interface{}
	// Pipeline Aggregate 操作下的 pipeLine
	Pipeline interface{}
	// FieldName Distinct 操作下的 fieldName
	FieldName string
	// Models BulkWrite 操作下的 models
	Models []mongo.WriteModel
	// Res 执行结果
	Res interface{}
	// ResErr 执行错误
	ResErr error
}
