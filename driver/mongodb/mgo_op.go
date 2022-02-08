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
	opts ...*options.InsertOneOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.InsertOne(ctx, document, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) insertManyMgoOp(ctx context.Context, documents []interface{},
	opts ...*options.InsertManyOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.InsertMany(ctx, documents, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) deleteOneMgoOp(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.DeleteOne(ctx, filter, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) deleteManyMgoOp(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.DeleteMany(ctx, filter, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) updateByIDMgoOp(ctx context.Context, id interface{}, update interface{},
	opts ...*options.UpdateOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.UpdateByID(ctx, id, update, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) updateByOneMgoOp(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.UpdateOne(ctx, filter, update, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) updateByManyMgoOp(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.UpdateMany(ctx, filter, update, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) replaceOneMgoOp(ctx context.Context, filter interface{},
	replacement interface{}, opts ...*options.ReplaceOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.ReplaceOne(ctx, filter, replacement, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) aggregateMgoOp(ctx context.Context, pipeline interface{},
	opts ...*options.AggregateOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.Aggregate(ctx, pipeline, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) countDocumentsMgoOp(ctx context.Context, filter interface{},
	opts ...*options.CountOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.CountDocuments(ctx, filter, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) distinctMgoOp(ctx context.Context, fieldName string, filter interface{},
	opts ...*options.DistinctOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.Distinct(ctx, fieldName, filter, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) findMgoOp(ctx context.Context, filter interface{},
	opts ...*options.FindOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, mgoErr := m.col.Collection.Find(ctx, filter, opts...)
		op.Res = res
		op.ResErr = mgoErr
	}
}

func (m *mgoOp) findOneMgoOp(ctx context.Context, filter interface{},
	opts ...*options.FindOneOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res := m.col.Collection.FindOne(ctx, filter, opts...)
		op.Res = res
	}
}

func (m *mgoOp) findOneAndDeleteMgoOp(ctx context.Context, filter interface{},
	opts ...*options.FindOneAndDeleteOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res := m.col.Collection.FindOneAndDelete(ctx, filter, opts...)
		op.Res = res
	}
}

func (m *mgoOp) findOneAndReplaceMgoOp(ctx context.Context, filter interface{},
	replacement interface{}, opts ...*options.FindOneAndReplaceOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res := m.col.Collection.FindOneAndReplace(ctx, filter, replacement, opts...)
		op.Res = res
	}
}

func (m *mgoOp) findOneAndUpdateMgoOp(ctx context.Context, filter interface{},
	update interface{}, opts ...*options.FindOneAndUpdateOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res := m.col.Collection.FindOneAndUpdate(ctx, filter, update, opts...)
		op.Res = res
	}
}

func (m *mgoOp) watchMgoOp(ctx context.Context, pipeline interface{},
	opts ...*options.ChangeStreamOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, err := m.col.Collection.Watch(ctx, pipeline, opts...)
		op.Res = res
		op.ResErr = err
	}
}

func (m *mgoOp) indexesMgoOp(_ context.Context) func(op *OpTrace) {
	return func(op *OpTrace) {
		res := m.col.Collection.Indexes()
		op.Res = res
	}
}

func (m *mgoOp) dropMgoOp(ctx context.Context) func(op *OpTrace) {
	return func(op *OpTrace) {
		res := m.col.Collection.Drop(ctx)
		op.Res = res
	}
}

func (m *mgoOp) bulkWriteMgoOp(ctx context.Context, models []mongo.WriteModel,
	opts ...*options.BulkWriteOptions) func(op *OpTrace) {
	return func(op *OpTrace) {
		res, err := m.col.Collection.BulkWrite(ctx, models, opts...)
		op.Res = res
		op.ResErr = err
	}
}
