package mongodb

type action string

const (
	Aggregate         action = "Aggregate"
	BulkWrite         action = "BulkWrite"
	CountDocuments    action = "Count"
	DeleteOne         action = "DeleteOne"
	DeleteMany        action = "DeleteMany"
	Distinct          action = "Distinct"
	Drop              action = "Drop"
	Find              action = "Find"
	FindOne           action = "FindOne"
	FindOneAndDelete  action = "FindOneAndDelete"
	FindOneAndReplace action = "FindOneAndReplace"
	FindOneAndUpdate  action = "FindOneAndUpdate"
	InsertOne         action = "InsertOne"
	InsertMany        action = "InsertMany"
	Indexes           action = "Indexes"
	ReplaceOne        action = "ReplaceOne"
	UpdateOne         action = "UpdateOne"
	UpdateMany        action = "UpdateMany"
	Watch             action = "Watch"
)
