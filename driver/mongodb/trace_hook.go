package mongodb

import (
	"fmt"
	"github.com/actorbuf/iota/trace"
	jsoniter "github.com/json-iterator/go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
)

const ExecStrLenMax = 3000

type jaegerHook struct{}

var _jaegerHook = new(jaegerHook)

func NewJaegerHook() HandlerFunc {
	return func(op *OpTrace) {
		_ = _jaegerHook.Before(op)
		op.Next()
		_ = _jaegerHook.After(op)
	}
}

func (j *jaegerHook) Before(op *OpTrace) error {
	// 开启span
	span := trace.ObtainChildSpan(op.Ctx, fmt.Sprintf("%s::%s::%s", string(op.Op), op.Dbname, op.Collection))
	op.Ctx = trace.NewTracerContext(op.Ctx, span)

	// 植入tag
	span.SetTag(trace.TagSpanKind, _traceSpanKind)
	span.SetTag(trace.TagComponent, _traceComponentName)
	span.SetTag(trace.TagPeerService, _tracePeerService)

	logField := []log.Field{
		log.String("db.exec.options", execStrLimitLen(execStr(op.Opts))),
	}

	// 限制部分字符长度
	switch op.Op {
	case OpInsertOne, OpInsertMany:
		logField = append(logField, log.String("db.exec.documents", execStrLimitLen(execStr(op.InsertDocuments))))
	case OpDeleteOne, OpDeleteMany, OpCountDocuments, OpFind, OpFindOne, OpFindOneAndDelete:
		logField = append(logField, log.String("db.exec.filter", execStrLimitLen(execStr(op.Filter))))
	case OpUpdateOne, OpUpdateMany, OpFindOneAndUpdate:
		logField = append(logField,
			log.String("db.exec.filter", execStrLimitLen(execStr(op.Filter))),
			log.String("db.exec.update", execStrLimitLen(execStr(op.Update))),
		)
	case OpReplaceOne, OpFindOneAndReplace:
		logField = append(logField,
			log.String("db.exec.filter", execStrLimitLen(execStr(op.Filter))),
			log.String("db.exec.replacement", execStrLimitLen(execStr(op.Update))),
		)
	case OpAggregate, OpWatch:
		logField = append(logField,
			log.String("db.exec.pipeline", execStrLimitLen(execStr(op.Pipeline))),
		)
	case OpDistinct:
		logField = append(logField,
			log.String("db.exec.fieldName", execStrLimitLen(execStr(op.FieldName))),
			log.String("db.exec.filter", execStrLimitLen(execStr(op.Filter))),
		)
	case OpBulkWrite:
		spanLogs := bson.M{
			"count": len(op.Models),
			"data":  SliceStruct2MapOmitEmpty(op.Models),
		}
		if len(op.Models) > 5 {
			spanLogs["data"] = SliceStruct2MapOmitEmpty(op.Models[:5])
			spanLogs["info"] = "数据过多,只显示前5项"
		}
		logField = append(logField,
			log.String("db.exec.models", execStrLimitLen(execStr(spanLogs))),
		)
	}
	// 植入对应的log
	span.LogFields(logField...)
	return nil
}

func (j *jaegerHook) After(op *OpTrace) error {
	span := trace.ObtainCtxSpan(op.Ctx)
	if trace.IsNoopSpan(span) {
		return nil
	}

	// TODO 记录到curAll跟curNext

	// 不是cursor类型的。直接记录执行结果
	if !op.IsCursor() && op.ResErr != nil {
		defer span.Finish()
	}

	if op.ResErr != nil {
		ext.Error.Set(span, true)
		span.LogFields(log.String("db.exec.err", op.ResErr.Error()))
	}
	return nil
}

func execStrLimitLen(execStr string) string {
	if len(execStr) > ExecStrLenMax {
		return execStr[0:ExecStrLenMax]
	}
	return execStr
}

func execStr(exec interface{}) string {
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
		if childKind.Kind() == reflect.Ptr {
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
		case reflect.Struct:
			b, _ := jsoniter.Marshal(exec)
			defaultFilter = string(b)
		default:
			b, _ := bson.MarshalExtJSONWithRegistry(builder, exec, true, true)
			defaultFilter = string(b)
		}
	}
	return defaultFilter
}
