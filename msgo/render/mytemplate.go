package render

import (
	"fanfan926.icu/msgo/v2/internal"
	"html/template"
	"net/http"
)

type MyTemplate struct {
	Name       string
	Data       any
	Template   *template.Template
	IsTemplate bool
}

func (m *MyTemplate) WriteContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

func (m *MyTemplate) Render(w http.ResponseWriter, status int) error {
	m.WriteContentType(w, "text/html; charset=utf-8")
	w.WriteHeader(status)
	if m.IsTemplate {
		return m.Template.ExecuteTemplate(w, m.Name, m.Data)

	}
	_, err := w.Write(internal.String2Bytes(m.Data.(string)))
	return err
}
