package toJson

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/serenize/snaker"
	"net/http"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func writeError(err error) {
	if Logger != nil {
		Logger.Error(err)
	}
}

type jsonWriter interface {
	ToJson() interface{}
}

type JsonWriter jsonWriter

var Debug bool = false

var TypeError = errors.New("Could not get type")
var UnsupportedDatatype = errors.New("Unsupported Datatype")
var ReflectTypeError = errors.New("Got passed reflect.Type")
var ReflectValueError = errors.New("Got passed reflect.Value")

func fieldName(field reflect.StructField) (name string) {
	value := field.Tag.Get("json")
	fields := strings.Split(value, ",")

	// The format of the json tag is "<field>,<options>", with fields possibly being
	// empty
	if len(fields) > 0 && fields[0] != "" {
		return fields[0]
	}

	return snaker.CamelToSnake(field.Name)
}

func ToJson(i interface{}) (interface{}, error) {
	var err error

	_, ok := i.(reflect.Value)

	if ok {
		return nil, ReflectValueError
	}
	_, ok = i.(reflect.Type)

	if ok {
		return nil, ReflectValueError
	}

	t := reflect.TypeOf(i)

	if t == nil {
		return nil, nil
	}

	value := reflect.ValueOf(i)

	writer, ok := i.(jsonWriter)

	if ok {
		if Debug {
			fmt.Printf("Processing %T with ToJson()\n", i)
		}
		return ToJson(writer.ToJson())
	}

	marshaler, ok := i.(json.Marshaler)

	if ok {
		if Debug {
			fmt.Printf("Processing %T with MarshalJson()\n", i)
		}
		return marshaler, nil
	}

	errorer, ok := i.(error)
	if ok {
		if Debug {
			fmt.Printf("Processing %T with Error()\n", i)
		}

		return errorer.Error(), nil
	}

	switch t.Kind() {
	case reflect.Struct:
		if Debug {
			fmt.Printf("Processing %T as Struct\n", i)
		}
		x := make(map[string]interface{})

		for idx := 0; idx < t.NumField(); idx++ {
			def := t.Field(idx)
			val := value.Field(idx)

			// do not try and render unexported fields
			if !val.CanInterface() {
				continue
			}

			res, err := ToJson(val.Interface())

			if err != nil {
				return nil, err
			}

			name := fieldName(def)

			x[name] = res
		}

		return x, nil
	case reflect.Map:
		if Debug {
			fmt.Printf("Processing %T as Map\n", i)
		}
		input, ok := i.(map[string]interface{})
		x := make(map[string]interface{})
		if ok {
			for key, value := range input {
				x[key], err = ToJson(value)

				if err != nil {
					return nil, err
				}
			}
		} else {
			keys := value.MapKeys()

			for _, key := range keys {
				keyval := value.MapIndex(key)
				x[fmt.Sprintf("%v", key.Interface())], err = ToJson(keyval.Interface())

				if err != nil {
					return nil, err
				}
			}
		}

		return x, nil
	case reflect.Slice:
		if Debug {
			fmt.Printf("Processing %T as Slice\n", i)
		}
		var x []interface{} = make([]interface{}, value.Len())
		for idx := 0; idx < value.Len(); idx++ {
			v := value.Index(idx)
			o, err := ToJson(v.Interface())
			if err != nil {
				return nil, err
			}
			x[idx] = o
		}

		return x, nil
	case reflect.Ptr:
		if Debug {
			fmt.Printf("Processing %T as Pointer\n", i)
		}

		if !value.IsValid() || value.IsNil() {
			return nil, nil
		}

		return ToJson(reflect.Indirect(value).Interface())
	default:
		if Debug {
			fmt.Printf("Processing %T as default\n", i)
		}
		return i, nil
	}
}

func WriteToJson(w http.ResponseWriter, obj interface{}) error {
	o, err := ToJson(obj)
	if err != nil {
		writeError(err)
		return err
	}

	wrapped := map[string]interface{}{"result": o}

	return WriteJson(w, &wrapped)
}

func WriteToJsonNotWrapped(w http.ResponseWriter, obj interface{}) error {
	o, err := ToJson(obj)
	if err != nil {
		writeError(err)
		return err
	}

	return WriteJson(w, &o)
}

func WriteJson(w http.ResponseWriter, obj interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	marshalled, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		writeError(err)
		return err
	}
	w.Write([]byte(fmt.Sprintf("%s\n", marshalled)))

	return nil
}
