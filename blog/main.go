package main

import (
	"fanfan926.icu/msgo/v2"
	"fmt"
)

func main() {
	e := msgo.New()
	userRg := e.Route.Group("user")
	userRg.Get("/info", func(ctx *msgo.Context) {
		fmt.Fprintln(ctx.W, "get hello")
	})
	userRg.Post("/info", func(ctx *msgo.Context) {
		fmt.Fprintln(ctx.W, "post hello")
	})

	userRg.Post("/login", func(ctx *msgo.Context) {
		fmt.Fprintln(ctx.W, "post login")
	})

	e.Run()

}
