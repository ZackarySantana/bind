package mongodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/zackarysantana/bind"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	TagMongoDB      = "mongodb"
	TagExactMongoDB = "exact-mongodb"
)

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
// The ref should be a pointer to the struct.
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
//	exact-mongodb:"_id=ID,name=Name"
//
// This would build a filter like:
//
//	{"_id": "ID", "name": "Name"}
//
// Note that the values in the options are treated as strings.
// If the field in the database is of a different type, the lookup may fail.
// It's recommended to use NewSupplier when possible, as it uses struct fields for type safety.
func NewExactSupplier[T any](collection *mongo.Collection) bind.Supplier {
	return bind.NewFuncSupplier(func(ctx context.Context, name string, options []string) (any, error) {
		var out T
		filter, err := buildFilterFromOptions(options)
		if err != nil {
			return nil, fmt.Errorf("building filter from options: %w", err)
		}
		return out, collection.FindOne(ctx, bson.M(filter)).Decode(&out)
	}, TagExactMongoDB)
}

func buildFilterFromOptions(options []string) (map[string]any, error) {
	filter := make(map[string]any)
	for _, opt := range options {
		parts := strings.SplitN(opt, "=", 2)
		if len(parts) == 2 {
			parsedValue, err := parseValue(parts[1])
			if err != nil {
				return nil, fmt.Errorf("parsing %s from %q: %w", parts[0], parts[1], err)
			}
			filter[parts[0]] = parsedValue
		} else {
			return nil, fmt.Errorf("invalid filter option %q", opt)
		}
	}
	return filter, nil
}
