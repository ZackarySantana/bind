package mongodb

import (
	"context"
	"fmt"

	"github.com/zackarysantana/bind"
	"github.com/zackarysantana/bind/parser"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	TagMongoDB      = "mongodb"
	TagExactMongoDB = "exact-mongodb"
)

func init() {
	parser.Register("ObjectID", func(s string) (bson.ObjectID, error) {
		return bson.ObjectIDFromHex(s)
	})
}

// NewSupplier creates a Supplier that looks up documents from the given MongoDB collection.
// It uses the fields of the ref to build a filter for the lookup.
// An example struct tag usage is:
//
//	mongodb:"_id=ID,name=Name"
//
// This would use the ID and Name fields of the ref struct to build a filter like:
//
//	{"_id": <value of ID field>, "name": <value of Name field>}
//
// The types of the fields will be used in the BSON encoding. The ref should be a pointer to the struct.
func NewSupplier[T any](collection *mongo.Collection, ref any) (bind.Supplier, error) {
	return bind.NewSelfSupplier(func(ctx context.Context, filter map[string]any) (T, error) {
		var out T
		return out, collection.FindOne(ctx, bson.M(filter)).Decode(&out)
	}, TagMongoDB, ref)
}

// NewExactSupplier creates a Supplier that looks up documents from the given MongoDB collection.
// It uses the options provided to build a filter for the lookup.
// An example struct tag usage is:
//
//	exact-mongodb:"_id=ObjectID(VALID_HEX_HERE),name=Name,age=Int(30)"
//
// This would build a filter like:
//
//	{"_id": ObjectID(VALID_HEX_HERE), "name": "Name", "age": 30}
//
// There's other built-in parsers for Int32, Int64, Float64, Bool, TimeRFC3339 and their Slice variants.
// You can register your own parsers with RegisterParser.
func NewExactSupplier[T any](collection *mongo.Collection) bind.Supplier {
	return bind.NewFuncSupplier(func(ctx context.Context, name string, options []string) (any, error) {
		// The 'name' is also an option.
		options = append(options, name)
		var out T
		filter, err := parser.BuildFilter(options)
		if err != nil {
			return nil, fmt.Errorf("building filter from options: %w", err)
		}
		return out, collection.FindOne(ctx, bson.M(filter)).Decode(&out)
	}, TagExactMongoDB)
}
