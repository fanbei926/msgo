package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type jsonBinding struct {
	DisallowUnknownFields bool
	IsValidate            bool
}

func (b jsonBinding) Name() string {
	return "json"
}

func (b jsonBinding) Bind(r *http.Request, object any) error {
	body := r.Body

	if body == nil {
		return errors.New("body is blank")
	}
	decoder := json.NewDecoder(body)
	if b.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if b.IsValidate {
		err := validateRequireParam(object, decoder)
		if err != nil {
			return err
		}
	} else {
		err := decoder.Decode(object)
		if err != nil {
			return err
		}
	}

	return validate(object)
}

func validateRequireParam(obj any, decoder *json.Decoder) error {
	// 1. get obj's value
	pointer := reflect.ValueOf(obj)
	// 2. actually obj's value is a pointer, so you must get it's real data
	if pointer.Kind() != reflect.Pointer {
		return errors.New("type is not a pointer")
	}
	data := pointer.Elem().Interface()
	// 3. so we mush get the real data of the data ^_^
	jsonData := reflect.ValueOf(data)
	switch jsonData.Kind() {
	case reflect.Struct:
		return checkParams(obj, jsonData, decoder)
	case reflect.Slice, reflect.Array:
		elem := jsonData.Type().Elem()
		if elem.Kind() == reflect.Struct {
			// now the data is a slice
			return checkParamsSlice(obj, elem, decoder)
		}

	default:
		err := decoder.Decode(obj)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func checkParamsSlice(obj any, sliceData reflect.Type, decoder *json.Decoder) error {
	sliceDatas := make([]map[string]interface{}, 0)
	err := decoder.Decode(&sliceDatas)
	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < sliceData.NumField(); i++ {
		field := sliceData.Field(i)
		name := field.Name
		jsonTag := field.Tag.Get("json")
		requiredTag := field.Tag.Get("msgo")
		if jsonTag != "" {
			name = jsonTag
		}
		name = strings.ToLower(name)
		for _, v := range sliceDatas {
			value := v[name]
			if value == nil && requiredTag == "required" {
				return errors.New(fmt.Sprintf("field %s is not exist, it must be required", name))
			}
			fmt.Println("  -->", name)
		}

	}
	b, err := json.Marshal(sliceDatas)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(b, obj)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func checkParams(obj any, jsonData reflect.Value, decoder *json.Decoder) error {
	jsonMap := make(map[string]interface{})
	err := decoder.Decode(&jsonMap)
	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < jsonData.NumField(); i++ {
		field := jsonData.Type().Field(i)
		name := field.Name
		jsonTag := field.Tag.Get("json")
		requiredTag := field.Tag.Get("msgo")
		if jsonTag != "" {
			name = jsonTag
		}
		name = strings.ToLower(name)
		value := jsonMap[name]
		if value == nil && requiredTag == "required" {
			return errors.New(fmt.Sprintf("field %s is not exist, it must be required", name))
		}
		fmt.Println("  -->", name)
	}
	b, err := json.Marshal(jsonMap)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(b, obj)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}
