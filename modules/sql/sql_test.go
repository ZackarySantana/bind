package sql_test

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"testing"

	_ "modernc.org/sqlite" // sqlite driver

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zackarysantana/bind"
	mSQL "github.com/zackarysantana/bind/modules/sql"
)

type Row struct {
	ID   string `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

type FromID struct {
	SomeID string

	Data       Row `sql:"id=SomeID"`
	DataCapsID Row `sql:"ID=SomeID"`
}

type FromNameAndAge struct {
	SomeName string
	SomeAge  int

	Data     Row `sql:"name=SomeName,age=SomeAge"`
	DataCaps Row `sql:"Name=SomeName,Age=SomeAge"`
}

type FromStatic struct {
	First      Row `exact-sql:"id=first_user"`
	Second     Row `exact-sql:"id=second_user"`
	SecondCaps Row `exact-sql:"ID=second_user"`
	UsingInt   Row `exact-sql:"age=Int(24)"`
	UsingInt2  Row `exact-sql:"age=Int(25)"`
}

// randomName returns a short random hex string.
func randomName() string {
	b := make([]byte, 4) // 8 hex chars
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func setupDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", randomName()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`
		CREATE TABLE users (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			age        INTEGER NOT NULL
		);
		INSERT INTO users (id, name, age) VALUES
			('first_user', 'Zack', 24),
			('second_user', 'Not Zack', 25);
	`)
	require.NoError(t, err)

	return db
}

func TestSQLSupplier(t *testing.T) {
	db := setupDB(t)

	t.Run("FromID", func(t *testing.T) {
		fromID := FromID{SomeID: "first_user"}

		sup, err := mSQL.NewSupplier[Row](db, "users", &fromID)
		require.NoError(t, err)
		require.NoError(t, bind.Bind(t.Context(), &fromID, []bind.Supplier{sup}))
		assert.Equal(t, "first_user", fromID.Data.ID)
		assert.Equal(t, "Zack", fromID.Data.Name)
		assert.Equal(t, 24, fromID.Data.Age)

		assert.Equal(t, "first_user", fromID.DataCapsID.ID)
		assert.Equal(t, "Zack", fromID.DataCapsID.Name)
		assert.Equal(t, 24, fromID.DataCapsID.Age)
	})

	t.Run("FromNameAndAge", func(t *testing.T) {
		fromNameAndAge := FromNameAndAge{SomeName: "Not Zack", SomeAge: 25}

		sup, err := mSQL.NewSupplier[Row](db, "users", &fromNameAndAge)
		require.NoError(t, err)
		require.NoError(t, bind.Bind(t.Context(), &fromNameAndAge, []bind.Supplier{sup}))
		assert.Equal(t, "second_user", fromNameAndAge.Data.ID)
		assert.Equal(t, "Not Zack", fromNameAndAge.Data.Name)
		assert.Equal(t, 25, fromNameAndAge.Data.Age)

		assert.Equal(t, "second_user", fromNameAndAge.DataCaps.ID)
		assert.Equal(t, "Not Zack", fromNameAndAge.DataCaps.Name)
		assert.Equal(t, 25, fromNameAndAge.DataCaps.Age)
	})

}

func TestSQLExactSupplier(t *testing.T) {
	db := setupDB(t)

	fromStatic := FromStatic{}

	sup := mSQL.NewExactSupplier[Row](db, "users")
	require.NoError(t, bind.Bind(t.Context(), &fromStatic, []bind.Supplier{sup}))

	t.Run("ByStrings", func(t *testing.T) {
		assert.Equal(t, "first_user", fromStatic.First.ID)
		assert.Equal(t, "Zack", fromStatic.First.Name)
		assert.Equal(t, 24, fromStatic.First.Age)

		assert.Equal(t, "second_user", fromStatic.Second.ID)
		assert.Equal(t, "Not Zack", fromStatic.Second.Name)
		assert.Equal(t, 25, fromStatic.Second.Age)

		assert.Equal(t, "second_user", fromStatic.SecondCaps.ID)
		assert.Equal(t, "Not Zack", fromStatic.SecondCaps.Name)
		assert.Equal(t, 25, fromStatic.SecondCaps.Age)
	})

	t.Run("ByInt", func(t *testing.T) {
		assert.Equal(t, "first_user", fromStatic.UsingInt.ID)
		assert.Equal(t, "Zack", fromStatic.UsingInt.Name)
		assert.Equal(t, 24, fromStatic.UsingInt.Age)

		assert.Equal(t, "second_user", fromStatic.UsingInt2.ID)
		assert.Equal(t, "Not Zack", fromStatic.UsingInt2.Name)
		assert.Equal(t, 25, fromStatic.UsingInt2.Age)
	})
}
