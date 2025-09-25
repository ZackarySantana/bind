package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildFilter(t *testing.T) {
	options := []string{
		"no_type=hello",
		"string=String(world)",
		"string_slice=StringSlice(a,b,c)",
		"int=Int(42)",
		"int_slice=IntSlice(1,2,3)",
		"int32=Int32(32)",
		"int32_slice=Int32Slice(4,5,6)",
		"int64=Int64(64)",
		"int64_slice=Int64Slice(7,8,9)",
		"float64=Float64(3.14)",
		"float64_slice=Float64Slice(1.1,2.2,3.3)",
		"bool=Bool(true)",
		"bool_slice=BoolSlice(true,false,true)",
		"time=Time(2002-01-01T01:00:00Z)",
		"time_slice=TimeSlice(2002-01-01T01:00:00Z,2000-01-01T00:00:00Z)",
	}

	filter, err := BuildFilter(options)
	require.NoError(t, err)

	assert.Equal(t, "hello", filter["no_type"])
	assert.Equal(t, "world", filter["string"])
	assert.Equal(t, []string{"a", "b", "c"}, filter["string_slice"])
	assert.Equal(t, 42, filter["int"])
	assert.Equal(t, []int{1, 2, 3}, filter["int_slice"])
	assert.Equal(t, int32(32), filter["int32"])
	assert.Equal(t, []int32{4, 5, 6}, filter["int32_slice"])
	assert.Equal(t, int64(64), filter["int64"])
	assert.Equal(t, []int64{7, 8, 9}, filter["int64_slice"])
	assert.Equal(t, 3.14, filter["float64"])
	assert.Equal(t, []float64{1.1, 2.2, 3.3}, filter["float64_slice"])
	assert.Equal(t, true, filter["bool"])
	assert.Equal(t, []bool{true, false, true}, filter["bool_slice"])
	assert.Equal(t, "2002-01-01 01:00:00 +0000 UTC", filter["time"].(interface{ String() string }).String())

	actualTimes := filter["time_slice"].([]time.Time)
	actualTimesAsString := make([]string, len(actualTimes))
	for i, tm := range actualTimes {
		actualTimesAsString[i] = tm.String()
	}
	assert.Equal(t, []string{"2002-01-01 01:00:00 +0000 UTC", "2000-01-01 00:00:00 +0000 UTC"}, actualTimesAsString)
}

func TestRegister(t *testing.T) {
	Register("Custom", func(s string) (string, error) {
		return "custom:" + s, nil
	})
	type myStruct struct {
		Value string
	}
	Register("RegisteredStruct", func(s string) (myStruct, error) {
		return myStruct{Value: s}, nil
	})

	options := []string{
		"custom=Custom(myvalue)",
		"custom_slice=CustomSlice(a,b,c)",
		"registered=RegisteredStruct(somevalue)",
		"registered_slice=RegisteredStructSlice(one,two,three)",
	}

	filter, err := BuildFilter(options)
	require.NoError(t, err)

	assert.Equal(t, "custom:myvalue", filter["custom"])
	assert.Equal(t, []string{"custom:a", "custom:b", "custom:c"}, filter["custom_slice"])
	assert.Equal(t, myStruct{Value: "somevalue"}, filter["registered"])
	assert.Equal(t, []myStruct{
		{Value: "one"},
		{Value: "two"},
		{Value: "three"},
	}, filter["registered_slice"])
}
