package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type Cursor struct {
	*mongo.Cursor
	col *Collection
	// opCtx 生成cur的op操作对应ctx，用于透传，cur的生成到结束链路
	// 这里只是拿的前置ctx
	opCtx context.Context
}

func (cur *Cursor) All(ctx context.Context, results interface{}) error {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpAll,
		Collection: cur.col.colname,
		Dbname:     cur.col.dbname,
		curOpCtx:   cur.opCtx,
	}
	// 构造mongo执行方法
	f := newMgoOp(cur.col, opTrace).cursor(cur).allMgoOp(ctx, results)

	// 执行操作
	do(f, opTrace)

	return opTrace.ResErr
}

func (cur *Cursor) Next(ctx context.Context) bool {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpNext,
		Collection: cur.col.colname,
		Dbname:     cur.col.dbname,
		curOpCtx:   cur.opCtx,
	}
	// 构造mongo执行方法
	f := newMgoOp(cur.col, opTrace).cursor(cur).nextMgoOp(ctx)

	// 执行操作
	do(f, opTrace)

	res, _ := opTrace.Res.(bool)
	return res
}

func (cur *Cursor) Decode(ctx context.Context, val interface{}) error {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpDecode,
		Collection: cur.col.colname,
		Dbname:     cur.col.dbname,
		curOpCtx:   cur.opCtx,
	}
	// 构造mongo执行方法
	f := newMgoOp(cur.col, opTrace).cursor(cur).decodeMgoOp(ctx, val)

	// 执行操作
	do(f, opTrace)

	return opTrace.ResErr
}

func (cur *Cursor) Close(ctx context.Context) error {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpClose,
		Collection: cur.col.colname,
		Dbname:     cur.col.dbname,
		curOpCtx:   cur.opCtx,
	}
	// 构造mongo执行方法
	f := newMgoOp(cur.col, opTrace).cursor(cur).curCloseMgoOp(ctx)

	// 执行操作
	do(f, opTrace)

	return opTrace.ResErr
}
