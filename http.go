package toJson

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteToJson(w http.ResponseWriter, obj interface{}) error {
	return WriteToJsonWithCode(w, obj, http.StatusOK)
}

func WriteToJsonWithCode(w http.ResponseWriter, obj interface{}, code int) error {
	o, err := ToJson(obj)
	if err != nil {
		writeError(err)
		return err
	}

	wrapped := map[string]interface{}{"result": o}

	return WriteJson(w, &wrapped, code)
}

func WriteToJsonNotWrapped(w http.ResponseWriter, obj interface{}) error {
	return WriteToJsonNotWrappedWithCode(w, obj, http.StatusOK)
}

func WriteToJsonNotWrappedWithCode(w http.ResponseWriter, obj interface{}, code int) error {
	o, err := ToJson(obj)
	if err != nil {
		writeError(err)
		return err
	}

	return WriteJson(w, &o, code)
}

func WriteJson(w http.ResponseWriter, obj interface{}, code int) error {
	marshalled, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		writeError(err)
		return err
	}

	data := []byte(fmt.Sprintf("%s\n", marshalled))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	w.WriteHeader(code)
	w.Write(data)

	return nil
}
