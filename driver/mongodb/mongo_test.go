package mongodb

import (
	"context"
	"fmt"
	"github.com/actorbuf/iota/trace"
	"github.com/actorbuf/iota/trace/jaeger"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"testing"
	"time"
)

func TestPipeline(t *testing.T) {
	//var a = bson.A{
	//	bson.M{
	//		"$match": bson.M{"a": "1000a"},
	//	},
	//	bson.M{
	//		"$project": bson.M{
	//			"_id":  1,
	//			"name": 1,
	//		},
	//	},
	//}
	//var a = bson.D{
	//	{
	//		Key:   "hello",
	//		Value: "world",
	//	},
	//}
	//var a = []interface{}{
	//	"hello",
	//	10,
	//	"world",
	//}
	var a = []bson.D{
		{
			{"$match", "1ooooo"},
		},
		{
			{"$sort", bson.M{
				"created_at": 1,
			}},
		},
	}

	var builder = RegisterTimestampCodec(nil).Build()
	vo := reflect.ValueOf(a)
	var data []interface{}
	if vo.Kind() == reflect.Slice {
		typ := vo.Type()
		if typ.Kind() == reflect.Slice {
			typ = typ.Elem()
		}
		fmt.Println(typ)

		for i := 0; i < vo.Len(); i++ {
			fmt.Println(vo.Index(i).Interface())
			var body interface{}
			b, _ := bson.MarshalExtJSONWithRegistry(builder, vo.Index(i).Interface(), true, true)
			_ = jsoniter.Unmarshal(b, &body)
			data = append(data, body)
		}
	}

	b, _ := jsoniter.Marshal(data)
	fmt.Println(string(b))
	//var typ = reflect.TypeOf(a)
	//if typ.Kind() == reflect.Slice {
	//	fmt.Println(typ.Elem().Kind())
	//	var ttyy = typ.Elem().Kind()
	//	if ttyy == reflect.Interface || ttyy == reflect.Array || ttyy == reflect.Slice {
	//		b, err := jsoniter.Marshal(a)
	//		if err != nil {
	//			panic(err)
	//		}
	//		fmt.Println(string(b))
	//	} else {
	//		b, err := bson.MarshalExtJSONWithRegistry(
	//			RegisterTimestampCodec(nil).Build(), a, true, true)
	//		if err != nil {
	//			panic(err)
	//		}
	//		fmt.Println(string(b))
	//	}
	//b, err := bson.MarshalExtJSONWithRegistry(RegisterTimestampCodec(nil).Build(), a, true, true)
	//if err != nil {
	//	panic(err)
	//}
	//b, _ := jsoniter.Marshal(a)
	//}
}

func TestSpan(t *testing.T) {
	//var exec = mongo.Pipeline{
	//	{
	//		{"$match", "1ooooo"},
	//	},
	//	{
	//		{"$sort", bson.M{
	//			"created_at": 1,
	//		}},
	//	},
	//}
	var setData = bson.M{
		"idd": "idd_dcscasdsadfasdfsdf",
	}
	var list []mongo.WriteModel
	list = append(list, mongo.NewUpdateOneModel().SetFilter(bson.M{
		"id": "id_fdsadcasdsfa",
	}).SetUpdate(setData).SetUpsert(true))

	var codes = bson.M{
		"count": 100,
		"data":  SliceStruct2MapOmitEmpty(list),
	}

	var defaultFilter string

	builder := RegisterTimestampCodec(nil).Build()
	vo := reflect.ValueOf(codes)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	fmt.Println(vo.Kind())
	switch vo.Kind() {
	case reflect.Struct, reflect.Map:
		// 正常序列化
		b, _ := bson.MarshalExtJSONWithRegistry(builder, codes, true, true)
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
			b, _ := jsoniter.Marshal(codes)
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
			b, _ := bson.MarshalExtJSONWithRegistry(builder, codes, true, true)
			defaultFilter = string(b)
		}
	}

	fmt.Println(defaultFilter)
}

func TestBulkWrite(t *testing.T) {
	config := &Config{
		URI:         "mongodb://admin:admin@10.0.0.135:27017/admin",
		DBName:      "heywoods_golang_jingliao_crm_dev",
		ConnTimeout: 10,
	}
	client, err := ConnInit(config)
	if err != nil {
		panic(err)
	}

	var ctx = context.Background()

	var updateData = bson.M{
		"$set": bson.M{
			"pinyin":  "Junesssss",
			"deleted": 0,
		},
		//"$setOnInsert": bson.M{
		//	"deleted": 0,
		//},
	}
	var opData []mongo.WriteModel
	opData = append(opData, mongo.NewUpdateOneModel().SetFilter(bson.M{
		"_id": "498b1c85be41266efb29b6a79560ec7f",
	}).SetUpsert(true).SetUpdate(updateData))

	if res, err := client.Database(config.DBName).Collection("tb_robot_friend").BulkWrite(ctx, opData); err != nil {
		panic(err)
	} else {
		fmt.Println(res)
	}
}

func TestTrace(t *testing.T) {
	config := &Config{
		URI:         "mongodb://root:root@127.0.0.1:20000/admin",
		DBName:      "test",
		MinPoolSize: 4,
		MaxPoolSize: 10,
		ConnTimeout: 10,
	}
	client, err := ConnInit(config)
	if err != nil {
		panic(err)
	}
	db := client.Database(config.DBName)

	// new jaeger
	closer, err := jaeger.NewJaeger(&jaeger.Config{
		ServiceName: "heyWoods",
		AgentHost:   "127.0.0.1",
		AgentPort:   "6831",
		LogSpans:    true,
		Disabled:    false,
		SamplerCfg: jaeger.SamplerCfg{
			Type:  jaeger.TypeConst,
			Param: jaeger.SamplerParam(1),
		},
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = closer.Close()
	}()

	// 加入jaeger
	AddMiddleware2(NewJaegerHook(true))

	start := time.Now()
	defer func() {
		fmt.Println("time", time.Now().Sub(start).Milliseconds())
	}()

	var ctx = context.Background()

	span := trace.New("test-mongo")
	ctx = trace.NewTracerContext(ctx, span)
	defer span.Finish()

	var users = []*struct {
		ID   string `bson:"_id"`
		Name string `bson:"name"`
	}{
		{ID: "1111", Name: "name1"},
		{ID: "2222", Name: "name2"},
	}

	var inserts []interface{}
	for _, user := range users {
		inserts = append(inserts, user)
	}

	// Insert
	insertOptions := &options.InsertManyOptions{}
	insertOptions.SetOrdered(false)
	insertOptions.SetBypassDocumentValidation(false)
	res, err := db.Collection("test").InsertMany(ctx, inserts, insertOptions)
	fmt.Println("res", res, "err", err)
	res, err = db.Collection("test").InsertMany(ctx, inserts)
	fmt.Println("res", res, "err", err)

	insertOneOptions := &options.InsertOneOptions{}
	insertOneOptions.SetBypassDocumentValidation(false)

	res2, err := db.Collection("test").InsertOne(ctx, inserts[0])
	fmt.Println("res", res2, "err", err)
	res2, err = db.Collection("test").InsertOne(ctx, inserts[0], insertOneOptions)
	fmt.Println("res", res2, "err", err)

	// update
	updateOneOptions := &options.UpdateOptions{}
	updateOneOptions.SetUpsert(false)
	res3, err := db.Collection("test").UpdateByID(ctx, "1111", bson.M{"$set": bson.M{"name": "name111"}}, updateOneOptions)
	fmt.Println("UpdateByID", "res", res3, "err", err)
	res3, err = db.Collection("test").UpdateByID(ctx, "1111", bson.M{"$set": bson.M{"name": "name111"}})
	fmt.Println("UpdateByID", "res", res3, "err", err)

	res4, err := db.Collection("test").UpdateOne(ctx, bson.M{"_id": "1111"}, bson.M{"$set": bson.M{"name": "name111"}}, updateOneOptions)
	fmt.Println("UpdateOne2", "res", res4, "err", err)

	res4, err = db.Collection("test").UpdateOne(ctx, bson.M{"_id": "1111"}, bson.M{"$set": bson.M{"name": "name111"}})
	fmt.Println("UpdateOne2", "res", res4, "err", err)

	res4, err = db.Collection("test").UpdateMany(ctx, bson.M{"_id": "1111"}, bson.M{"$set": bson.M{"name": "name111"}})
	fmt.Println("UpdateMany2", "res", res4, "err", err)

	// find
	findOneOptions := &options.FindOneOptions{}
	findOneOptions.SetSkip(0)

	_ = db.Collection("test").FindOne(ctx, bson.M{"_id": "1111"}, findOneOptions)

	findOptions := &options.FindOptions{}
	findOptions.SetSkip(0)
	cur, err := db.Collection("test").Find(ctx, bson.M{"_id": "1111"}, findOptions)
	fmt.Println("Find2", "err", err)

	var data []map[string]interface{}
	err = cur.All(ctx, &data)
	fmt.Println("All", "err", err)
	fmt.Println("All", "data", data)

	cur, err = db.Collection("test").Find(ctx, bson.M{"_id": "1111"})
	fmt.Println("Find2", "err", err)

	data = make([]map[string]interface{}, 0)
	for cur.Next(ctx) {
		tmp := make(map[string]interface{})
		err = cur.Decode(ctx, &tmp)
		if err != nil {
			fmt.Println("Decode", "err", err)
			return
		}
		data = append(data, tmp)
	}
	err = cur.Close(ctx)
	if err != nil {
		fmt.Println("Close", "err", err)
	}

	pipeline := mongo.Pipeline{
		bson.D{bson.E{Key: "$match", Value: bson.M{
			"_id": bson.M{"$ne": ""},
		}}},
		bson.D{bson.E{
			Key:   "$skip",
			Value: 0,
		}},
	}
	cur, err = db.Collection("test").Aggregate(ctx, pipeline)
	if err != nil {
		fmt.Println("Aggregate", "err", err)
	} else {
		err = cur.All(ctx, &data)
		fmt.Println("Aggregate All", "data", data)
		fmt.Println("Aggregate All", "err", err)
	}
}

func TestExecStr(t *testing.T) {
	insertOptions := &options.InsertManyOptions{}
	insertOptions.SetOrdered(false)
	insertOptions.SetBypassDocumentValidation(false)
	exec(insertOptions)
}

func exec(opts ...*options.InsertManyOptions) {
	opTrace := &OpTrace{
		Op:   OpInsertMany,
		Opts: opts,
	}
	fmt.Println(fmtStr(opTrace))
}

func fmtStr(op *OpTrace) string {
	fmt.Println(op.Opts)
	return execStr(op.Opts)
}

func TestMiddleware(t *testing.T) {
	f1 := func(op *OpTrace) {
		fmt.Println("f1 before")
		op.Next()
		fmt.Println("f1 after")
	}
	f2 := func(op *OpTrace) {
		fmt.Println("f2 before")
		op.Next()
		fmt.Println("f2 after")
	}
	f3 := func(op *OpTrace) {
		op.Next()
		fmt.Println("f3 after")
	}
	f4 := func(op *OpTrace) {
		fmt.Println("f4 before")
		op.Next()
	}

	AddMiddleware2(f1, f2, f3, f4)
	op := new(OpTrace)
	f := func(op *OpTrace) { fmt.Println("f do") }

	do(f, op)
}

func TestAddMiddleware2(t *testing.T) {
	config := &Config{
		URI:         "mongodb://root:root@127.0.0.1:20000/admin",
		DBName:      "test",
		MinPoolSize: 4,
		MaxPoolSize: 10,
		ConnTimeout: 10,
	}
	client, err := ConnInit(config)
	if err != nil {
		panic(err)
	}
	db := client.Database(config.DBName)

	f1 := func(op *OpTrace) {
		fmt.Println("f1 before")
		op.Next()
		fmt.Println("f1 after")
	}
	f2 := func(op *OpTrace) {
		fmt.Println("f2 before")
		op.Next()
		fmt.Println("f2 after")
	}
	f3 := func(op *OpTrace) {
		op.Next()
		fmt.Println("f3 after")
	}
	f4 := func(op *OpTrace) {
		fmt.Println("f4 before")
		op.Next()
	}

	AddMiddleware2(f1, f2, f3, f4)

	ctx := context.Background()

	var users = []*struct {
		ID   string `bson:"_id"`
		Name string `bson:"name"`
	}{
		{ID: "3333", Name: "name3"},
		{ID: "4444", Name: "name4"},
	}

	var inserts []interface{}
	for _, user := range users {
		inserts = append(inserts, user)
	}

	res2, err := db.Collection("test").InsertOne(ctx, inserts[0])
	fmt.Println("res", res2, "err", err)
}
