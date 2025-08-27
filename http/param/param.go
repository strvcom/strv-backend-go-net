package param

import (
	"encoding"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

const (
	defaultTagName      = "param"
	defaultMaxMemory    = 32 << 20 // 32 MB
	queryTagValuePrefix = "query"
	pathTagValuePrefix  = "path"
	formTagValuePrefix  = "form"
)

// TagResolver is a function that decides from a field tag what parameter should be searched.
// Second return value should return whether the parameter should be searched at all.
type TagResolver func(fieldTag reflect.StructTag) (string, bool)

// TagNameResolver returns a TagResolver that returns the value of tag with tagName, and whether the tag exists at all.
// It can be used to replace Parser.ParamTagResolver to change what tag name the Parser reacts to.
func TagNameResolver(tagName string) TagResolver {
	return func(fieldTag reflect.StructTag) (string, bool) {
		tagValue := fieldTag.Get(tagName)
		if tagValue == "" {
			return "", false
		}
		return tagValue, true
	}
}

// PathParamFunc is a function that returns value of specified http path parameter.
type PathParamFunc func(r *http.Request, key string) string

// FormParamFunc is a function that returns value of specified form parameter.
type FormParamFunc func(r *http.Request, key string) string

func DefaultFormParamFunc(r *http.Request, key string) string {
	return r.PostFormValue(key)
}

// Parser can Parse query and path parameters from http.Request into a struct.
// Fields struct have to be tagged such that either QueryParamTagResolver or PathParamTagResolver returns
// valid parameter name from the provided tag.
//
// PathParamFunc is for getting path parameter from http.Request, as each http router handles it in different way (if at all).
// For example for chi, use WithPathParamFunc(chi.URLParam) to be able to use tags for path parameters.
type Parser struct {
	ParamTagResolver TagResolver
	PathParamFunc    PathParamFunc
	FormParamFunc    FormParamFunc
}

// DefaultParser returns query and path parameter Parser with intended struct tags
// `param:"query=param_name"` for query parameters and `param:"path=param_name"` for path parameters
func DefaultParser() Parser {
	return Parser{
		ParamTagResolver: TagNameResolver(defaultTagName),
		PathParamFunc:    nil, // keep nil, as there is no sensible default of how to get value of path parameter
		FormParamFunc:    DefaultFormParamFunc,
	}
}

// WithPathParamFunc returns a copy of Parser with set function for getting path parameters from http.Request.
// For more see Parser description.
func (p Parser) WithPathParamFunc(f PathParamFunc) Parser {
	p.PathParamFunc = f
	return p
}

// WithFormParamFunc returns a copy of Parser with set function for getting form parameters from http.Request.
// For more see Parser description.
func (p Parser) WithFormParamFunc(f FormParamFunc) Parser {
	p.FormParamFunc = f
	return p
}

// Parse accepts the request and a pointer to struct with its fields tagged with appropriate tags set in Parser.
// Such tagged fields must be in top level struct, or in exported struct embedded in top-level struct.
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

	fieldIndexPaths := p.findTaggedIndexPaths(v.Type(), []int{}, []taggedFieldIndexPath{})

	for i := range fieldIndexPaths {
		// Zero the value, even if it would not be set by following path or query parameter.
		// This will cause potential partial result from previous parser (e.g. json.Unmarshal) to be discarded on
		// fields that are tagged for path or query parameter.
		err := zeroPath(v, &fieldIndexPaths[i])
		if err != nil {
			return err
		}
	}

	for _, path := range fieldIndexPaths {
		err := p.parseParam(r, path)
		if err != nil {
			return err
		}
	}
	return nil
}

type paramType int

const (
	paramTypeQuery paramType = iota
	paramTypePath
	paramTypeForm
)

type taggedFieldIndexPath struct {
	paramType paramType
	paramName string
	indexPath []int
	destValue reflect.Value
}

func (p Parser) findTaggedIndexPaths(typ reflect.Type, currentNestingIndexPath []int, paths []taggedFieldIndexPath) []taggedFieldIndexPath {
	for i := range typ.NumField() {
		typeField := typ.Field(i)
		if typeField.Anonymous {
			t := typeField.Type
			if t.Kind() == reflect.Pointer {
				t = t.Elem()
			}
			if t.Kind() == reflect.Struct {
				paths = p.findTaggedIndexPaths(t, append(currentNestingIndexPath, i), paths)
			}
		}
		if !typeField.IsExported() {
			continue
		}
		tag := typeField.Tag
		pathParamName, okPath := p.resolvePath(tag)
		formParamName, okForm := p.resolveForm(tag)
		queryParamName, okQuery := p.resolveQuery(tag)

		newPath := append(append([]int{}, currentNestingIndexPath...), i)
		if okPath {
			paths = append(paths, taggedFieldIndexPath{
				paramType: paramTypePath,
				paramName: pathParamName,
				indexPath: newPath,
			})
		}
		if okForm {
			paths = append(paths, taggedFieldIndexPath{
				paramType: paramTypeForm,
				paramName: formParamName,
				indexPath: newPath,
			})
		}
		if okQuery {
			paths = append(paths, taggedFieldIndexPath{
				paramType: paramTypeQuery,
				paramName: queryParamName,
				indexPath: newPath,
			})
		}
	}
	return paths
}

func zeroPath(v reflect.Value, path *taggedFieldIndexPath) error {
	for n, i := range path.indexPath {
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		// findTaggedIndexPaths prepared a path.indexPath in such a way, that respective field is always
		// pointer to struct or struct -> should be always able to .Field() here
		typeField := v.Type().Field(i)
		v = v.Field(i)

		if n == len(path.indexPath)-1 {
			v.Set(reflect.Zero(typeField.Type))
			path.destValue = v
		} else if v.Kind() == reflect.Pointer && v.IsNil() {
			if !v.CanSet() {
				return fmt.Errorf("cannot set embedded pointer to unexported struct: %v", v.Type().Elem())
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
	}
	return nil
}

func (p Parser) parseParam(r *http.Request, path taggedFieldIndexPath) error {
	switch path.paramType {
	case paramTypePath:
		err := p.parsePathParam(r, path.paramName, path.destValue)
		if err != nil {
			return err
		}
	case paramTypeForm:
		err := p.parseFormParam(r, path.paramName, path.destValue)
		if err != nil {
			return err
		}
	case paramTypeQuery:
		err := p.parseQueryParam(r, path.paramName, path.destValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p Parser) parsePathParam(r *http.Request, paramName string, v reflect.Value) error {
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
	return nil
}

func (p Parser) parseFormParam(r *http.Request, paramName string, v reflect.Value) error {
	if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
		return fmt.Errorf("struct's field was tagged for parsing the form parameter (%s) but request method is not POST, PUT or PATCH", paramName)
	}
	if err := r.ParseMultipartForm(defaultMaxMemory); err != nil {
		if !errors.Is(err, http.ErrNotMultipart) {
			return fmt.Errorf("parsing multipart form: %w", err)
		}
		// Try to parse regular form if not multipart form.
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("parsing form: %w", err)
		}
	}
	paramValue := p.FormParamFunc(r, paramName)
	if paramValue != "" {
		err := unmarshalValue(paramValue, v)
		if err != nil {
			return fmt.Errorf("unmarshaling form parameter %s: %w", paramName, err)
		}
	}
	return nil
}

func (p Parser) parseQueryParam(r *http.Request, paramName string, v reflect.Value) error {
	query := r.URL.Query()
	if values, ok := query[paramName]; ok && len(values) > 0 {
		err := unmarshalValueOrSlice(values, v)
		if err != nil {
			return fmt.Errorf("unmarshaling query parameter %s: %w", paramName, err)
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

// resolveTagValueWithModifier returns a parameter value in tag value containing a prefix "tagModifier=".
// Example: resolveTagValueWithModifier("query=param_name", "query") returns "param_name", true.
func (p Parser) resolveTagValueWithModifier(tagValue string, tagModifier string) (string, bool) {
	splits := strings.Split(tagValue, "=")
	//nolint:mnd // 2 not really that magic number - one value before '=', one after
	if len(splits) != 2 {
		return "", false
	}
	if splits[0] == tagModifier {
		return splits[1], true
	}
	return "", false
}

func (p Parser) resolveTagWithModifier(fieldTag reflect.StructTag, tagModifier string) (string, bool) {
	tagValue, ok := p.ParamTagResolver(fieldTag)
	if !ok {
		return "", false
	}
	return p.resolveTagValueWithModifier(tagValue, tagModifier)
}

func (p Parser) resolvePath(fieldTag reflect.StructTag) (string, bool) {
	return p.resolveTagWithModifier(fieldTag, pathTagValuePrefix)
}

func (p Parser) resolveForm(fieldTag reflect.StructTag) (string, bool) {
	return p.resolveTagWithModifier(fieldTag, formTagValuePrefix)
}

func (p Parser) resolveQuery(fieldTag reflect.StructTag) (string, bool) {
	return p.resolveTagWithModifier(fieldTag, queryTagValuePrefix)
}
