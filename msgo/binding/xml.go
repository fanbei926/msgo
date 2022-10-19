package binding

import (
	"encoding/xml"
	"net/http"
)

type xmlBinding struct {
}

func (b xmlBinding) Name() string {
	return "xml"
}

func (b xmlBinding) Bind(r *http.Request, object any) error {
	if r.Body == nil {
		return nil
	}
	decoder := xml.NewDecoder(r.Body)
	err := decoder.Decode(object)
	if err != nil {
		return err
	}
	return validate(object)
}
