package msgpack

import (
	"encoding"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/vmihailenco/tagparser/v2"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()

var (
	customEncoderType = reflect.TypeOf((*CustomEncoder)(nil)).Elem()
	customDecoderType = reflect.TypeOf((*CustomDecoder)(nil)).Elem()
)

var (
	marshalerType   = reflect.TypeOf((*Marshaler)(nil)).Elem()
	unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
)

var (
	binaryMarshalerType   = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
	binaryUnmarshalerType = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
)

var (
	textMarshalerType   = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

type (
	encoderFunc func(*Encoder, reflect.Value) error
	decoderFunc func(*Decoder, reflect.Value) error
)

var (
	typeEncMap sync.Map
	typeDecMap sync.Map
)

// Register registers encoder and decoder functions for a value.
// This is low level API and in most cases you should prefer implementing
// CustomEncoder/CustomDecoder or Marshaler/Unmarshaler interfaces.
func Register(value interface{}, enc encoderFunc, dec decoderFunc) {
	typ := reflect.TypeOf(value)
	if enc != nil {
		typeEncMap.Store(typ, enc)
	}
	if dec != nil {
		typeDecMap.Store(typ, dec)
	}
}

//------------------------------------------------------------------------------

const defaultStructTag = "msgpack"

var structs = newStructCache()

type structCache struct {
	m sync.Map
}

type structCacheKey struct {
	typ reflect.Type
	tag string
}

func newStructCache() *structCache {
	return new(structCache)
}

func (m *structCache) Fields(typ reflect.Type, tag string) *fields {
	key := structCacheKey{tag: tag, typ: typ}

	if v, ok := m.m.Load(key); ok {
		return v.(*fields)
	}

	fs := getFields(typ, tag)
	m.m.Store(key, fs)

	return fs
}

//------------------------------------------------------------------------------

type field struct {
	encoder   encoderFunc
	decoder   decoderFunc
	name      string
	index     []int
	omitEmpty bool
	omitZero  bool // honor the encoding/json 1.24 `omitzero` tag option.
}

func (f *field) Omit(e *Encoder, strct reflect.Value) bool {
	v, ok := fieldByIndex(strct, f.index)
	if !ok {
		return true
	}
	// `omitzero` omits the field when it holds the zero value for its
	// type, matching encoding/json (Go 1.24+). It is independent of
	// `omitempty`: a custom IsZero takes precedence, and unlike omitempty it
	// does not treat a non-nil empty slice/map as omittable.
	if f.omitZero && isZeroValue(v) {
		return true
	}
	forced := e.flags&omitEmptyFlag != 0
	return (f.omitEmpty || forced) && e.isEmptyValue(v)
}

func (f *field) EncodeValue(e *Encoder, strct reflect.Value) error {
	v, ok := fieldByIndex(strct, f.index)
	if !ok {
		return e.EncodeNil()
	}
	return f.encoder(e, v)
}

func (f *field) DecodeValue(d *Decoder, strct reflect.Value) error {
	v := fieldByIndexAlloc(strct, f.index)
	return f.decoder(d, v)
}

//------------------------------------------------------------------------------

type fields struct {
	Type    reflect.Type
	Map     map[string]*field
	List    []*field
	AsArray bool

	hasOmitEmpty bool
	hasOmitZero  bool // any field carries the `omitzero` tag option.
}

func newFields(typ reflect.Type) *fields {
	return &fields{
		Type: typ,
		Map:  make(map[string]*field, typ.NumField()),
		List: make([]*field, 0, typ.NumField()),
	}
}

func (fs *fields) Add(field *field) {
	fs.warnIfFieldExists(field.name)
	fs.Map[field.name] = field
	fs.List = append(fs.List, field)
	if field.omitEmpty {
		fs.hasOmitEmpty = true
	}
	if field.omitZero {
		fs.hasOmitZero = true
	}
}

func (fs *fields) warnIfFieldExists(name string) {
	if _, ok := fs.Map[name]; ok {
		log.Printf("msgpack: %s already has field=%s", fs.Type, name)
	}
}

func (fs *fields) OmitEmpty(e *Encoder, strct reflect.Value) []*field {
	forced := e.flags&omitEmptyFlag != 0
	if !fs.hasOmitEmpty && !fs.hasOmitZero && !forced {
		return fs.List
	}

	fields := make([]*field, 0, len(fs.List))

	for _, f := range fs.List {
		if !f.Omit(e, strct) {
			fields = append(fields, f)
		}
	}

	return fields
}

func getFields(typ reflect.Type, fallbackTag string) *fields {
	fs := newFields(typ)

	var omitEmpty bool
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)

		tagStr := f.Tag.Get(defaultStructTag)
		if tagStr == "" && fallbackTag != "" {
			tagStr = f.Tag.Get(fallbackTag)
		}

		tag := tagparser.Parse(tagStr)
		if tag.Name == "-" {
			continue
		}

		if f.Name == "_msgpack" {
			fs.AsArray = tag.HasOption("as_array") || tag.HasOption("asArray")
			if tag.HasOption("omitempty") {
				omitEmpty = true
			}
		}

		if f.PkgPath != "" && !f.Anonymous {
			continue
		}

		field := &field{
			name:      tag.Name,
			index:     f.Index,
			omitEmpty: omitEmpty || tag.HasOption("omitempty"),
			omitZero:  tag.HasOption("omitzero"),
		}

		if tag.HasOption("intern") {
			switch f.Type.Kind() {
			case reflect.Interface:
				field.encoder = encodeInternedInterfaceValue
				field.decoder = decodeInternedInterfaceValue
			case reflect.String:
				field.encoder = encodeInternedStringValue
				field.decoder = decodeInternedStringValue
			default:
				err := fmt.Errorf("msgpack: intern strings are not supported on %s", f.Type)
				panic(err)
			}
		} else {
			field.encoder = getEncoder(f.Type)
			field.decoder = getDecoder(f.Type)
		}

		if field.name == "" {
			field.name = f.Name
		}

		if f.Anonymous && !tag.HasOption("noinline") {
			inline := tag.HasOption("inline")
			if inline {
				inlineFields(fs, f.Type, field, fallbackTag)
			} else {
				inline = shouldInline(fs, f.Type, field, fallbackTag)
			}

			if inline {
				if _, ok := fs.Map[field.name]; ok {
					log.Printf("msgpack: %s already has field=%s", fs.Type, field.name)
				}
				fs.Map[field.name] = field
				continue
			}
		}

		fs.Add(field)

		if alias, ok := tag.Options["alias"]; ok {
			fs.warnIfFieldExists(alias)
			fs.Map[alias] = field
		}
	}
	return fs
}

var (
	encodeStructValuePtr uintptr
	decodeStructValuePtr uintptr
)

//nolint:gochecknoinits
func init() {
	encodeStructValuePtr = reflect.ValueOf(encodeStructValue).Pointer()
	decodeStructValuePtr = reflect.ValueOf(decodeStructValue).Pointer()
}

func inlineFields(fs *fields, typ reflect.Type, f *field, tag string) {
	inlinedFields := getFields(typ, tag).List
	for _, field := range inlinedFields {
		if _, ok := fs.Map[field.name]; ok {
			// Don't inline shadowed fields.
			continue
		}
		field.index = append(f.index, field.index...)
		fs.Add(field)
	}
}

func shouldInline(fs *fields, typ reflect.Type, f *field, tag string) bool {
	var encoder encoderFunc
	var decoder decoderFunc

	if typ.Kind() == reflect.Struct {
		encoder = f.encoder
		decoder = f.decoder
	} else {
		for typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
			encoder = getEncoder(typ)
			decoder = getDecoder(typ)
		}
		if typ.Kind() != reflect.Struct {
			return false
		}
	}

	if reflect.ValueOf(encoder).Pointer() != encodeStructValuePtr {
		return false
	}
	if reflect.ValueOf(decoder).Pointer() != decodeStructValuePtr {
		return false
	}

	inlinedFields := getFields(typ, tag).List
	for _, field := range inlinedFields {
		if _, ok := fs.Map[field.name]; ok {
			// Don't auto inline if there are shadowed fields.
			return false
		}
	}

	for _, field := range inlinedFields {
		field.index = append(f.index, field.index...)
		fs.Add(field)
	}
	return true
}

type isZeroer interface {
	IsZero() bool
}

// isZeroValue reports whether v is the zero value for its type, using the same
// rule as encoding/json's `omitzero` (Go 1.24+): a type's own IsZero method
// takes precedence; otherwise the value is compared against the zero value via
// reflection. This is intentionally distinct from isEmptyValue (omitempty),
// which also omits non-nil empty slices, maps, and strings.
func isZeroValue(v reflect.Value) bool {
	// A nil pointer or interface is the zero value; guard before invoking
	// IsZero so a typed-nil (e.g. (*time.Time)(nil)) whose type carries an
	// IsZero method does not get dereferenced. encoding/json guards this too.
	if k := v.Kind(); (k == reflect.Pointer || k == reflect.Interface) && v.IsNil() {
		return true
	}
	if z, ok := v.Interface().(isZeroer); ok {
		return z.IsZero()
	}
	if v.CanAddr() {
		if z, ok := v.Addr().Interface().(isZeroer); ok {
			return z.IsZero()
		}
	}
	return v.IsZero()
}

func (e *Encoder) isEmptyValue(v reflect.Value) bool {
	kind := v.Kind()

	for kind == reflect.Interface {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
		kind = v.Kind()
	}

	if z, ok := v.Interface().(isZeroer); ok {
		return nilable(kind) && v.IsNil() || z.IsZero()
	}

	switch kind {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Struct:
		structFields := structs.Fields(v.Type(), e.structTag)
		fields := structFields.OmitEmpty(e, v)
		return len(fields) == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}

func fieldByIndex(v reflect.Value, index []int) (_ reflect.Value, ok bool) {
	if len(index) == 1 {
		return v.Field(index[0]), true
	}

	for i, idx := range index {
		if i > 0 {
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					return v, false
				}
				v = v.Elem()
			}
		}
		v = v.Field(idx)
	}

	return v, true
}

func fieldByIndexAlloc(v reflect.Value, index []int) reflect.Value {
	if len(index) == 1 {
		return v.Field(index[0])
	}

	for i, idx := range index {
		if i > 0 {
			var ok bool
			v, ok = indirectNil(v)
			if !ok {
				return v
			}
		}
		v = v.Field(idx)
	}

	return v
}

func indirectNil(v reflect.Value) (reflect.Value, bool) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			if !v.CanSet() {
				return v, false
			}
			elemType := v.Type().Elem()
			if elemType.Kind() != reflect.Struct {
				return v, false
			}
			v.Set(cachedValue(elemType))
		}
		v = v.Elem()
	}
	return v, true
}
