package main

import (
	"fanfan926.icu/msgo/v2"
	"fmt"
	"net/http"
)

func main() {
	e := msgo.New()
	userRg := e.Route.Group("user")
	userRg.Get("/info", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	})
	userRg.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "login")
	})

	e.Run()

}
