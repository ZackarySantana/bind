package valkey_test

import (
	"net/url"
	"strconv"
	"testing"

	valkey "github.com/zackarysantana/bind/modules/valkey"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	containerValkey "github.com/testcontainers/testcontainers-go/modules/valkey"
	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/config"
	"github.com/zackarysantana/bind"
)

type Document struct {
	ID   string `bson:"_id"`
	Name string `bson:"name"`
	Age  int    `bson:"age"`
}

type FromNameAndAge struct {
	SomeName string
	SomeAge  int

	Data Document `valkey:"name=SomeName,age=SomeAge"`
}

type FromStatic struct {
	UsingString  Document `exact-valkey:"Zackary"`
	UsingString2 Document `exact-valkey:"NotZackary"`
	UsingInt     Document `exact-valkey:"age=Int(24)"`
	UsingInt2    Document `exact-valkey:"age=Int(25)"`
}

func setupDB(t *testing.T) *glide.Client {
	// mongoC, err := mongodb.Run(t.Context(), "mongo:8")
	// require.NoError(t, err)
	// t.Cleanup(func() { _ = mongoC.Terminate(t.Context()) })

	// uri, err := mongoC.ConnectionString(t.Context())
	// require.NoError(t, err)

	// client, err := mongo.Connect(options.Client().ApplyURI(uri))
	// require.NoError(t, err)

	// coll := client.Database("test").Collection("users")

	// _, err = coll.InsertMany(t.Context(), []any{
	// 	bson.M{"name": "Zack", "age": 24},
	// 	bson.M{"name": "Not Zack", "age": 25},
	// })
	// require.NoError(t, err)

	// return coll
	valkeyContainer, err := containerValkey.Run(t.Context(), "valkey/valkey:8.1.3")
	require.NoError(t, err)
	t.Cleanup(func() { _ = valkeyContainer.Terminate(t.Context()) })

	// create glide client
	endpoint, err := valkeyContainer.ConnectionString(t.Context())
	require.NoError(t, err)

	valkeyURL, err := url.Parse(endpoint)
	require.NoError(t, err)
	valkeyPort := 6379
	if port := valkeyURL.Port(); port != "" {
		var err error
		valkeyPort, err = strconv.Atoi(port)
		require.NoError(t, err)
	}

	client, err := glide.NewClient(config.NewClientConfiguration().WithAddress(&config.NodeAddress{Host: valkeyURL.Hostname(), Port: valkeyPort}))
	require.NoError(t, err)

	client.Set(t.Context(), "Zackary", "test")

	client.Get(t.Context(), "Zackary")

	return client
}

// func TestValkeyDBSupplier(t *testing.T) {
// 	coll := setupDB(t)

// 	fromNameAndAge := FromNameAndAge{SomeName: "Not Zack", SomeAge: 25}

// 	sup, err := valkey.NewExactSupplier[Document](coll, &fromNameAndAge)
// 	require.NoError(t, err)
// 	require.NoError(t, bind.Bind(t.Context(), &fromNameAndAge, []bind.Supplier{sup}))
// 	assert.Equal(t, "Not Zack", fromNameAndAge.Data.Name)
// 	assert.Equal(t, 25, fromNameAndAge.Data.Age)
// }

func TestMongoDBExactSupplier(t *testing.T) {
	coll := setupDB(t)

	fromStatic := FromStatic{}

	sup := valkey.NewExactSupplier[Document](coll)
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
