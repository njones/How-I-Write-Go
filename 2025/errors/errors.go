package e

import (
	"encoding/json"
	"fmt"
)

type ErrorString = ErrorStr[any]

type errorF interface {
	error
	F(...any) error
}

type errorNil struct{ error }

func IsNilableErr(err error) errorNil { return errorNil{err} }

type ErrorStr[T any] string

func (e ErrorStr[T]) Error() string        { return string(e) }
func (e ErrorStr[T]) Is(target error) bool { _, ok := target.(T); return ok }
func (e ErrorStr[T]) KV(v ...any) errorF   { return rtnErr[T](e, v, e) }
func (e ErrorStr[T]) F(v ...any) error {
	if hasNilErr(v) {
		return nil
	}
	return rtnErr[T](e, []any{}, fmt.Errorf(string(e), v...))
}

type errorFun[T any] func() (ErrorStr[T], []any, error)

func (efn errorFun[T]) Interface() ([]any, func() any) {
	a, b, c := efn()
	return b, func() any { return rtnErr(a, []any{}, c) }
}

func (efn errorFun[T]) Error() string {
	_, kv, err := efn()

	if len(kv) == 0 {
		return err.Error()
	}

	var m = make(map[string]any)
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}

	return err.Error() + " " + string(must(json.Marshal(m)))
}
func (efn errorFun[T]) Is(target error) bool { _, ok := target.(errorFun[T]); return ok }
func (efn errorFun[T]) Unwrap() error        { baseErr, _, _ := efn(); return baseErr }
func (efn errorFun[T]) F(v ...any) error {
	if hasNilErr(v) {
		return nil
	}

	errBase, keyValue, err := efn()
	kv, v := mergeKeyValues[T](keyValue, v)

	return rtnErr[T](errBase, kv, fmt.Errorf(err.Error(), v...))
}

func hasNilErr(v []any) bool {
	for _, value := range v {
		switch val := value.(type) {
		case errorNil:
			if val.error == nil {
				return true
			}
		}
	}
	return false
}

func mergeKeyValues[T any](keyValue []any, v []any) ([]any, []any) {
	for i, value := range v {
	Recheck:
		switch val := value.(type) {
		case errorFun[T]:
			b, kv, e := val()
			keyValue = append(keyValue, kv...)
			v[i] = rtnErr[T](b, []any{}, e)
		case interface{ Interface() ([]any, func() any) }:
			kv, fn := val.Interface()
			keyValue = append(keyValue, kv...)
			v[i] = fn()
		case errorNil:
			value = val.error
			goto Recheck
		}
	}
	return keyValue, v
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func rtnErr[T any](errBase ErrorStr[T], kv []any, err error) errorFun[T] {
	return errorFun[T](func() (ErrorStr[T], []any, error) { return errBase, kv, err })
}
