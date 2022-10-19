package binding

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
	"sync"
)

var Validator StructValidator = &defaultValidator{}

type SliceValidationError []error

func (err SliceValidationError) Error() string {
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			fmt.Fprintf(&b, "[%d]: %s", 0, err[0].Error())
		}
		if n > 1 {
			for i := 0; i < n; i++ {
				if err[i] != nil {
					b.WriteString("\n")
					fmt.Fprintf(&b, "[%d]: %s", i, err[i].Error())
				}
			}
		}
		return b.String()
	}
}

type StructValidator interface {
	ValidatorStruct(any) error

	Engine() any
}

type defaultValidator struct {
	one      sync.Once // single instance
	validate *validator.Validate
}

func (d *defaultValidator) ValidatorStruct(obj any) error {
	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Struct:
		return d.validateStruct(obj)
	case reflect.Pointer:
		pointerData := value.Elem().Interface()
		return d.ValidatorStruct(pointerData)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		errs := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := d.ValidatorStruct(value.Index(i).Interface()); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return errs
		}
		return nil
	}
	return nil
}

func (d *defaultValidator) validateStruct(obj any) error {
	d.lazyInit()
	fmt.Println(obj)
	return d.validate.Struct(obj)
}

func (d *defaultValidator) Engine() any {
	d.lazyInit()
	return d.validate
}

func (d *defaultValidator) lazyInit() {
	d.one.Do(func() {
		d.validate = validator.New()
	})
}

func validate(obj any) error {
	return Validator.ValidatorStruct(obj)
}
