package toJson

import (
    "encoding/json"
    "reflect"
    "net/http"
    "errors"
    "github.com/serenize/snaker"
    "unicode"
    "unicode/utf8"
    "fmt"
)

type jsonWriter interface {
    ToJson() (interface{})
}

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
        return nil, nil
    }
    
    value := reflect.ValueOf(i)
       
    writer, ok := i.(jsonWriter)

    if ok {
        return ToJson(writer.ToJson())
    }
    
    marshaler, ok := i.(json.Marshaler)
    
    if ok {
        return marshaler, nil
    }
    
    switch t.Kind() {
        case reflect.Struct:
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
                input, ok := i.(map[string]interface{})
    
                if ok {
                    x := make(map[string]interface{})
                    for key, value := range input {
                        x[key], err = ToJson(value)
            
                        if err != nil {
                            return nil, err
                        }
                    }
        
                    return x, nil
                }
                return nil, UnsupportedDatatype
        case reflect.Slice:
            var x []interface{} = make([]interface{}, 0)
            for idx := 0; idx< value.Len(); idx++ {
                v := value.Index(idx)
                o, err := ToJson(v.Interface())
                if err != nil {
                    return nil, err
                }
                x = append(x, o)
            }
        
            return x, nil
        default:        
            return i, nil
    }
}

func WriteToJson(w http.ResponseWriter, obj interface{}) error {
    o , err:= ToJson(obj)
    if err != nil { 
        return err
    }
    return WriteJson(w, &o)
}

func WriteJson(w http.ResponseWriter, obj interface{}) error {
    w.Header().Add("Content-Type", "application/json")
	marshalled, err := json.MarshalIndent(map[string]interface{}{"result":obj}, "", "  ")
	if err != nil {
		return err
	}
	w.Write([]byte(fmt.Sprintf("%s\n", marshalled)))

	return nil
}