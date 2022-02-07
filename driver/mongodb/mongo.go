package mongodb

import (
	"context"
	"fmt"
	"github.com/actorbuf/iota/trace"
	jsoniter "github.com/json-iterator/go"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"reflect"
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
	span := spanFunc(ctx, col.dbname, col.colname, InsertOne, document)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.InsertOne(ctx, document, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) InsertOne1(ctx context.Context, document interface{},
	opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:             ctx,
		Op:              OpInsertOne,
		OpStep:          OpStepBefore,
		Collection:      col.colname,
		Dbname:          col.dbname,
		Opts:            opts,
		InsertDocuments: []interface{}{document},
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return nil, err
	}

	// 执行Mongo
	res, mgoErr := col.Collection.InsertOne(ctx, document, opts...)

	// 执行后置操作
	opTrace.Res = res
	opTrace.ResErr = mgoErr
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	// TODO 记录后置操作错误
	return res, mgoErr
}

func (col *Collection) InsertMany(ctx context.Context, documents []interface{},
	opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	span := spanFunc(ctx, col.dbname, col.colname, InsertMany, documents)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.InsertMany(ctx, documents, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) InsertMany2(ctx context.Context, documents []interface{},
	opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:             ctx,
		Op:              OpInsertMany,
		OpStep:          OpStepBefore,
		Collection:      col.colname,
		Dbname:          col.dbname,
		Opts:            opts,
		InsertDocuments: documents,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return nil, err
	}

	// 执行Mongo
	res, mgoErr := col.Collection.InsertMany(ctx, documents, opts...)

	// 执行后置操作
	opTrace.Res = res
	opTrace.ResErr = mgoErr
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	// TODO 记录后置操作错误
	return res, mgoErr
}

func (col *Collection) DeleteOne(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	span := spanFunc(ctx, col.dbname, col.colname, DeleteOne, filter)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.DeleteOne(ctx, filter, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) DeleteOne2(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpDeleteOne,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return nil, err
	}

	// 执行Mongo
	res, mgoErr := col.Collection.DeleteOne(ctx, filter, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.ResErr = mgoErr
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	// TODO 记录后置操作错误

	return res, mgoErr
}

func (col *Collection) DeleteMany(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	span := spanFunc(ctx, col.dbname, col.colname, DeleteMany, filter)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.DeleteMany(ctx, filter, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) DeleteMany2(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpDeleteMany,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return nil, err
	}

	// 执行Mongo
	res, mgoErr := col.Collection.DeleteMany(ctx, filter, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.ResErr = mgoErr
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	// TODO 记录后置操作错误

	return res, mgoErr
}

func (col *Collection) UpdateByID(ctx context.Context, id interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpUpdateByID,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     bson.D{{"_id", id}},
		Update:     update,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return nil, err
	}

	// 执行Mongo
	res, mgoErr := col.Collection.UpdateByID(ctx, id, update, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.ResErr = mgoErr
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	// TODO 记录后置操作错误
	return res, mgoErr
}

func (col *Collection) UpdateOne(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	spanInfo := map[string]interface{}{"filter": filter, "update": update}
	span := spanFunc(ctx, col.dbname, col.colname, UpdateOne, spanInfo)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) UpdateOne2(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpUpdateOne,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
		Update:     update,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return nil, err
	}

	// 执行Mongo
	res, mgoErr := col.Collection.UpdateOne(ctx, filter, update, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.ResErr = mgoErr
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	// TODO 记录后置操作错误

	return res, mgoErr
}

func (col *Collection) UpdateMany(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	spanInfo := map[string]interface{}{"filter": filter, "update": update}
	span := spanFunc(ctx, col.dbname, col.colname, UpdateMany, spanInfo)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) UpdateMany2(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpUpdateMany,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
		Update:     update,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return nil, err
	}
	// 执行Mongo
	res, mgoErr := col.Collection.UpdateMany(ctx, filter, update, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.ResErr = mgoErr
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	// TODO 记录后置操作错误
	return res, mgoErr
}

func (col *Collection) ReplaceOne(ctx context.Context, filter interface{},
	replacement interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	spanInfo := map[string]interface{}{"filter": filter, "replacement": replacement}
	span := spanFunc(ctx, col.dbname, col.colname, ReplaceOne, spanInfo)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.ReplaceOne(ctx, filter, replacement, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) ReplaceOne2(ctx context.Context, filter interface{},
	replacement interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpReplaceOne,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
		Update:     replacement,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return nil, err
	}
	// 执行Mongo
	res, mgoErr := col.Collection.ReplaceOne(ctx, filter, replacement, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.ResErr = mgoErr
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	// TODO 记录后置操作错误
	return res, mgoErr
}

func (col *Collection) Aggregate(ctx context.Context, pipeline interface{},
	opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	span := spanFunc(ctx, col.dbname, col.colname, Aggregate, pipeline)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) Aggregate2(ctx context.Context, pipeline interface{},
	opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpAggregate,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Pipeline:   pipeline,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return nil, err
	}
	// 执行Mongo
	res, mgoErr := col.Collection.Aggregate(ctx, pipeline, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.ResErr = mgoErr
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	// TODO 记录后置操作错误
	return res, mgoErr
}

func (col *Collection) CountDocuments(ctx context.Context, filter interface{},
	opts ...*options.CountOptions) (int64, error) {
	span := spanFunc(ctx, col.dbname, col.colname, CountDocuments, filter)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.CountDocuments(ctx, filter, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) CountDocuments2(ctx context.Context, filter interface{},
	opts ...*options.CountOptions) (int64, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpCountDocuments,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return 0, err
	}
	// 执行Mongo
	res, mgoErr := col.Collection.CountDocuments(ctx, filter, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.ResErr = mgoErr
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	// TODO 记录后置操作错误
	return res, mgoErr
}

func (col *Collection) Distinct(ctx context.Context, fieldName string, filter interface{},
	opts ...*options.DistinctOptions) ([]interface{}, error) {
	spanInfo := map[string]interface{}{"fieldName": fieldName, "filter": filter}
	span := spanFunc(ctx, col.dbname, col.colname, Distinct, spanInfo)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.Distinct(ctx, fieldName, filter, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) Distinct2(ctx context.Context, fieldName string, filter interface{},
	opts ...*options.DistinctOptions) ([]interface{}, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpDistinct,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return nil, err
	}
	// 执行Mongo
	res, mgoErr := col.Collection.Distinct(ctx, fieldName, filter, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.ResErr = mgoErr
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	// TODO 记录后置操作错误
	return res, mgoErr
}

func (col *Collection) Find(ctx context.Context, filter interface{},
	opts ...*options.FindOptions) (*mongo.Cursor, error) {
	span := spanFunc(ctx, col.dbname, col.colname, Find, filter)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.Find(ctx, filter, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) Find2(ctx context.Context, filter interface{},
	opts ...*options.FindOptions) (*mongo.Cursor, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpFind,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return nil, err
	}
	// 执行Mongo
	res, mgoErr := col.Collection.Find(ctx, filter, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.ResErr = mgoErr
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	// TODO 记录后置操作错误
	return res, mgoErr
}

func (col *Collection) FindOne(ctx context.Context, filter interface{},
	opts ...*options.FindOneOptions) *mongo.SingleResult {
	span := spanFunc(ctx, col.dbname, col.colname, FindOne, filter)
	defer span.Finish()
	defer spanFinishAt(span)
	res := col.Collection.FindOne(ctx, filter, opts...)
	return res
}

func (col *Collection) FindOne2(ctx context.Context, filter interface{},
	opts ...*options.FindOneOptions) *mongo.SingleResult {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpFindOne,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return &mongo.SingleResult{}
	}
	// 执行Mongo
	res := col.Collection.FindOne(ctx, filter, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	return res
}

func (col *Collection) FindOneAndDelete(ctx context.Context, filter interface{},
	opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult {
	span := spanFunc(ctx, col.dbname, col.colname, FindOneAndDelete, filter)
	defer span.Finish()
	defer spanFinishAt(span)
	res := col.Collection.FindOneAndDelete(ctx, filter, opts...)
	return res
}

func (col *Collection) FindOneAndDelete2(ctx context.Context, filter interface{},
	opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpFindOneAndDelete,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return &mongo.SingleResult{}
	}
	// 执行Mongo
	res := col.Collection.FindOneAndDelete(ctx, filter, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	return res
}

func (col *Collection) FindOneAndReplace(ctx context.Context, filter interface{},
	replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult {
	spanInfo := map[string]interface{}{"filter": filter, "replacement": replacement}
	span := spanFunc(ctx, col.dbname, col.colname, FindOneAndReplace, spanInfo)
	defer span.Finish()
	defer spanFinishAt(span)
	res := col.Collection.FindOneAndReplace(ctx, filter, replacement, opts...)
	return res
}

func (col *Collection) FindOneAndReplace2(ctx context.Context, filter interface{},
	replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpFindOneAndDelete,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
		Update:     replacement,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return &mongo.SingleResult{}
	}
	// 执行Mongo
	res := col.Collection.FindOneAndReplace(ctx, filter, replacement, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	return res
}

func (col *Collection) FindOneAndUpdate(ctx context.Context, filter interface{},
	update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	spanInfo := map[string]interface{}{"filter": filter, "update": update}
	span := spanFunc(ctx, col.dbname, col.colname, FindOneAndUpdate, spanInfo)
	defer span.Finish()
	defer spanFinishAt(span)
	res := col.Collection.FindOneAndUpdate(ctx, filter, update, opts...)
	return res
}

func (col *Collection) FindOneAndUpdate2(ctx context.Context, filter interface{},
	update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpFindOneAndDelete,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Filter:     filter,
		Update:     update,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return &mongo.SingleResult{}
	}
	// 执行Mongo
	res := col.Collection.FindOneAndUpdate(ctx, filter, update, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	return res
}

func (col *Collection) Watch(ctx context.Context, pipeline interface{},
	opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	span := spanFunc(ctx, col.dbname, col.colname, Watch, pipeline)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.Watch(ctx, pipeline, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) Indexes(ctx context.Context) mongo.IndexView {
	span := spanFunc(ctx, col.dbname, col.colname, Indexes, "")
	defer span.Finish()
	defer spanFinishAt(span)
	res := col.Collection.Indexes()
	return res
}

func (col *Collection) Drop(ctx context.Context) error {
	span := spanFunc(ctx, col.dbname, col.colname, Drop, "")
	defer span.Finish()
	defer spanFinishAt(span)
	err := col.Collection.Drop(ctx)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return err
}

func (col *Collection) BulkWrite(ctx context.Context, models []mongo.WriteModel,
	opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	spanLogs := bson.M{
		"count": len(models),
		"data":  SliceStruct2MapOmitEmpty(models),
	}
	if len(models) > 5 {
		spanLogs["data"] = SliceStruct2MapOmitEmpty(models[:5])
		spanLogs["info"] = "数据过多,只显示前5项"
	}

	span := spanFunc(ctx, col.dbname, col.colname, BulkWrite, spanLogs)
	defer span.Finish()
	defer spanFinishAt(span)
	res, err := col.Collection.BulkWrite(ctx, models, opts...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return res, err
}

func (col *Collection) BulkWrite2(ctx context.Context, models []mongo.WriteModel,
	opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	// 构造OpTrace
	opTrace := &OpTrace{
		Ctx:        ctx,
		Op:         OpFindOneAndDelete,
		OpStep:     OpStepBefore,
		Collection: col.colname,
		Dbname:     col.dbname,
		Opts:       opts,
		Models:     models,
	}
	// 执行 before
	err := middlewareBefore(opTrace)
	if err != nil {
		return nil, err
	}
	// 执行Mongo
	res, mgoErr := col.Collection.BulkWrite(ctx, models, opts...)
	// 执行后置操作
	opTrace.Res = res
	opTrace.OpStep = OpStepAfter
	_ = middlewareAfter(opTrace)
	return res, mgoErr
}

// TODO 将 cursor下的操作记下来，不然span只有请求的部分

func spanFunc(ctx context.Context, dbname, collection string, action action, exec interface{}) opentracing.Span {
	span := trace.ObtainChildSpan(ctx, string(action)+"::"+dbname+"::"+collection)
	span.SetTag(trace.TagSpanKind, _traceSpanKind)
	span.SetTag(trace.TagComponent, _traceComponentName)
	span.SetTag(trace.TagPeerService, _tracePeerService)
	span.SetTag(trace.TagBeginAt, time.Now().Format("2006-01-02 15:04:05.000"))
	defaultFilter := fmt.Sprintf("%+v", exec)
	builder := RegisterTimestampCodec(nil).Build()
	vo := reflect.ValueOf(exec)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	switch vo.Kind() {
	case reflect.Struct, reflect.Map:
		// 正常序列化
		b, _ := bson.MarshalExtJSONWithRegistry(builder, exec, true, true)
		defaultFilter = string(b)
	case reflect.Slice, reflect.Array:
		childKind := vo.Type()
		if childKind.Kind() == reflect.Ptr || childKind.Kind() == reflect.Slice ||
			childKind.Kind() == reflect.Array {
			childKind = childKind.Elem()
		}
		switch childKind.Kind() {
		case reflect.Interface:
			// 对于[]interface型式 使用json原样
			b, _ := jsoniter.Marshal(exec)
			defaultFilter = string(b)
		case reflect.Slice, reflect.Array:
			// 对于[][]型式 使用append
			var data []interface{}
			for i := 0; i < vo.Len(); i++ {
				var body interface{}
				b, _ := bson.MarshalExtJSONWithRegistry(builder, vo.Index(i).Interface(), true, true)
				_ = jsoniter.Unmarshal(b, &body)
				data = append(data, body)
			}
			b, _ := jsoniter.Marshal(data)
			defaultFilter = string(b)
		default:
			b, _ := bson.MarshalExtJSONWithRegistry(builder, exec, true, true)
			defaultFilter = string(b)
		}
	}
	span.LogFields(
		log.String(trace.LogEvent, string(action)+"::"+dbname+"::"+collection),
		log.String("db.exec", defaultFilter),
	)
	return span
}

func spanFinishAt(span opentracing.Span) {
	span.SetTag(trace.TagFinishAt, time.Now().Format("2006-01-02 15:04:05.000"))
}
