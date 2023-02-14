package param

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type myString string

type myComplicatedType struct {
	Value string
}

func (m *myComplicatedType) UnmarshalText(text []byte) error {
	// differ from simple assignment to underlying (string) type to be sure this was called
	m.Value = "my" + string(text)
	return nil
}

type structWithSlice struct {
	SlicePrimitiveField         []string            `queryparam:"a"`
	SliceCustomField            []myString          `queryparam:"b"`
	SliceCustomUnmarshalerField []myComplicatedType `queryparam:"c"`
	OtherField                  string              `queryparam:"d"`
}

func TestParser_Parse_QueryParam_Slice(t *testing.T) {
	testCases := []struct {
		name     string
		query    string
		expected structWithSlice
	}{
		{
			name:  "multiple items",
			query: "https://test.com/hello?a=vala1&a=vala2&b=valb1&b=valb2&c=valc1&c=valc2&d=vald",
			expected: structWithSlice{
				SlicePrimitiveField:         []string{"vala1", "vala2"},
				SliceCustomField:            []myString{"valb1", "valb2"},
				SliceCustomUnmarshalerField: []myComplicatedType{{"myvalc1"}, {"myvalc2"}},
				OtherField:                  "vald",
			},
		},
		{
			name:  "single item",
			query: "https://test.com/hello?a=vala1&b=valb1&c=valc1&d=vald",
			expected: structWithSlice{
				SlicePrimitiveField:         []string{"vala1"},
				SliceCustomField:            []myString{"valb1"},
				SliceCustomUnmarshalerField: []myComplicatedType{{"myvalc1"}},
				OtherField:                  "vald",
			},
		},
		{
			name:  "no items",
			query: "https://test.com/hello?something_else=hmm",
			expected: structWithSlice{
				SlicePrimitiveField:         nil,
				SliceCustomField:            nil,
				SliceCustomUnmarshalerField: nil,
				OtherField:                  "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := DefaultParser()
			result := structWithSlice{
				SlicePrimitiveField: []string{"existing data should be overwritten in all cases"},
				OtherField:          "in all tagged fields",
			}
			req := httptest.NewRequest(http.MethodGet, tc.query, nil)
			err := parser.Parse(req, &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

type structWithPrimitiveTypes struct {
	Bool    bool    `queryparam:"b"`
	Int     int     `queryparam:"i0"`
	Int8    int8    `queryparam:"i1"`
	Int16   int16   `queryparam:"i2"`
	Int32   int32   `queryparam:"i3"`
	Int64   int64   `queryparam:"i4"`
	Uint    uint    `queryparam:"u0"`
	Uint8   uint8   `queryparam:"u1"`
	Uint16  uint16  `queryparam:"u2"`
	Uint32  uint32  `queryparam:"u3"`
	Uint64  uint64  `queryparam:"u4"`
	Float32 float32 `queryparam:"f1"`
	Float64 float64 `queryparam:"f2"`
	String  string  `queryparam:"s"`
	// nolint:unused
	ignoredUnexported string `queryparam:"ignored"`
}

func TestParser_Parse_QueryParam_PrimitiveTypes(t *testing.T) {
	query := "https://test.com/hello?b=true&i0=-32768&i1=-127&i2=-32768&i3=-2147483648&i4=-9223372036854775808&u0=65535&u1=255&u2=65535&u3=4294967295&u4=18446744073709551615&f1=3e38&f2=1e308&s=hello%20world%5C\"&ignored=hello"
	expected := structWithPrimitiveTypes{
		Bool: true,
		// chosen edge of range numbers most that are most likely to cause problems
		Int:     -32768, // assumes it's at least 16 bits :)
		Int8:    -127,
		Int16:   -32768,
		Int32:   -2147483648,
		Int64:   -9223372036854775808,
		Uint:    65535,
		Uint8:   255,
		Uint16:  65535,
		Uint32:  4294967295,
		Uint64:  18446744073709551615,
		Float32: 3e38,
		Float64: 1e308,
		String:  "hello world\\\"",
	}

	parser := DefaultParser()
	result := structWithPrimitiveTypes{}
	req := httptest.NewRequest(http.MethodGet, query, nil)
	err := parser.Parse(req, &result)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

type structWithPointers struct {
	BoolPtr        *bool              `queryparam:"b"`
	IntPtr         *int               `queryparam:"i"`
	StrPtr         *string            `queryparam:"s"`
	Str2Ptr        **string           `queryparam:"sp"`
	UnmarshalerPtr *myComplicatedType `queryparam:"c"`
}

func TestParser_Parse_QueryParam_Pointers(t *testing.T) {
	testCases := []struct {
		name     string
		query    string
		expected structWithPointers
	}{
		{
			name:  "filled",
			query: "https://test.com/hello?b=true&i=42&s=somestring&sp=pointers&c=wow",
			expected: structWithPointers{
				BoolPtr:        ptr(true),
				IntPtr:         ptr(42),
				StrPtr:         ptr("somestring"),
				Str2Ptr:        ptr(ptr("pointers")),
				UnmarshalerPtr: &myComplicatedType{"mywow"},
			},
		},
		{
			name:  "no params",
			query: "https://test.com/hello",
			expected: structWithPointers{
				BoolPtr:        nil,
				IntPtr:         nil,
				StrPtr:         nil,
				Str2Ptr:        nil,
				UnmarshalerPtr: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := DefaultParser()
			result := structWithPointers{}
			req := httptest.NewRequest(http.MethodGet, tc.query, nil)
			err := parser.Parse(req, &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

type valueReceiverUnmarshaler struct{}

var valueReceiverResult string

func (s valueReceiverUnmarshaler) UnmarshalText(bytes []byte) error {
	valueReceiverResult = string(bytes)
	return nil
}

type StructWithValueReceiverUnmarshal struct {
	Data valueReceiverUnmarshaler `queryparam:"s"`
}

func TestParser_Parse_QueryParam_ValueReceiverUnmarshaler(t *testing.T) {
	query := "https://test.com/hello?s=changed"
	valueReceiverResult = "orig"
	parser := DefaultParser()
	theStruct := StructWithValueReceiverUnmarshal{
		valueReceiverUnmarshaler{},
	}
	req := httptest.NewRequest(http.MethodGet, query, nil)
	err := parser.Parse(req, &theStruct)
	assert.NoError(t, err)
	assert.Equal(t, "changed", valueReceiverResult)
}

func TestParser_Parse_QueryParam_MultipleToNonSlice(t *testing.T) {
	testCases := []struct {
		name         string
		query        string
		resultStruct any
	}{
		{
			name:         "primitive type",
			query:        "https://test.com/hello?b=true&b=true",
			resultStruct: &structWithPrimitiveTypes{},
		},
		{
			name:         "text unmarshaler",
			query:        "https://test.com/hello?c=yes&c=no",
			resultStruct: &structWithPointers{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := DefaultParser()
			req := httptest.NewRequest(http.MethodGet, tc.query, nil)
			err := parser.Parse(req, tc.resultStruct)
			assert.Error(t, err)
		})
	}
}

func TestParser_Parse_QueryParam_InvalidType(t *testing.T) {
	var str string
	testCases := []struct {
		name         string
		query        string
		resultStruct any
	}{
		{
			name:         "not a pointer",
			query:        "https://test.com/hello?b=true",
			resultStruct: structWithPrimitiveTypes{},
		},
		{
			name:         "pointer to not struct",
			query:        "https://test.com/hello",
			resultStruct: &str,
		},
		{
			name:  "map",
			query: "https://test.com/hello?map=something",
			resultStruct: &struct {
				Map map[string]any `queryparam:"map"`
			}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := DefaultParser()
			req := httptest.NewRequest(http.MethodGet, tc.query, nil)
			err := parser.Parse(req, tc.resultStruct)
			assert.Error(t, err)
		})
	}
}

func TestParser_Parse_QueryParam_CannotBeParsed(t *testing.T) {
	testCases := []struct {
		name         string
		query        string
		resultStruct any
		errorTarget  error
	}{
		{
			name:         "invalid bool",
			query:        "https://test.com/hello?b=frue",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrSyntax,
		},
		{
			name:         "invalid int",
			query:        "https://test.com/hello?i0=18446744073709551615",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrRange,
		},
		{
			name:         "invalid int8",
			query:        "https://test.com/hello?i1=128",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrRange,
		},
		{
			name:         "invalid int16",
			query:        "https://test.com/hello?i2=32768",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrRange,
		},
		{
			name:         "invalid int32",
			query:        "https://test.com/hello?i3=2147483648",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrRange,
		},
		{
			name:         "invalid int64",
			query:        "https://test.com/hello?i4=18446744073709551615",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrRange,
		},
		{
			name:         "invalid uint",
			query:        "https://test.com/hello?u0=-1",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrSyntax,
		},
		{
			name:         "invalid uint8",
			query:        "https://test.com/hello?u1=-1",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrSyntax,
		},
		{
			name:         "invalid uint16",
			query:        "https://test.com/hello?u2=-1",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrSyntax,
		},
		{
			name:         "invalid uint32",
			query:        "https://test.com/hello?u3=-1",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrSyntax,
		},
		{
			name:         "invalid uint64",
			query:        "https://test.com/hello?u4=-1",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrSyntax,
		},
		{
			name:         "invalid float32",
			query:        "https://test.com/hello?f1=4e38",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrRange,
		},
		{
			name:         "invalid float64",
			query:        "https://test.com/hello?f2=1e309",
			resultStruct: &structWithPrimitiveTypes{},
			errorTarget:  strconv.ErrRange,
		},
		{
			name:  "invalid int8 in slice",
			query: "https://test.com/hello?x=127&x=128",
			resultStruct: &struct {
				Slice []int8 `queryparam:"x"`
			}{},
			errorTarget: strconv.ErrRange,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := DefaultParser()
			req := httptest.NewRequest(http.MethodGet, tc.query, nil)
			err := parser.Parse(req, tc.resultStruct)
			assert.ErrorIs(t, err, tc.errorTarget)
		})
	}
}

type maybeShinyObject struct {
	IsShiny bool
	Object  string
}

func (m *maybeShinyObject) UnmarshalText(text []byte) error {
	if strings.HasPrefix(string(text), "shiny-") {
		m.IsShiny = true
		m.Object = string(text[6:])
		return nil
	}
	m.Object = string(text)
	return nil
}

type structWithPathParams struct {
	Subject string            `pathparam:"subject"`
	Amount  *int              `pathparam:"amount"`
	Object  *maybeShinyObject `pathparam:"object"`
	Nothing string            `pathparam:"nothing"`
}

func TestParser_Parse_PathParam(t *testing.T) {
	r := chi.NewRouter()
	p := DefaultParser().WithPathParamFunc(chi.URLParam)
	result := structWithPathParams{Nothing: "should be replaced"}
	expected := structWithPathParams{
		Subject: "world",
		Amount:  ptr(69),
		Object: &maybeShinyObject{
			IsShiny: true,
			Object:  "apples",
		},
		Nothing: "",
	}
	var parseError error
	r.Get("/hello/{subject}/i/have/{amount}/{object}", func(w http.ResponseWriter, r *http.Request) {
		parseError = p.Parse(r, &result)
	})

	req := httptest.NewRequest(http.MethodGet, "https://test.com/hello/world/i/have/69/shiny-apples", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	assert.NoError(t, parseError)
	assert.Equal(t, expected, result)
}

type simpleStringPathParamStruct struct {
	Param int `pathparam:"param"`
}

func TestParser_Parse_PathParam_ParseError(t *testing.T) {
	r := chi.NewRouter()
	p := DefaultParser().WithPathParamFunc(chi.URLParam)
	var parseError error
	r.Get("/hello/{param}", func(w http.ResponseWriter, r *http.Request) {
		parseError = p.Parse(r, &simpleStringPathParamStruct{})
	})

	req := httptest.NewRequest(http.MethodGet, "https://test.com/hello/not-a-number", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	assert.Error(t, parseError)
}

func TestParser_Parse_PathParam_FuncNotDefinedError(t *testing.T) {
	p := DefaultParser()
	req := httptest.NewRequest(http.MethodGet, "https://test.com/hello/not-a-number", nil)

	err := p.Parse(req, &simpleStringPathParamStruct{})

	assert.Error(t, err)
}

func ptr[T any](x T) *T {
	return &x
}
