package toJson

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/serenize/snaker"
	"net/http"
	"reflect"
	"unicode"
	"unicode/utf8"
)

type jsonWriter interface {
	ToJson() interface{}
}

type JsonWriter jsonWriter

var Debug bool = false

var TypeError = errors.New("Could not get type")
var UnsupportedDatatype = errors.New("Unsupported Datatype")
var ReflectTypeError = errors.New("Got passed reflect.Type")
var ReflectValueError = errors.New("Got passed reflect.Value")

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
		return nil, TypeError
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
			runeValue, _ := utf8.DecodeRuneInString(def.Name)
			if !unicode.IsUpper(runeValue) {
				continue
			}

			val := value.Field(idx)

			res, err := ToJson(val.Interface())

			if err != nil {
				return nil, err
			}

			x[snaker.CamelToSnake(def.Name)] = res
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
		var x []interface{} = make([]interface{}, 0)
		for idx := 0; idx < value.Len(); idx++ {
			v := value.Index(idx)
			o, err := ToJson(v.Interface())
			if err != nil {
				return nil, err
			}
			x = append(x, o)
		}

		return x, nil
	case reflect.Ptr:
		if Debug {
			fmt.Printf("Processing %T as Pointer\n", i)
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
		return err
	}
	return WriteJson(w, &o)
}

func WriteJson(w http.ResponseWriter, obj interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	marshalled, err := json.MarshalIndent(map[string]interface{}{"result": obj}, "", "  ")
	if err != nil {
		return err
	}
	w.Write([]byte(fmt.Sprintf("%s\n", marshalled)))

	return nil
}
