package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const (
	_traceSpanKind      = "/driver/mongodb"
	_traceComponentName = "mongo"
	_tracePeerService   = "collection"
)

type ClientInit struct {
	*mongo.Client
}

type Database struct {
	*mongo.Database
	dbname string
}

type Collection struct {
	*mongo.Collection
	dbname  string
	colname string
}

//ConnInit 初始化mongo
func ConnInit(config *Config) (*ClientInit, error) {
	if config == nil {
		return nil, fmt.Errorf("config nil")
	}
	if config.URI == "" {
		return nil, fmt.Errorf("empty uri")
	}
	if config.MinPoolSize == 0 {
		config.MinPoolSize = 1
	}
	if config.MaxPoolSize == 0 {
		config.MaxPoolSize = 32
	}
	var timeout time.Duration
	if config.ConnTimeout == 0 {
		config.ConnTimeout = 10
	}
	timeout = time.Duration(config.ConnTimeout) * time.Second
	if config.ReadPreference == nil {
		config.ReadPreference = readpref.PrimaryPreferred()
	}

	op := options.Client().ApplyURI(config.URI).SetMinPoolSize(config.MinPoolSize).
		SetMaxPoolSize(config.MaxPoolSize).SetConnectTimeout(timeout).
		SetReadPreference(config.ReadPreference).SetRetryWrites(config.RetryWrites)

	if config.RegistryBuilder != nil {
		op.SetRegistry(config.RegistryBuilder.Build())
	}

	c, err := mongo.NewClient(op)
	if err != nil {
		return nil, err
	}
	var ctx = context.Background()
	err = c.Connect(ctx)
	if err != nil {
		return nil, err
	}
	err = c.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	client = &ClientInit{c}
	return client, nil
}

func (c *ClientInit) Database(dbname string, opts ...*options.DatabaseOptions) *Database {
	db := c.Client.Database(dbname, opts...)
	return &Database{db, dbname}
}

func (db *Database) Collection(collection string, opts ...*options.CollectionOptions) *Collection {
	col := db.Database.Collection(collection, opts...)
	return &Collection{col, db.dbname, collection}
}

func (col *Collection) InsertOne(ctx context.Context, document interface{},
	opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:             ctx,
		Op:              OpInsertOne,
		Collection:      col.colname,
		Dbname:          col.dbname,
		Opts:            opts,
		InsertDocuments: []interface{}{document},
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).insertOneMgoOp(ctx, document, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.InsertOneResult)
	return res, opTrace.ResErr
}

func (col *Collection) InsertMany(ctx context.Context, documents []interface{},
	opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:             ctx,
		Op:              OpInsertMany,
		Collection:      col.colname,
		Dbname:          col.dbname,
		Opts:            opts,
		InsertDocuments: documents,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).insertManyMgoOp(ctx, documents, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.InsertManyResult)
	return res, opTrace.ResErr
}

func (col *Collection) DeleteOne(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpDeleteOne,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).deleteOneMgoOp(ctx, filter, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.DeleteResult)
	return res, opTrace.ResErr
}

func (col *Collection) DeleteMany(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpDeleteMany,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).deleteManyMgoOp(ctx, filter, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.DeleteResult)
	return res, opTrace.ResErr
}

func (col *Collection) UpdateByID(ctx context.Context, id interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpUpdateByID,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     bson.D{{"_id", id}},
		Update:     update,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).updateByIDMgoOp(ctx, id, update, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.UpdateResult)
	return res, opTrace.ResErr
}

func (col *Collection) UpdateOne(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpUpdateOne,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
		Update:     update,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).updateByOneMgoOp(ctx, filter, update, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.UpdateResult)
	return res, opTrace.ResErr
}

func (col *Collection) UpdateMany(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpUpdateMany,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
		Update:     update,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).updateByManyMgoOp(ctx, filter, update, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.UpdateResult)
	return res, opTrace.ResErr
}

func (col *Collection) ReplaceOne(ctx context.Context, filter interface{},
	replacement interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpReplaceOne,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
		Update:     replacement,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).replaceOneMgoOp(ctx, filter, replacement, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.UpdateResult)
	return res, opTrace.ResErr
}

func (col *Collection) Aggregate(ctx context.Context, pipeline interface{},
	opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpAggregate,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Pipeline:   pipeline,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).aggregateMgoOp(ctx, pipeline, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.Cursor)
	return res, opTrace.ResErr
}

func (col *Collection) CountDocuments(ctx context.Context, filter interface{},
	opts ...*options.CountOptions) (int64, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpCountDocuments,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).countDocumentsMgoOp(ctx, filter, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(int64)
	return res, opTrace.ResErr
}

func (col *Collection) Distinct(ctx context.Context, fieldName string, filter interface{},
	opts ...*options.DistinctOptions) ([]interface{}, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpDistinct,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).distinctMgoOp(ctx, fieldName, filter, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.([]interface{})
	return res, opTrace.ResErr
}

func (col *Collection) Find(ctx context.Context, filter interface{},
	opts ...*options.FindOptions) (*mongo.Cursor, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpFind,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).findMgoOp(ctx, filter, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.Cursor)
	return res, opTrace.ResErr
}

func (col *Collection) FindOne(ctx context.Context, filter interface{},
	opts ...*options.FindOneOptions) *mongo.SingleResult {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpFindOne,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}
	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).findOneMgoOp(ctx, filter, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.SingleResult)
	return res
}

func (col *Collection) FindOneAndDelete(ctx context.Context, filter interface{},
	opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpFindOneAndDelete,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).findOneAndDeleteMgoOp(ctx, filter, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.SingleResult)
	return res
}

func (col *Collection) FindOneAndReplace(ctx context.Context, filter interface{},
	replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpFindOneAndDelete,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
		Update:     replacement,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).findOneAndReplaceMgoOp(ctx, filter, replacement, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.SingleResult)
	return res
}

func (col *Collection) FindOneAndUpdate(ctx context.Context, filter interface{},
	update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpFindOneAndDelete,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
		Update:     update,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).findOneAndUpdateMgoOp(ctx, filter, update, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.SingleResult)
	return res
}

func (col *Collection) Watch(ctx context.Context, pipeline interface{},
	opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpWatch,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Pipeline:   pipeline,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).watchMgoOp(ctx, pipeline, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.ChangeStream)
	return res, opTrace.ResErr
}

func (col *Collection) Indexes(ctx context.Context) mongo.IndexView {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpIndexes,
		Collection: col.colname,
		Dbname:     col.dbname,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).indexesMgoOp(ctx)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(mongo.IndexView)
	return res
}

func (col *Collection) Drop(ctx context.Context) error {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpDrop,
		Collection: col.colname,
		Dbname:     col.dbname,
	}

	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).dropMgoOp(ctx)

	// 执行操作
	do(f, opTrace)

	return opTrace.ResErr
}

func (col *Collection) BulkWrite(ctx context.Context, models []mongo.WriteModel,
	opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpFindOneAndDelete,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Models:     models,
	}
	// 构造mongo执行方法
	f := newMgoOp(col, opTrace).bulkWriteMgoOp(ctx, models, opts...)

	// 执行操作
	do(f, opTrace)

	// 设置返回结果
	res, _ := opTrace.Res.(*mongo.BulkWriteResult)
	return res, opTrace.ResErr
}

// TODO 将 cursor下的操作记下来，不然span只有请求的部分
