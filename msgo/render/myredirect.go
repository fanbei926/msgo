package render

import (
	"errors"
	"fmt"
	"net/http"
)

type MyRedirect struct {
	StatusCode int
	Url        string
	Request    *http.Request
}

func (r *MyRedirect) WriteContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

func (r *MyRedirect) Render(w http.ResponseWriter, status int) error {
	r.WriteContentType(w, "text/html; charset=utf-8")
	w.WriteHeader(status)
	if status < http.StatusMultiStatus || status > http.StatusPermanentRedirect && status != http.StatusCreated {
		return errors.New(fmt.Sprintf("Can not redirect with status code %d", status))
	}
	http.Redirect(w, r.Request, r.Url, status)
	return nil
}
