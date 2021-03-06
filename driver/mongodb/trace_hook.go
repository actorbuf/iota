package mongodb

import (
	"context"
	"fmt"
	"github.com/actorbuf/iota/trace"
	jsoniter "github.com/json-iterator/go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
)

const DefaultExecStrLenMax = 3000

type jaegerHook struct {
	// useCur 是否使用 jaeger 的all跟close方法。
	// 建议设置为true，且使用 iota mongodb 的 Cursor
	// 如果开启 jaegerHook，且使用 iota mongodb 的find等返回cur的方法，但是没有使用 iota cur 做all、decode等方法处理
	// 而是拿 iota cur 下的 mongo.Cursor 做处理，设置 useCur false，否则 span将不会 Finish。
	useCur bool

	// execEncode 参数解析器
	execEncode func(exec interface{}) string

	// execStrLenMax 记录执行结果的最大长度，默认 DefaultExecStrLenMax
	execStrLenMax int
}

var _jaegerHook = &jaegerHook{useCur: true, execEncode: execStr, execStrLenMax: DefaultExecStrLenMax}

func NewJaegerHook(useCur bool) HandlerFunc {
	_jaegerHook.useCur = useCur
	return func(op *OpTrace) {
		_ = _jaegerHook.Before(op)
		op.Next()
		_ = _jaegerHook.After(op)
	}
}

// CustomJaegerHook 自定义jaegerHook
func CustomJaegerHook(useCur bool, execEncode func(exec interface{}) string, execStrLenMax int) HandlerFunc {
	_jaegerHook := &jaegerHook{useCur: useCur, execEncode: execEncode, execStrLenMax: execStrLenMax}
	return func(op *OpTrace) {
		_ = _jaegerHook.Before(op)
		op.Next()
		_ = _jaegerHook.After(op)
	}
}

func (j *jaegerHook) Before(op *OpTrace) error {
	// cur 相关操作没有前置
	if op.Op == OpAll || op.Op == OpClose || op.Op == OpNext || op.Op == OpDecode {
		return nil
	}

	// 开启span
	span := trace.ObtainChildSpan(op.Ctx, fmt.Sprintf("%s::%s::%s", string(op.Op), op.Dbname, op.Collection))
	op.Ctx = trace.NewTracerContext(op.Ctx, span)

	// 植入tag
	span.SetTag(trace.TagSpanKind, _traceSpanKind)
	span.SetTag(trace.TagComponent, _traceComponentName)
	span.SetTag(trace.TagPeerService, _tracePeerService)

	logField := []log.Field{
		log.String("db.exec.options", j.execStrLimitLen(j.execEncode(op.Opts))),
	}

	// 限制部分字符长度
	switch op.Op {
	case OpInsertOne, OpInsertMany:
		logField = append(logField, log.String("db.exec.documents", j.execStrLimitLen(j.execEncode(op.InsertDocuments))))
	case OpDeleteOne, OpDeleteMany, OpCountDocuments, OpFind, OpFindOne, OpFindOneAndDelete:
		logField = append(logField, log.String("db.exec.filter", j.execStrLimitLen(j.execEncode(op.Filter))))
	case OpUpdateOne, OpUpdateMany, OpFindOneAndUpdate:
		logField = append(logField,
			log.String("db.exec.filter", j.execStrLimitLen(j.execEncode(op.Filter))),
			log.String("db.exec.update", j.execStrLimitLen(j.execEncode(op.Update))),
		)
	case OpReplaceOne, OpFindOneAndReplace:
		logField = append(logField,
			log.String("db.exec.filter", j.execStrLimitLen(j.execEncode(op.Filter))),
			log.String("db.exec.replacement", j.execStrLimitLen(j.execEncode(op.Update))),
		)
	case OpAggregate, OpWatch:
		logField = append(logField,
			log.String("db.exec.pipeline", j.execStrLimitLen(j.execEncode(op.Pipeline))),
		)
	case OpDistinct:
		logField = append(logField,
			log.String("db.exec.fieldName", j.execStrLimitLen(j.execEncode(op.FieldName))),
			log.String("db.exec.filter", j.execStrLimitLen(j.execEncode(op.Filter))),
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
			log.String("db.exec.models", j.execStrLimitLen(j.execEncode(spanLogs))),
		)
	}
	// 植入对应的log
	span.LogFields(logField...)
	return nil
}

func (j *jaegerHook) After(op *OpTrace) error {
	// 如果没有使用cur，那么对cur不做处理
	if !j.useCur && (op.Op == OpAll || op.Op == OpClose) {
		return nil
	}

	// next跟decode不处理
	if op.Op == OpDecode || op.Op == OpNext {
		return nil
	}

	// 获取对应的ctx
	var spanCtx context.Context
	if op.Op == OpAll || op.Op == OpClose {
		spanCtx = op.curOpCtx
	} else {
		spanCtx = op.Ctx
	}

	span := trace.ObtainCtxSpan(spanCtx)
	if trace.IsNoopSpan(span) {
		return nil
	}

	// 不是用cur的直接设置span为finish
	if !j.useCur {
		defer span.Finish()
	} else {
		// 如果是all或者close，记录到错误
		if op.Op == OpAll || op.Op == OpClose {
			defer span.Finish()
		} else if !op.IsCursor() || op.ResErr != nil {
			// 如果返回cur获取有错误直接记录，返回cur的等到curClose或者curAll再做处理
			defer span.Finish()
		}
	}

	// 记录错误信息
	if op.ResErr != nil {
		ext.Error.Set(span, true)
		span.LogFields(log.String("db.exec.err", op.ResErr.Error()))
	}
	return nil
}

func (j *jaegerHook) execStrLimitLen(execStr string) string {
	if len(execStr) > j.execStrLenMax {
		return execStr[0:j.execStrLenMax]
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
