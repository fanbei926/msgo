package render

import (
	"fanfan926.icu/msgo/v2/internal"
	"fmt"
	"net/http"
)

type MyString struct {
	Format string
	Data   []any
}

func (s *MyString) Render(w http.ResponseWriter, status int) error {
	w.WriteHeader(status)
	s.WriteContentType(w, "text/plain; charset=utf-8")
	if len(s.Data) > 0 {
		_, err := fmt.Fprintf(w, s.Format, s.Data...)
		return err
	}
	_, err := w.Write(internal.String2Bytes(s.Format))
	return err
}

func (s *MyString) WriteContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}
