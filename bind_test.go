package bind

import (
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nested struct {
	Code int `json:"code"`
}

// TODO: This should test all created suppliers e2e.
type mixed struct {
	Int     int            `json:"int,omitempty"`
	Float64 float64        `json:"float"`
	Bool    bool           `json:"bool"`
	Float32 float32        `json:"-"` // should remain zero
	StrPtr  *string        `json:"str_ptr"`
	TimeVal time.Time      `json:"time"`
	Tags    []string       `json:"tags"`
	Meta    map[string]int `json:"meta"`
	Inner   nested         `json:"inner"`
	InnerP  *nested        `json:"inner_p"`
	ResID   int            `path:"res_id"`
	NoTag   string         // must remain zero
	Path    string         `yaml:"org_id"` // no supplier present
}

type required struct {
	Must   string `json:"must" options:"required"`
	Mustnt string `json:"mustnt"`
}

type level struct {
	Lvl2 string `json:"lvl2" options:"level=2"`
	Lvl5 string `json:"lvl5" options:"level=5"`
}

type requiredAndLevel struct {
	Lvl2 string `json:"lvl2" options:"level=2,required"`
	Lvl5 string `json:"lvl5" options:"required,level=5"`
}

func TestBind(t *testing.T) {
	pathSup := func(t *testing.T, pathValues map[string]string) Supplier {
		req := &http.Request{}
		for key, val := range pathValues {
			req.SetPathValue(key, val)
		}
		ps, err := NewPathSupplier(req)
		require.NoError(t, err)
		return ps
	}
	s := "hello"

	for name, tc := range map[string]struct {
		suppliers           []Supplier
		level               float64
		destination         any
		expected            any
		expectedErr         error
		expectedErrContains string
	}{
		"Error/NilDestination": {
			expectedErr: ErrNilDestination,
		},
		"Error/NonPointerDestination": {
			destination: mixed{},
			expectedErr: ErrNonPointer,
		},
		"Error/NonStructPointerDestination": {
			destination: new(int),
			expectedErr: ErrNonStructPtr,
		},
		"NoSuppliersFillsNothing": {
			destination: &mixed{},
			expected:    &mixed{},
		},
		"RequiredFieldMissing": {
			suppliers:           []Supplier{createJSONSupplier(t, `{"mustnt":"present"}`)},
			destination:         &required{},
			expectedErrContains: "required field 'Must' (type 'string') not found",
		},
		"RequiredFieldFilled": {
			suppliers:   []Supplier{createJSONSupplier(t, `{"must":"present"}`)},
			destination: &required{},
			expected:    &required{Must: "present"},
		},
		"HigherLevelsNotFilled": {
			suppliers:   []Supplier{createJSONSupplier(t, `{"lvl2":"present","lvl5":"present"}`)},
			destination: &level{},
			expected:    &level{},
		},
		"LowerLevelFilledHigherLevelNotFilled": {
			suppliers:   []Supplier{createJSONSupplier(t, `{"lvl2":"present","lvl5":"present"}`)},
			level:       2,
			destination: &level{},
			expected: &level{
				Lvl2: "present",
			},
		},
		"AllLevelsFilled": {
			suppliers:   []Supplier{createJSONSupplier(t, `{"lvl2":"present","lvl5":"present"}`)},
			level:       5,
			destination: &level{},
			expected: &level{
				Lvl2: "present",
				Lvl5: "present",
			},
		},
		"RequiredAndLevel/HigherLevelsNotFilled": {
			suppliers:   []Supplier{createJSONSupplier(t, `{"lvl2":"present","lvl5":"present"}`)},
			destination: &requiredAndLevel{},
			expected:    &requiredAndLevel{},
		},
		"RequiredAndLevel/HighestLevelNotFilled": {
			suppliers:   []Supplier{createJSONSupplier(t, `{"lvl2":"present","lvl5":"present"}`)},
			level:       2,
			destination: &requiredAndLevel{},
			expected: &requiredAndLevel{
				Lvl2: "present",
			},
		},
		"RequiredAndLevel/AllLevelsFilled": {
			suppliers:   []Supplier{createJSONSupplier(t, `{"lvl2":"present","lvl5":"present"}`)},
			level:       5,
			destination: &requiredAndLevel{},
			expected: &requiredAndLevel{
				Lvl2: "present",
				Lvl5: "present",
			},
		},
		"RequiredAndLevel/LevelAllowedButNotFilledError": {
			suppliers:           []Supplier{createJSONSupplier(t, `{}`)},
			level:               2,
			destination:         &requiredAndLevel{},
			expectedErrContains: "required field 'Lvl2' (type 'string') not found",
		},
		"IgnoresAlreadySetFields": {
			suppliers: []Supplier{createJSONSupplier(t, `{"int":7,"float":9.42}`)},
			destination: &mixed{
				Int: 3,
			},
			expected: &mixed{
				Int:     3,
				Float64: 9.42,
			},
		},
		"MultipleSuppliers": {
			suppliers: []Supplier{
				createJSONSupplier(t, `{
					"int": 7,
					"float": 3.14,
					"-": 6.28,
					"bool": true,
					"str_ptr": "hello",
					"time": "2025-01-02T03:04:05Z",
					"tags": ["a","b"],
					"meta": {"x":1,"y":2},
					"inner": {"code": 200},
					"inner_p": {"code": 201},
					"no_tag": "ignored",
					"NoTag": "ignored",
					"path": "ignored"
				}`),
				pathSup(t, map[string]string{"res_id": "42"}),
			},
			destination: &mixed{},
			expected: &mixed{
				Int:     7,
				Float64: 3.14,
				Bool:    true,
				StrPtr:  &s,
				TimeVal: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC),
				Tags:    []string{"a", "b"},
				Meta:    map[string]int{"x": 1, "y": 2},
				Inner:   nested{Code: 200},
				InnerP:  &nested{Code: 201},
				ResID:   42,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := Bind(t.Context(), tc.destination, tc.suppliers, WithLevel(int(math.Max(tc.level, 1))))

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr, err)
				return
			}
			if tc.expectedErrContains != "" {
				require.Error(t, err, tc.destination)
				assert.ErrorContains(t, err, tc.expectedErrContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, tc.destination)
		})
	}
}

// runSupplierTests is a shared test util that suppliers use
// to test themselves.
func runSupplierTests(t *testing.T, s Supplier, emptySupplier Supplier, wantKind string) {
	t.Helper()

	t.Run("Fill", func(t *testing.T) {
		t.Run("Missing", func(t *testing.T) {
			var v string
			ok, err := s.Fill(t.Context(), keyMissing, nil, &v)
			require.NoError(t, err)
			require.False(t, ok)
			assert.Empty(t, v)
		})
		t.Run("Empty", func(t *testing.T) {
			var v string
			ok, err := emptySupplier.Fill(t.Context(), "empty", nil, &v)
			require.NoError(t, err)
			require.False(t, ok)
			assert.Empty(t, v)
		})
		t.Run("String", func(t *testing.T) {
			var v string
			ok, err := s.Fill(t.Context(), keyStr, nil, &v)
			require.NoError(t, err)
			require.True(t, ok)
			assert.Equal(t, valStr, v)
		})
		t.Run("Int", func(t *testing.T) {
			var v int
			ok, err := s.Fill(t.Context(), keyInt, nil, &v)
			require.NoError(t, err)
			require.True(t, ok)
			assert.Equal(t, 123, v)
		})
	})

	t.Run("Kind", func(t *testing.T) {
		assert.True(t, s.IsKind(wantKind))
	})
}

func runNestedSupplierTest(t *testing.T, s Supplier) {
	t.Run("Struct", func(t *testing.T) {
		var v struct {
			Bar string `json:"bar"`
		}
		ok, err := s.Fill(t.Context(), keyNested, nil, &v)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "baz", v.Bar)
	})
}
