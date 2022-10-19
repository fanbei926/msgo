package render

import (
	"encoding/json"
	"net/http"
)

type MyJson struct {
	Data any
}

func (j *MyJson) WriteContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

func (j *MyJson) Render(w http.ResponseWriter, status int) error {
	if status == 200 {
		w.WriteHeader(status)
	}

	j.WriteContentType(w, "application/json; charset=utf-8")
	jsonData, err := json.Marshal(j.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonData)
	return err
}
