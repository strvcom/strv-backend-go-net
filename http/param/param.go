package param

import (
	"encoding"
	"fmt"
	"net/http"
	"reflect"
)

func DefaultParamParser() ParamParser {
	return ParamParser{
		QueryParamTagResolver: FixedTagNameParamTagResolver("queryparam"),
		PathParamTagResolver:  FixedTagNameParamTagResolver("pathparam"),
		PathParamFunc:         nil, // keep nil, as there is no sensible default of how to get value of path parameter
	}
}

// ParamTagResolver is a function that decides from a field type what key of http parameter should be searched.
// Second return value should return whether the key should be searched in http parameter at all.
type ParamTagResolver func(fieldTag reflect.StructTag) (string, bool)

// PathParamFunc is a function that returns value of specified http path parameter
type PathParamFunc func(r *http.Request, key string) string

type ParamParser struct {
	QueryParamTagResolver ParamTagResolver
	PathParamTagResolver  ParamTagResolver
	PathParamFunc         PathParamFunc
}

func (p ParamParser) WithPathParamFunc(f PathParamFunc) ParamParser {
	copied := p
	copied.PathParamFunc = f
	return copied
}

func (p ParamParser) Parse(r *http.Request, dest any) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Pointer {
		return fmt.Errorf("cannot set non-pointer value of type %s", v.Type().Name())
	}
	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("can only parse into struct, but got %s", v.Type().Name())
	}

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag
		if paramName, ok := p.QueryParamTagResolver(tag); ok {
			query := r.URL.Query()
			if query.Has(paramName) {
				paramValue := query.Get(paramName)
				err := unmarshalValue(paramValue, v.Field(i).Addr().Interface())
				if err != nil {
					return fmt.Errorf("unmarshaling query parameter %s: %w", paramName, err)
				}
			}
		}
		if paramName, ok := p.PathParamTagResolver(tag); ok {
			if p.PathParamFunc == nil {
				return fmt.Errorf("struct's field was tagged for parsing the path parameter (%s) but PathParamFunc to get value of path parameter is not defined", paramName)
			}
			paramValue := p.PathParamFunc(r, paramName)
			if paramValue != "" {
				err := unmarshalValue(paramValue, v.Field(i).Addr().Interface())
				if err != nil {
					return fmt.Errorf("unmarshaling path parameter %s: %w", paramName, err)
				}
			}
		}
	}
	return nil
}

func unmarshalValue(text string, dest any) error {
	if unmarshaler, ok := dest.(encoding.TextUnmarshaler); ok {
		return unmarshaler.UnmarshalText([]byte(text))
	}
	t := reflect.TypeOf(dest).Elem()
	if t.Kind() == reflect.Pointer {
		return unmarshalValue(text, reflect.New(t))
	}
	_, err := fmt.Sscan(text, dest)
	return err
}

func FixedTagNameParamTagResolver(tagName string) ParamTagResolver {
	return func(fieldTag reflect.StructTag) (string, bool) {
		taggedParamName := fieldTag.Get(tagName)
		return taggedParamName, taggedParamName != ""
	}
}
