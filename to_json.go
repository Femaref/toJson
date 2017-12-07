package toJson

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/serenize/snaker"

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

type props struct {
	Name      string
	Omitempty bool
	Omit      bool
}

func fieldProperties(field reflect.StructField) props {
	value := field.Tag.Get("json")
	fields := strings.Split(value, ",")

	// The format of the json tag is "<field>,<options>", with fields possibly being
	// empty
	if value != "" {
		name := fields[0]
		opts := fields[1:]
		p := props{
			Name: name,
			Omit: name == "-",
		}

		for _, opt := range opts {
			if opt == "omitempty" {
				p.Omitempty = true
			}
		}

		return p
	}
	return props{Name: snaker.CamelToSnake(field.Name)}
}

// copied from encoding/json
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
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

			p := fieldProperties(def)

			// same behaviour as encoding/json
			if p.Omit || (p.Omitempty && isEmptyValue(val)) {
				continue
			}

			res, err := ToJson(val.Interface())

			if err != nil {
				return nil, err
			}

			x[p.Name] = res
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
