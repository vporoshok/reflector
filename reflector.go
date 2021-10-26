package reflector

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Reflector of structure.
type Reflector struct {
	stype reflect.Type
	value reflect.Value
}

// New is an constructor from structure instance.
func New(i interface{}) Reflector {
	r := Reflector{
		stype: reflect.TypeOf(i),
		value: reflect.ValueOf(i),
	}
	if r.stype.Kind() == reflect.Ptr {
		r.stype = r.stype.Elem()
		r.value = r.value.Elem()
	}

	return r
}

// FromValue is an constructor from existed reflect.Value.
func FromValue(v reflect.Value) Reflector {
	r := Reflector{
		stype: v.Type(),
		value: v,
	}

	if r.stype.Kind() == reflect.Ptr {
		r.stype = r.stype.Elem()
		r.value = r.value.Elem()
	}

	return r
}

// Value of reflected variable.
func (r Reflector) Value() interface{} {
	return r.value.Addr().Interface()
}

type extractConfig struct {
	tagName      string
	skipEmbedded bool
	skipEmpty    bool
	skipMinus    bool
}

func (cfg extractConfig) getTag(f reflect.StructField) (string, bool) {
	tag, ok := f.Tag.Lookup(cfg.tagName)
	if !ok && cfg.skipEmpty {
		return "", false
	}

	if tag == "-" && cfg.skipMinus {
		return "", false
	}

	return tag, true
}

// ExtractOption an option for Reflector.ExtractValues and Reflector.ExtractTags.
type ExtractOption interface {
	Apply(extractConfig) extractConfig
}

type extractOptionFunc func(extractConfig) extractConfig

func (f extractOptionFunc) Apply(cfg extractConfig) extractConfig {
	return f(cfg)
}

// WithoutEmbedded skip embedded structures.
func WithoutEmbedded() ExtractOption {
	return extractOptionFunc(func(cfg extractConfig) extractConfig {
		cfg.skipEmbedded = true

		return cfg
	})
}

// WithoutEmpty skip fields without tag.
func WithoutEmpty() ExtractOption {
	return extractOptionFunc(func(cfg extractConfig) extractConfig {
		cfg.skipEmpty = true

		return cfg
	})
}

// WithoutMinus skip fields with tag setted to minus sign.
func WithoutMinus() ExtractOption {
	return extractOptionFunc(func(cfg extractConfig) extractConfig {
		cfg.skipMinus = true

		return cfg
	})
}

// ExtractValues returns hash-map which tag value refer to field value.
func (r Reflector) ExtractValues(tagName string, skipNils bool, opts ...ExtractOption) map[string]interface{} {
	tags := r.ExtractTags(tagName, opts...)
	res := make(map[string]interface{}, len(tags))

	for fieldName, tag := range tags {
		val := r.value.FieldByName(fieldName)
		if skipNils {
			if val.Kind() == reflect.Ptr && val.IsNil() || val.Kind() == reflect.Slice && val.Len() == 0 {
				continue
			}
		}

		res[tag] = val.Interface()
	}

	return res
}

// ExtractTags returns hash-map which field value refer to tag value
func (r Reflector) ExtractTags(tagName string, opts ...ExtractOption) map[string]string {
	cfg := extractConfig{
		tagName: tagName,
	}
	for _, opt := range opts {
		cfg = opt.Apply(cfg)
	}

	m := map[string]string{}
	r.extractTags(cfg, r.stype, "", m)

	return m
}

func (r Reflector) extractTags(cfg extractConfig, t reflect.Type, prefix string, m map[string]string) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			if !cfg.skipEmbedded {
				r.extractTags(cfg, f.Type, prefix, m)
			}

			continue
		}

		switch {
		case f.Type.Kind() == reflect.Struct:
			r.extractTags(cfg, f.Type, prefix+f.Name+".", m)
		case f.Type.Kind() == reflect.Ptr && f.Type.Elem().Kind() == reflect.Struct:
			r.extractTags(cfg, f.Type.Elem(), prefix+f.Name+".", m)
		default:
			if tag, ok := cfg.getTag(f); ok {
				m[prefix+f.Name] = tag
			}
		}
	}
}

// Apply hash-table to reflected variable.
func (r Reflector) Apply(m map[string]string) error {
	s := r.value
	for k, v := range m {
		f := s
		for _, sk := range strings.Split(k, ".") {
			f = f.FieldByName(sk)
		}

		if err := r.processValue(f, v); err != nil {
			ft, _ := r.stype.FieldByName(k)

			return errors.WithMessage(err, ft.Name)
		}
	}

	return nil
}

func (r Reflector) processValue(value reflect.Value, source string) error {
	t := value.Type()
	if source == "" {
		value.Set(reflect.Zero(t))

		return nil
	}

	dst := reflect.New(t).Interface()
	if unmarshaler, ok := dst.(encoding.TextUnmarshaler); ok {
		if err := unmarshaler.UnmarshalText([]byte(source)); err != nil {
			return errors.WithStack(err)
		}

		value.Set(reflect.ValueOf(dst).Elem())

		return nil
	}

	if t.PkgPath() == "time" && t.Name() == "Duration" {
		d, err := time.ParseDuration(source)
		if err != nil {
			return errors.WithStack(err)
		}

		value.SetInt(int64(d))

		return nil
	}

	if t.Kind() == reflect.Slice {
		return r.processSlice(t, value, source)
	}

	if t.Kind() == reflect.String {
		value.SetString(source)

		return nil
	}

	if _, err := fmt.Sscan(source, dst); err != nil {
		return errors.WithStack(err)
	}

	value.Set(reflect.ValueOf(dst).Elem())

	return nil
}

func (r Reflector) processSlice(t reflect.Type, value reflect.Value, source string) error {
	sources := strings.Split(source, ",")
	values := reflect.MakeSlice(t, len(sources), len(sources))

	for i, val := range sources {
		err := r.processValue(values.Index(i), val)
		if err != nil {
			return err
		}
	}

	value.Set(values)

	return nil
}
