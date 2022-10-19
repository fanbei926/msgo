package render

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type MyXML struct {
	Data any
}

func (x *MyXML) WriteContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

func (x *MyXML) Render(w http.ResponseWriter, status int) error {
	w.WriteHeader(status)
	x.WriteContentType(w, "application/xml; charset=utf-8")
	b, err := ioutil.ReadFile("/Users/4work/GolandProjects/awesomeProject/msgo/test.xml")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
	_, err = fmt.Fprintln(w, string(b))
	return err
	//return xml.NewEncoder(w).Encode(x.Data)
}
