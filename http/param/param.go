package param

import (
	"encoding"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

// TagResolver is a function that decides from a field type what key of http parameter should be searched.
// Second return value should return whether the key should be searched in http parameter at all.
type TagResolver func(fieldTag reflect.StructTag) (string, bool)

// FixedTagNameParamTagResolver returns a TagResolver, that matches struct params by specific tag.
// Example: FixedTagNameParamTagResolver("mytag") matches a field tagged with `mytag:"query_param_name"`
func FixedTagNameParamTagResolver(tagName string) TagResolver {
	return func(fieldTag reflect.StructTag) (string, bool) {
		taggedParamName := fieldTag.Get(tagName)
		return taggedParamName, taggedParamName != ""
	}
}

// PathParamFunc is a function that returns value of specified http path parameter
type PathParamFunc func(r *http.Request, key string) string

// Parser can Parse query and path parameters from http.Request into a struct.
// Fields struct have to be tagged such that either QueryParamTagResolver or PathParamTagResolver returns
// valid parameter name from the provided tag.
//
// PathParamFunc is for getting path parameter from http.Request, as each http router handles it in different way (if at all).
// For example for chi, use WithPathParamFunc(chi.URLParam) to be able to use tags for path parameters.
type Parser struct {
	QueryParamTagResolver TagResolver
	PathParamTagResolver  TagResolver
	PathParamFunc         PathParamFunc
}

// DefaultParser returns query and path parameter Parser with intended struct tags
// `queryparam:"name"` for query parameters and `pathparam:"name"` for path parameters
func DefaultParser() Parser {
	return Parser{
		QueryParamTagResolver: FixedTagNameParamTagResolver("queryparam"),
		PathParamTagResolver:  FixedTagNameParamTagResolver("pathparam"),
		PathParamFunc:         nil, // keep nil, as there is no sensible default of how to get value of path parameter
	}
}

// WithPathParamFunc returns a copy of Parser with set function for getting path parameters from http.Request.
// For more see Parser description.
func (p Parser) WithPathParamFunc(f PathParamFunc) Parser {
	p.PathParamFunc = f
	return p
}

// Parse accepts the request and a pointer to struct that is tagged with appropriate tags set in Parser.
// All such tagged fields are assigned the respective parameter from the actual request.
//
// Fields are assigned their zero value if the field was tagged but request did not contain such parameter.
//
// Supported tagged field types are:
// - primitive types - bool, all ints, all uints, both floats, and string
// - pointer to any supported type
// - slice of non-slice supported type (only for query parameters)
// - any type that implements encoding.TextUnmarshaler
//
// For query parameters, the tagged type can be a slice. This means that a query like /endpoint?key=val1&key=val2
// is allowed, and in such case the slice field will be assigned []T{"val1", "val2"} .
// Otherwise, only single query parameter is allowed in request.
func (p Parser) Parse(r *http.Request, dest any) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Pointer {
		return fmt.Errorf("cannot set non-pointer value of type %s", v.Type().Name())
	}
	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("can only parse into struct, but got %s", v.Type().Name())
	}

	for i := 0; i < v.NumField(); i++ {
		typeField := v.Type().Field(i)
		if !typeField.IsExported() {
			continue
		}
		valueField := v.Field(i)
		// Zero the value, even if it would not be set by following path or query parameter.
		// This will cause potential partial result from previous parser (e.g. json.Unmarshal) to be discarded on
		// fields that are tagged for path or query parameter.
		valueField.Set(reflect.Zero(typeField.Type))
		tag := typeField.Tag
		err := p.parseQueryParam(r, tag, valueField)
		if err != nil {
			return err
		}
		err = p.parsePathParam(r, tag, valueField)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p Parser) parsePathParam(r *http.Request, tag reflect.StructTag, v reflect.Value) error {
	if paramName, ok := p.PathParamTagResolver(tag); ok {
		if p.PathParamFunc == nil {
			return fmt.Errorf("struct's field was tagged for parsing the path parameter (%s) but PathParamFunc to get value of path parameter is not defined", paramName)
		}
		paramValue := p.PathParamFunc(r, paramName)
		if paramValue != "" {
			err := unmarshalValue(paramValue, v)
			if err != nil {
				return fmt.Errorf("unmarshaling path parameter %s: %w", paramName, err)
			}
		}
	}
	return nil
}

func (p Parser) parseQueryParam(r *http.Request, tag reflect.StructTag, v reflect.Value) error {
	if paramName, ok := p.QueryParamTagResolver(tag); ok {
		query := r.URL.Query()
		if texts, ok := (map[string][]string)(query)[paramName]; ok && len(texts) > 0 {
			err := unmarshalValueOrSlice(texts, v)
			if err != nil {
				return fmt.Errorf("unmarshaling query parameter %s: %w", paramName, err)
			}
		}
	}
	return nil
}

func unmarshalValueOrSlice(texts []string, dest reflect.Value) error {
	if unmarshaler, ok := dest.Addr().Interface().(encoding.TextUnmarshaler); ok {
		if len(texts) != 1 {
			return fmt.Errorf("too many parameters unmarshaling to %s, expected up to 1 value", dest.Type().Name())
		}
		return unmarshaler.UnmarshalText([]byte(texts[0]))
	}
	t := dest.Type()
	if t.Kind() == reflect.Pointer {
		ptrValue := reflect.New(t.Elem())
		dest.Set(ptrValue)
		return unmarshalValueOrSlice(texts, dest.Elem())
	}
	if t.Kind() == reflect.Slice {
		sliceValue := reflect.MakeSlice(t, len(texts), len(texts))
		for i, text := range texts {
			if err := unmarshalValue(text, sliceValue.Index(i)); err != nil {
				return fmt.Errorf("unmarshaling %dth element: %w", i, err)
			}
		}
		dest.Set(sliceValue)
		return nil
	}
	if len(texts) != 1 {
		return fmt.Errorf("too many parameters unmarshaling to %s, expected up to 1 value", dest.Type().Name())
	}
	return unmarshalPrimitiveValue(texts[0], dest)
}

func unmarshalValue(text string, dest reflect.Value) error {
	if unmarshaler, ok := dest.Addr().Interface().(encoding.TextUnmarshaler); ok {
		return unmarshaler.UnmarshalText([]byte(text))
	}
	t := dest.Type()
	if t.Kind() == reflect.Pointer {
		ptrValue := reflect.New(t.Elem())
		dest.Set(ptrValue)
		return unmarshalValue(text, dest.Elem())
	}
	return unmarshalPrimitiveValue(text, dest)
}

func unmarshalPrimitiveValue(text string, dest reflect.Value) error {
	//nolint:exhaustive
	switch dest.Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(text)
		if err != nil {
			return fmt.Errorf("parsing into field of type %s: %w", dest.Type().Name(), err)
		}
		dest.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(text, 10, dest.Type().Bits())
		if err != nil {
			return fmt.Errorf("parsing into field of type %s: %w", dest.Type().Name(), err)
		}
		dest.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(text, 10, dest.Type().Bits())
		if err != nil {
			return fmt.Errorf("parsing into field of type %s: %w", dest.Type().Name(), err)
		}
		dest.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(text, dest.Type().Bits())
		if err != nil {
			return fmt.Errorf("parsing into field of type %s: %w", dest.Type().Name(), err)
		}
		dest.SetFloat(v)
	case reflect.String:
		dest.SetString(text)
	default:
		return fmt.Errorf("unsupported field type %s", dest.Type().Name())
	}
	return nil
}
