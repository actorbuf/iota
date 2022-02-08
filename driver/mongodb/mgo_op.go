package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mgoOp struct {
	col *Collection
	op  *OpTrace
}

func newMgoOp(col *Collection, op *OpTrace) *mgoOp {
	return &mgoOp{col: col, op: op}
}

func (m *mgoOp) insertOneMgoOp(ctx context.Context, document interface{},
	opts ...*options.InsertOneOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.InsertOne(ctx, document, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) insertManyMgoOp(ctx context.Context, documents []interface{},
	opts ...*options.InsertManyOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.InsertMany(ctx, documents, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) deleteOneMgoOp(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.DeleteOne(ctx, filter, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) deleteManyMgoOp(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.DeleteMany(ctx, filter, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) updateByIDMgoOp(ctx context.Context, id interface{}, update interface{},
	opts ...*options.UpdateOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.UpdateByID(ctx, id, update, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) updateByOneMgoOp(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.UpdateOne(ctx, filter, update, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) updateByManyMgoOp(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.UpdateMany(ctx, filter, update, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) replaceOneMgoOp(ctx context.Context, filter interface{},
	replacement interface{}, opts ...*options.ReplaceOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.ReplaceOne(ctx, filter, replacement, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) aggregateMgoOp(ctx context.Context, pipeline interface{},
	opts ...*options.AggregateOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.Aggregate(ctx, pipeline, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) countDocumentsMgoOp(ctx context.Context, filter interface{},
	opts ...*options.CountOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.CountDocuments(ctx, filter, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) distinctMgoOp(ctx context.Context, fieldName string, filter interface{},
	opts ...*options.DistinctOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.Distinct(ctx, fieldName, filter, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) findMgoOp(ctx context.Context, filter interface{},
	opts ...*options.FindOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.Find(ctx, filter, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) findOneMgoOp(ctx context.Context, filter interface{},
	opts ...*options.FindOneOptions) HandlerFunc {
	return func(op *OpTrace) {
		res := m.col.Collection.FindOne(ctx, filter, opts...)
		op.Res = res
	}
}

func (m *mgoOp) findOneAndDeleteMgoOp(ctx context.Context, filter interface{},
	opts ...*options.FindOneAndDeleteOptions) HandlerFunc {
	return func(op *OpTrace) {
		res := m.col.Collection.FindOneAndDelete(ctx, filter, opts...)
		op.Res = res
	}
}

func (m *mgoOp) findOneAndReplaceMgoOp(ctx context.Context, filter interface{},
	replacement interface{}, opts ...*options.FindOneAndReplaceOptions) HandlerFunc {
	return func(op *OpTrace) {
		res := m.col.Collection.FindOneAndReplace(ctx, filter, replacement, opts...)
		op.Res = res
	}
}

func (m *mgoOp) findOneAndUpdateMgoOp(ctx context.Context, filter interface{},
	update interface{}, opts ...*options.FindOneAndUpdateOptions) HandlerFunc {
	return func(op *OpTrace) {
		res := m.col.Collection.FindOneAndUpdate(ctx, filter, update, opts...)
		op.Res = res
	}
}

func (m *mgoOp) watchMgoOp(ctx context.Context, pipeline interface{},
	opts ...*options.ChangeStreamOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, err := m.col.Collection.Watch(ctx, pipeline, opts...)
		op.Res = res
		op.ResErr = err
	}
}

func (m *mgoOp) indexesMgoOp(_ context.Context) HandlerFunc {
	return func(op *OpTrace) {
		res := m.col.Collection.Indexes()
		op.Res = res
	}
}

func (m *mgoOp) dropMgoOp(ctx context.Context) HandlerFunc {
	return func(op *OpTrace) {
		res := m.col.Collection.Drop(ctx)
		op.Res = res
	}
}

func (m *mgoOp) bulkWriteMgoOp(ctx context.Context, models []mongo.WriteModel,
	opts ...*options.BulkWriteOptions) HandlerFunc {
	return func(op *OpTrace) {
		res, err := m.col.Collection.BulkWrite(ctx, models, opts...)
		op.Res = res
		op.ResErr = err
	}
}

func (m *mgoOp) allMgoOp(ctx context.Context, results interface{}) HandlerFunc {
	return func(op *OpTrace) {
		err := m.col.cur.All(ctx, results)
		op.ResErr = err
	}
}

func (m *mgoOp) nextMgoOp(ctx context.Context) HandlerFunc {
	return func(op *OpTrace) {
		res := m.col.cur.Next(ctx)
		op.Res = res
	}
}

func (m *mgoOp) DecodeMgoOp(_ context.Context, val interface{}) HandlerFunc {
	return func(op *OpTrace) {
		res := m.col.cur.Decode(val)
		op.Res = res
	}
}
