package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"sync/atomic"
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
	OpWatch             OpType = "Watch"
	OpIndexes           OpType = "Indexes"
	OpDrop              OpType = "Drop"
	OpBulkWrite         OpType = "BulkWrite"
	OpAll               OpType = "All"
	OpNext              OpType = "Next"
	OpDecode            OpType = "Decode"
	OpClose             OpType = "Close"
)

// OpTrace 记录操作的执行过程
type OpTrace struct {
	Ctx        context.Context
	Collection string
	Dbname     string
	Op         OpType // 操作类型
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
	ResErr      error
	handlers    []HandlerFunc
	handleIndex int32
	curOpCtx    context.Context
}

func (op *OpTrace) Next() {
	atomic.AddInt32(&op.handleIndex, 1)
	for op.handleIndex < int32(len(op.handlers)) {
		op.handlers[op.handleIndex](op)
		atomic.AddInt32(&op.handleIndex, 1)
	}
}

func (op *OpTrace) do() {
	for op.handleIndex < int32(len(op.handlers)) {
		op.handlers[op.handleIndex](op)
		atomic.AddInt32(&op.handleIndex, 1)
	}
}

func (op *OpTrace) IsCursor() bool {
	_, resCur := op.Res.(*Cursor)
	return resCur
}
