package mongodb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/zackarysantana/bind"
	mmongodb "github.com/zackarysantana/bind/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Document struct {
	ID   bson.ObjectID `bson:"_id"`
	Name string        `bson:"name"`
	Age  int           `bson:"age"`
}

type FromNameAndAge struct {
	SomeName string
	SomeAge  int

	Data Document `mongodb:"name=SomeName,age=SomeAge"`
}

type FromStatic struct {
	UsingString  Document `exact-mongodb:"name=Zack"`
	UsingString2 Document `exact-mongodb:"name=Not Zack"`
	UsingInt     Document `exact-mongodb:"age=Int(24)"`
	UsingInt2    Document `exact-mongodb:"age=Int(25)"`
}

func setupDB(t *testing.T) *mongo.Collection {
	mongoC, err := mongodb.Run(t.Context(), "mongo:8")
	require.NoError(t, err)
	t.Cleanup(func() { _ = mongoC.Terminate(t.Context()) })

	uri, err := mongoC.ConnectionString(t.Context())
	require.NoError(t, err)

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err)

	coll := client.Database("test").Collection("users")

	_, err = coll.InsertMany(t.Context(), []any{
		bson.M{"name": "Zack", "age": 24},
		bson.M{"name": "Not Zack", "age": 25},
	})
	require.NoError(t, err)

	return coll
}

func TestMongoDBSupplier(t *testing.T) {
	coll := setupDB(t)

	fromNameAndAge := FromNameAndAge{SomeName: "Not Zack", SomeAge: 25}

	sup, err := mmongodb.NewSupplier[Document](coll, &fromNameAndAge)
	require.NoError(t, err)
	require.NoError(t, bind.Bind(t.Context(), &fromNameAndAge, []bind.Supplier{sup}))
	assert.Equal(t, "Not Zack", fromNameAndAge.Data.Name)
	assert.Equal(t, 25, fromNameAndAge.Data.Age)
}

func TestMongoDBExactSupplier(t *testing.T) {
	coll := setupDB(t)

	fromStatic := FromStatic{}

	sup := mmongodb.NewExactSupplier[Document](coll)
	require.NoError(t, bind.Bind(t.Context(), &fromStatic, []bind.Supplier{sup}))

	t.Run("ByString", func(t *testing.T) {
		assert.Equal(t, "Zack", fromStatic.UsingString.Name)
		assert.Equal(t, 24, fromStatic.UsingString.Age)

		assert.Equal(t, "Not Zack", fromStatic.UsingString2.Name)
		assert.Equal(t, 25, fromStatic.UsingString2.Age)
	})

	t.Run("ByInt", func(t *testing.T) {
		assert.Equal(t, "Zack", fromStatic.UsingInt.Name)
		assert.Equal(t, 24, fromStatic.UsingInt.Age)

		assert.Equal(t, "Not Zack", fromStatic.UsingInt2.Name)
		assert.Equal(t, 25, fromStatic.UsingInt2.Age)
	})
}
