# bind

[![Go Reference](https://pkg.go.dev/badge/github.com/ZackarySantana/bind.svg)](https://pkg.go.dev/github.com/ZackarySantana/bind)
[![Go Report Card](https://goreportcard.com/badge/github.com/ZackarySantana/bind)](https://goreportcard.com/report/github.com/ZackarySantana/bind)

A flexible binding library for Go that maps external values (YAML, JSON, CLI args, environment, HTTP request paths, etc.) into struct fields via struct tags.

## Table of Contents

-   [Overview](#overview)
-   [Installation](#installation)
-   [Usage](#usage)
    -   [Options](#options)
    -   [Options Struct Tag](#options-struct-tag)
-   [Suppliers](#suppliers)
    -   [JSONSupplier](#jsonsupplier)
    -   [YAMLSupplier](#yamlsupplier)
    -   [EnvSupplier](#envsupplier)
    -   [SelfSupplier](#selfsupplier)
    -   [Other Suppliers](#other-suppliers)
-   [Testing](#testing)
-   [Contributing](#contributing)
-   [License](#license)

## Overview

`bind` provides a simple way to populate Go structs from various external sources using struct tags. It supports multiple suppliers that can be combined to fill in struct fields from different sources.

## Installation

```bash
go get github.com/ZackarySantana/bind
```

## Usage

```go
type Config struct {
    Port int    `json:"port"`
    Host string `yaml:"host"`
    DB   string `env:"DB_URL"`
}

yaml := []byte(`host: localhost`)
yamlSup, _ := bind.NewYAMLSupplier(bytes.NewReader(yaml))

json := []byte(`{"port":8080}`)
jsonSup, _ := bind.NewJSONSupplier(json)

os.Setenv("DB_URL", "postgres://user:pass@localhost/db")

var cfg Config
bind.Bind(ctx, &cfg, []bind.Supplier{yamlSup, jsonSup, bind.NewEnvSupplier()})
```

If the target struct already has values, they will not be overwritten. This means calls to `bind.Bind` will only fill in missing values.

### Options

#### WithLogger

Bind outputs debug information using the provided `slog.Logger`. If not provided, no logging is done.

**Example:**

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

bind.Bind(ctx, &var, suppliers, bind.WithLogger(logger))
```

#### WithLevel

Bind only sets fields with a level less than or equal to the provided level. Default is `1`. For more information, see [Options Struct Tag](#options-struct-tag). This is to support multiple levels of configuration (e.g. the first level is non-auth fields, the second level is auth-required fields).

**Example:**

```go
var test struct {
    Retries  int    `json:"retries"`
    Name     string `json:"name" options:"level=1"`
    DBURL    string `json:"db_url" options:"level=2"`
}

// Only the Retries and Name fields will be set.
bind.Bind(ctx, &var, suppliers)

// DBURL will also be set.
bind.Bind(ctx, &var, suppliers, bind.WithLevel(2))
```

## Options Struct Tag

The `options` struct tag allows you to specify additional options for each field.

### Required

**Example:**

```go
var test struct {
    Name string `json:"name" options:"required"`
    Age  int    `json:"age"`
}

jsonSup, _ := bind.NewJSONSupplier(strings.NewReader(`{"age":30}`))

// This will return an error because Name is required but not provided.
err := bind.Bind(ctx, &test, []bind.Supplier{jsonSup})
```

### Level

**Example:**

```go
var test struct {
    Retries  int    `json:"retries"`
    Name     string `json:"name" options:"level=1"`
    DBURL    string `json:"db_url" options:"level=2"`
}

jsonSup, _ := bind.NewJSONSupplier(strings.NewReader(`{"retries":3,"name":"Alice","db_url":"postgres://user:pass@localhost/db"}`))

// Only the Retries and Name fields will be set.
bind.Bind(ctx, &test, []bind.Supplier{jsonSup})
```

## Suppliers

### JSONSupplier

Parses raw JSON into a `map[string]json.RawMessage` and extracts values by `json` tags.

#### Usage

```go
func NewJSONSupplier(src io.Reader) (*JSONSupplier, error)
```

**Example:**

```go
jsonData := `{"name":"Alice","age":30}`
sup, _ := bind.NewJSONSupplier(strings.NewReader(jsonData))

var age int
sup.Fill(ctx, "age", nil, &age)

// Or bind directly to a struct:
var test struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
bind.Bind(ctx, &test, []bind.Supplier{sup})
```

### YAMLSupplier

Parses raw YAML into a `map[string]yaml.Node` and extracts values by `yaml` tags.

```go
func NewYAMLSupplier(r io.Reader) (*YAMLSupplier, error)
```

**Example:**

```go
yamlData := `name: Bob
age: 42`
sup, _ := bind.NewYAMLSupplier(strings.NewReader(yamlData))

var age int
sup.Fill(ctx, "age", nil, &age)

// Or bind directly to a struct:
var test struct {
    Name string `yaml:"name"`
    Age  int    `yaml:"age"`
}
bind.Bind(ctx, &test, []bind.Supplier{sup})
```

### EnvSupplier

Looks up environment variables based on `env:"..."` tags.

**Example:**

```go
os.Setenv("PORT", "9090")
sup := bind.NewEnvSupplier()

var port int
sup.Fill(ctx, "PORT", nil, &port)

// Or bind directly to a struct:
var test struct {
    Port int `env:"PORT"`
}
bind.Bind(ctx, &test, []bind.Supplier{sup})
```

### SelfSupplier

The SelfSupplier is used to populate fields in a struct based on values from the struct itself. This is particularly useful for scenarios where you want to use certain fields as keys to look up additional data from a store or database.

**Example:**

```go
type User struct {
    ID    int
    Phone string
    Name  string `test:"id=ID"`
    Age   int    `test2:"num=Phone,other=ID"`
}

u := User{
    ID:    9001,
    Phone: "970-4133",
}

testSup, _ := bind.NewSelfSupplier(func(ctx context.Context, filter map[string]any) (string, error) {
    // filter == map[string]any{"id": 9001}
    // Notice how the filter includes the value of ID from the struct
    return "found!", nil
}, "test", &u)

test2Sup, _ := bind.NewSelfSupplier(func(ctx context.Context, filter map[string]any) (int, error) {
    // filter == map[string]any{"num": "970-4133", "other": 9001}
    // Notice how the filter includes the values of Phone and ID from the struct
    return 42, nil
}, "test2", &u)

bind.Bind(ctx, &u, []bind.Supplier{testSup, test2Sup})

// u.Name == "found!"
// u.Age  == 42
```

### Other Suppliers

-   PathSupplier: Extracts values from HTTP request paths via `req.PathValue`. Using `path:"..."` tags.
-   QuerySupplier: Extracts values from URL query parameters using `query:"..."` tags.
-   HeaderSupplier: Extracts values from HTTP headers using `header:"..."` tags.
-   FormSupplier: Extracts values from form data using `form:"..."` tags.
-   RequestSuppliers: From a given `*http.Request`, creates a PathSupplier, QuerySupplier, HeaderSupplier and, FormSupplier.
-   FlagSupplier: Binds values from CLI flags using `flag:"..."` tags.
-   FuncSupplier: Uses a user-defined function to supply values based on a given tag.
-   FuncStringSupplier: Uses a user-defined function that returns strings to supply values based on a given tag. The strings are then attempted to be converted to the target field type.

## Testing

Run all tests with:

```bash
go test ./...
```

## Contributing

PRs and issues are welcome!
If you add a new supplier, please include:

-   Unit tests
-   Example usage in the README
-   Documentation comments

## License

MIT License. See [LICENSE](./LICENSE) for details.
