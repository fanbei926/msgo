package msgo

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

type Accounts struct {
	UnAuthHandle func(ctx *Context)
	Users        map[string]string
	realm        string
}

func (a *Accounts) BasicAuth(next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		username, password, ok := ctx.R.BasicAuth()
		fmt.Println("username", username, "password", password, ok)
		if !ok {
			a.unAuthHandle(ctx)
			return
		}
		pwd, exist := a.Users[username]
		if !exist {
			a.unAuthHandle(ctx)
			return
		}
		if pwd != password {
			a.unAuthHandle(ctx)
			return
		}
		ctx.Set("user", username)
		next(ctx)
	}
}

func (a *Accounts) unAuthHandle(ctx *Context) {
	if a.UnAuthHandle != nil {
		a.UnAuthHandle(ctx)
	} else {
		ctx.W.Header().Set("WWW-Authenticate", a.realm)
		ctx.W.WriteHeader(http.StatusUnauthorized)
	}
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
