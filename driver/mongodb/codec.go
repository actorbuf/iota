package mongodb

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"time"
)

// 该文件帮助自定义序列化bson文档
// RegisterTimestampCodec 注册一个针对 timestamppb.Timestamp 结构的 bson文档解析器

var (
	timeTimeType  = reflect.TypeOf(time.Time{})
	timestampType = reflect.TypeOf(&timestamppb.Timestamp{})
)

// TimestampCodec 对 timestamppb.Timestamp <-> time.Time 进行互向转换
// time.Time 在bson中被转换为 Date 对象
type TimestampCodec struct{}

func (t *TimestampCodec) EncodeValue(encodeContext bsoncodec.EncodeContext, writer bsonrw.ValueWriter, value reflect.Value) error {
	var rawv time.Time

	switch t := value.Interface().(type) {
	case *timestamppb.Timestamp:
		rawv = t.AsTime()
	case time.Time:
		rawv = t
	default:
		panic("TimestampCodec get type: " + reflect.TypeOf(value.Interface()).String() + ", not support")
	}

	enc, err := encodeContext.LookupEncoder(timeTimeType)
	if err != nil {
		return err
	}
	return enc.EncodeValue(encodeContext, writer, reflect.ValueOf(rawv.In(time.UTC)))
}

func (t *TimestampCodec) DecodeValue(decodeContext bsoncodec.DecodeContext, reader bsonrw.ValueReader, value reflect.Value) error {
	enc, err := decodeContext.LookupDecoder(timeTimeType)
	if err != nil {
		return err
	}
	var tt time.Time
	if err := enc.DecodeValue(decodeContext, reader, reflect.ValueOf(&tt).Elem()); err != nil {
		return err
	}

	ts := timestamppb.New(tt.In(time.UTC))
	value.Set(reflect.ValueOf(ts))
	return nil
}

// RegisterTimestampCodec 注册一个针对 timestamppb.Timestamp 结构的 bson文档解析器
// 将 mongodb 中 bson 字段的 Date(Go中的 time.Time ) 对象解析成 timestamppb.Timestamp
func RegisterTimestampCodec(rb *bsoncodec.RegistryBuilder) *bsoncodec.RegistryBuilder {
	if rb == nil {
		rb = bson.NewRegistryBuilder()
	}

	return rb.RegisterCodec(timestampType, &TimestampCodec{})
}
