package stanza

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"
)

// ---------------------------------------------------------------------------
// Package-level convenience
// ---------------------------------------------------------------------------

// MapTo loads source(s) into an INI File, then maps the default section
// into the struct pointed to by v.
func MapTo(v interface{}, source interface{}, others ...interface{}) error {
	return MapToWithMapper(v, nil, source, others...)
}

// MapToWithMapper is like MapTo but uses the given NameMapper.
func MapToWithMapper(v interface{}, mapper NameMapper, source interface{}, others ...interface{}) error {
	cfg, err := Load(source, others...)
	if err != nil {
		return err
	}
	cfg.NameMapper = mapper
	return cfg.MapTo(v)
}

// StrictMapTo is like MapTo but returns an error for unmapped keys.
func StrictMapTo(v interface{}, source interface{}, others ...interface{}) error {
	return StrictMapToWithMapper(v, nil, source, others...)
}

// StrictMapToWithMapper is like StrictMapTo with a custom NameMapper.
func StrictMapToWithMapper(v interface{}, mapper NameMapper, source interface{}, others ...interface{}) error {
	cfg, err := Load(source, others...)
	if err != nil {
		return err
	}
	cfg.NameMapper = mapper
	return cfg.StrictMapTo(v)
}

// ---------------------------------------------------------------------------
// File methods
// ---------------------------------------------------------------------------

// MapTo maps the File into the struct pointed to by v.
func (f *File) MapTo(v interface{}) error {
	return f.mapTo(v, false)
}

// StrictMapTo maps the File into v, returning an error if any
// INI keys are not mapped to struct fields.
func (f *File) StrictMapTo(v interface{}) error {
	return f.mapTo(v, true)
}

func (f *File) mapTo(v interface{}, strict bool) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("MapTo requires a non-nil pointer to a struct")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errors.New("MapTo requires a pointer to a struct")
	}
	return f.mapSection(rv, f.Section(DefaultSection), strict)
}

// ReflectFrom populates the File from the struct v.
func (f *File) ReflectFrom(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return errors.New("ReflectFrom requires a struct or pointer to struct")
	}
	return f.reflectSection(rv, f.Section(DefaultSection))
}

// ReflectFromWithMapper is like ReflectFrom with a custom NameMapper.
func (f *File) ReflectFromWithMapper(v interface{}, mapper NameMapper) error {
	f.NameMapper = mapper
	return f.ReflectFrom(v)
}

// ---------------------------------------------------------------------------
// Section → struct
// ---------------------------------------------------------------------------

func (f *File) mapSection(rv reflect.Value, sec *Section, strict bool) error {
	rt := rv.Type()
	mapped := make(map[string]bool)

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)

		if !fv.CanSet() {
			continue
		}

		tag := field.Tag.Get("ini")
		if tag == "-" {
			continue
		}

		keyName, opts := parseTag(tag)
		if keyName == "" {
			keyName = f.fieldName(field.Name)
		}
		_ = opts // omitempty only used on write

		// Nested struct → map to a section with that name.
		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			subSec := f.Section(keyName)
			if err := f.mapSection(fv, subSec, strict); err != nil {
				return err
			}
			continue
		}

		// Pointer to struct.
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct && field.Type.Elem() != reflect.TypeOf(time.Time{}) {
			if !f.HasSection(keyName) {
				continue
			}
			if fv.IsNil() {
				fv.Set(reflect.New(field.Type.Elem()))
			}
			subSec := f.Section(keyName)
			if err := f.mapSection(fv.Elem(), subSec, strict); err != nil {
				return err
			}
			continue
		}

		if !sec.HasKey(keyName) {
			continue
		}
		mapped[keyName] = true

		key := sec.Key(keyName)
		delim := field.Tag.Get("ini-delim")
		if delim == "" {
			delim = ","
		}

		if err := setField(fv, key, delim); err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
	}

	if strict {
		for _, k := range sec.Keys() {
			if !mapped[k.Name()] {
				return fmt.Errorf("key %q in section %q has no corresponding struct field", k.Name(), sec.Name())
			}
		}
	}

	return nil
}

func setField(fv reflect.Value, key *Key, delim string) error {
	val := key.Value()

	// Handle slices.
	if fv.Kind() == reflect.Slice {
		return setSlice(fv, key, delim)
	}

	switch fv.Kind() {
	case reflect.String:
		fv.SetString(val)
	case reflect.Bool:
		v, err := parseBool(val)
		if err != nil {
			return err
		}
		fv.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// time.Duration is int64
		if fv.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(val)
			if err != nil {
				return err
			}
			fv.SetInt(int64(d))
		} else {
			v, err := key.Int64()
			if err != nil {
				return err
			}
			fv.SetInt(v)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := key.Uint64()
		if err != nil {
			return err
		}
		fv.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := key.Float64()
		if err != nil {
			return err
		}
		fv.SetFloat(v)
	case reflect.Struct:
		if fv.Type() == reflect.TypeOf(time.Time{}) {
			v, err := key.Time()
			if err != nil {
				return err
			}
			fv.Set(reflect.ValueOf(v))
		} else {
			return fmt.Errorf("unsupported struct type %s", fv.Type())
		}
	default:
		return fmt.Errorf("unsupported type %s", fv.Type())
	}
	return nil
}

func setSlice(fv reflect.Value, key *Key, delim string) error {
	parts := key.Strings(delim)
	elemType := fv.Type().Elem()

	slice := reflect.MakeSlice(fv.Type(), 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		elem := reflect.New(elemType).Elem()
		tmpKey := newKey(nil, "", p)
		if err := setField(elem, tmpKey, delim); err != nil {
			return err
		}
		slice = reflect.Append(slice, elem)
	}
	fv.Set(slice)
	return nil
}

// ---------------------------------------------------------------------------
// struct → Section
// ---------------------------------------------------------------------------

func (f *File) reflectSection(rv reflect.Value, sec *Section) error {
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)

		if !unicode.IsUpper(rune(field.Name[0])) {
			continue
		}

		tag := field.Tag.Get("ini")
		if tag == "-" {
			continue
		}

		keyName, opts := parseTag(tag)
		if keyName == "" {
			keyName = f.fieldName(field.Name)
		}

		comment := field.Tag.Get("ini-comment")
		delim := field.Tag.Get("ini-delim")
		if delim == "" {
			delim = ","
		}

		// Nested struct → subsection.
		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			subSec, err := f.NewSection(keyName)
			if err != nil {
				return err
			}
			if comment != "" {
				subSec.Comment = comment
			}
			if err := f.reflectSection(fv, subSec); err != nil {
				return err
			}
			continue
		}

		// Pointer to struct.
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct && field.Type.Elem() != reflect.TypeOf(time.Time{}) {
			if fv.IsNil() {
				continue
			}
			subSec, err := f.NewSection(keyName)
			if err != nil {
				return err
			}
			if comment != "" {
				subSec.Comment = comment
			}
			if err := f.reflectSection(fv.Elem(), subSec); err != nil {
				return err
			}
			continue
		}

		// Check omitempty.
		if opts.contains("omitempty") && isZero(fv) {
			continue
		}

		val := fieldToString(fv, delim)
		k, err := sec.NewKey(keyName, val)
		if err != nil {
			return err
		}
		if comment != "" {
			k.Comment = comment
		}
	}
	return nil
}

func fieldToString(fv reflect.Value, delim string) string {
	// Slices.
	if fv.Kind() == reflect.Slice {
		parts := make([]string, fv.Len())
		for i := 0; i < fv.Len(); i++ {
			parts[i] = fieldToString(fv.Index(i), delim)
		}
		return strings.Join(parts, delim)
	}

	switch fv.Kind() {
	case reflect.String:
		return fv.String()
	case reflect.Bool:
		if fv.Bool() {
			return "true"
		}
		return "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fv.Type() == reflect.TypeOf(time.Duration(0)) {
			return time.Duration(fv.Int()).String()
		}
		return fmt.Sprintf("%d", fv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", fv.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", fv.Float())
	case reflect.Struct:
		if fv.Type() == reflect.TypeOf(time.Time{}) {
			t := fv.Interface().(time.Time)
			if t.IsZero() {
				return ""
			}
			return t.Format(time.RFC3339)
		}
	}
	return fmt.Sprintf("%v", fv.Interface())
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (f *File) fieldName(name string) string {
	if f.NameMapper != nil {
		return f.NameMapper(name)
	}
	return name
}

type tagOpts string

func parseTag(tag string) (string, tagOpts) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOpts(tag[idx+1:])
	}
	return tag, ""
}

func (o tagOpts) contains(opt string) bool {
	for o != "" {
		var name string
		if i := strings.Index(string(o), ","); i >= 0 {
			name, o = string(o[:i]), o[i+1:]
		} else {
			name, o = string(o), ""
		}
		if name == opt {
			return true
		}
	}
	return false
}

func isZero(fv reflect.Value) bool {
	switch fv.Kind() {
	case reflect.String:
		return fv.String() == ""
	case reflect.Bool:
		return !fv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return fv.Float() == 0
	case reflect.Slice:
		return fv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return fv.IsNil()
	case reflect.Struct:
		if fv.Type() == reflect.TypeOf(time.Time{}) {
			return fv.Interface().(time.Time).IsZero()
		}
		return false
	}
	return false
}
