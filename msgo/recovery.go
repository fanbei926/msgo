package msgo

import (
	"fmt"
	"net/http"
)

func Recovery(next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(ctx.Logger.Outs)
				ctx.Logger.Error(err)
				ctx.Fail(http.StatusInternalServerError, "Internal server error-----")
			}
		}()
		next(ctx)
	}

}
