package htadaptor

import (
	"context"
	"errors"
	"reflect"
)

type FuncType uint8

const (
	FuncTypeUnary FuncType = iota + 1
	FuncTypeNullary
	FuncTypeVoid
)

var (
	// read https://github.com/golang/go/issues/35427
	// to understand how this magic works =>
	// the pointer to interface is important
	contextType     = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType       = reflect.TypeOf((*error)(nil)).Elem()
	validatableType = reflect.TypeOf((*interface {
		Validate(context.Context) error
	})(nil)).Elem()
)

func Detect(f any) (FuncType, error) {
	// https://medium.com/kokster/go-reflection-creating-objects-from-types-part-ii-composite-types-69a0e8134f20
	// https://github.com/golang/go/issues/50741
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return 0, errors.New("not a function")
	}
	parameterCount := t.NumIn()
	if parameterCount < 1 {
		return 0, errors.New("function must have at least one parameter")
	}
	if parameterCount > 2 {
		return 0, errors.New("functions with more than two parameters are not supported")
	}
	if !t.In(0).Implements(contextType) {
		return 0, errors.New("first function parameter must implement context.Context interface")
	}
	resultCount := t.NumOut()
	if resultCount < 1 {
		return 0, errors.New("function must return at least one value")
	}
	if resultCount > 2 {
		return 0, errors.New("functions with more than two return values are not supported")
	}
	if !t.Out(resultCount - 1).Implements(errorType) {
		return 0, errors.New("last return value must implement error interface")
	}

	if parameterCount == 2 {
		last := t.In(parameterCount - 1)
		if last.Kind() != reflect.Pointer {
			return 0, errors.New("last parameter must be a pointer")
		}
		if !last.Implements(validatableType) {
			return 0, errors.New("last parameter must implement Validatable interface")
		}
		if resultCount == 2 {
			return FuncTypeUnary, nil
		}
		return FuncTypeVoid, nil
	}
	return FuncTypeNullary, nil
}
