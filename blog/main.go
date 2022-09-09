package main

import (
	"fanfan926.icu/msgo/v2"
	fmt "fmt"
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

	userRg.Post("/first/:id", func(ctx *msgo.Context) {
		fmt.Fprintln(ctx.W, "11 any id")
	})

	userRg.Post("/hello/*/:id", func(ctx *msgo.Context) {
		fmt.Fprintln(ctx.W, "222 any id")
	})

	userRg.Post("/hello1/xxx", func(ctx *msgo.Context) {
		fmt.Fprintln(ctx.W, "xxx")
	})

	e.Run()

}
