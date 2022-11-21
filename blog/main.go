package main

import (
	"fanfan926.icu/msgo/v2"
	"fanfan926.icu/msgo/v2/mspool"
	"fanfan926.icu/msgo/v2/token"
	fmt "fmt"
	"log"
	"net/http"
	"time"
)

type User struct {
	Name      string   `json:"name" xml:"name" msgo:"required"`
	Age       int      `json:"age" xml:"age" validate:"required,max=40,min=10"`
	Addresses []string `json:"addresses" xml:"addresses"`
	Email     string   `json:"email" msgo:"required"`
}

func Log(next msgo.HandleFunc) msgo.HandleFunc {
	return func(ctx *msgo.Context) {
		fmt.Println("print args")
		next(ctx)
		fmt.Println("return")
	}
}

func main() {
	x := []int{0, 1, 2, 3}
	fmt.Println(x[:1])
	e := msgo.Default()
	//fmt.Println(msgo.BasicAuth("fane", "123456"))
	//auth := &msgo.Accounts{
	//	Users: make(map[string]string),
	//}
	//
	//auth.Users["fane"] = "123456"
	jh := &token.JWTHandler{Key: []byte("123456")}
	e.Use(jh.AuthInterceptor)
	//e.Use(auth.BasicAuth)
	userRg := e.Route.Group("user")
	//userRg.Use(msgo.Logging)
	//userRg.Use(msgo.Recovery)

	//userRg.Use(func(next msgo.HandleFunc) msgo.HandleFunc {
	//	return func(ctx *msgo.Context) {
	//		fmt.Println("xxxxxx")
	//		next(ctx)
	//	}
	//})
	//
	//userRg.Use(func(next msgo.HandleFunc) msgo.HandleFunc {
	//	return func(ctx *msgo.Context) {
	//		fmt.Println("yyyy")
	//		next(ctx)
	//		fmt.Println("post middle")
	//	}
	//})
	//e.Logger.Level = msLog.LevelDebug
	//logger.Outs = append(logger.Outs, msLog.FileWriter("./log/log.log"))
	//e.Logger.SetLogPath("./log")
	//e.Logger.LogFileSize = 1 << 10 //1k
	//var u *User
	p, err := mspool.NewPool(3)
	if err != nil {
		fmt.Println(err)
	}
	userRg.Get("/info", func(ctx *msgo.Context) {
		//u.Age = 10
		//ctx.Logger.WithFields(msLog.Fields{
		//	"name": "fkdyy",
		//	"age":  1000,
		//}).Debug("Debug")
		//ctx.Logger.Info("Info")
		//ctx.Logger.Error("Error")
		p.Submit(func() {
			fmt.Println("11111111111111")
			//panic("xxx")
			time.Sleep(8 * time.Second)
		})

		p.Submit(func() {
			fmt.Println("22222")
			time.Sleep(5 * time.Second)
		})
		fmt.Println("------")
		p.Submit(func() {
			fmt.Println("3333")
			time.Sleep(5 * time.Second)
		})
		p.Submit(func() {
			fmt.Println("4444")
			time.Sleep(4 * time.Second)
		})

		fmt.Fprintln(ctx.W, "get hello")
	})
	userRg.Post("/info", func(ctx *msgo.Context) {
		fmt.Fprintln(ctx.W, "post hello")
	})

	userRg.Get("/html", func(ctx *msgo.Context) {
		ctx.Html("<html>hello,html</html>", 200)
	})

	userRg.Get("/htmlTemplate", func(ctx *msgo.Context) {
		user := &User{
			Name: "lalal",
		}
		err := ctx.HtmlTemplate("index.html", user, "template/index.html", "template/header.html")
		if err != nil {
			fmt.Println(err)
		}
	})

	userRg.Get("/htmlTemplateGlob", func(ctx *msgo.Context) {
		user := &User{
			Name: "lalal",
		}
		err := ctx.HtmlTemplateGlob("login.html", user, "template/*.html")
		if err != nil {
			fmt.Println(err)
		}
	})

	userRg.Get("/login", func(ctx *msgo.Context) {
		jwt := &token.JWTHandler{}
		jwt.Key = []byte("123456")
		jwt.SendCookie = true
		jwt.TimeOut = 10 * time.Minute
		jwt.RefreshTimeOut = 20 * time.Minute
		jwt.Authenticator = func(ctx *msgo.Context) (map[string]any, error) {
			data := make(map[string]any)
			data["userId"] = 1
			return data, nil
		}
		token, err := jwt.LoginHandler(ctx)
		if err != nil {
			log.Println(err)
			ctx.JSON(http.StatusOK, err.Error())
		}
		ctx.JSON(http.StatusOK, token)

	}, Log)

	userRg.Get("/refresh", func(ctx *msgo.Context) {
		jwt := &token.JWTHandler{}
		jwt.Key = []byte("123456")
		jwt.SendCookie = true
		jwt.TimeOut = 10 * time.Minute
		jwt.RefreshTimeOut = 20 * time.Minute

		jwt.RefreshKey = "blog_refresh_token"
		ctx.Set(jwt.RefreshKey, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NjkwMTgwOTAsImlhdCI6MTY2OTAxNjg5MCwidXNlcklkIjoxfQ.fH3Vrv3o2t9j3KMg65mkDXJqLwDBe1MIpQBeYxNxFZ4")
		token, err := jwt.RefreshHandler(ctx)
		if err != nil {
			log.Println(err)
			ctx.JSON(http.StatusOK, err.Error())
		}
		ctx.JSON(http.StatusOK, token)

	}, Log)

	userRg.Post("/first/:id", func(ctx *msgo.Context) {
		fmt.Fprintln(ctx.W, "11 any id")
	})

	userRg.Post("/hello/*/:id", func(ctx *msgo.Context) {
		fmt.Fprintln(ctx.W, "222 any id")
	})

	userRg.Post("/hello1/xxx", func(ctx *msgo.Context) {
		fmt.Fprintln(ctx.W, "xxx")
	})

	e.LoadTemplate("template/*.html")
	userRg.Get("/template", func(ctx *msgo.Context) {
		user := &User{
			Name: "dyy",
		}
		err := ctx.Template("login.html", user, http.StatusOK)
		if err != nil {
			fmt.Println(err)
		}
	})

	userRg.Get("/json", func(ctx *msgo.Context) {
		user := &User{
			Name: "fkdyy",
			Age:  29,
		}
		ctx.JSON(200, user)
	})

	userRg.Get("/xml", func(ctx *msgo.Context) {
		user := &User{
			Name: "fkdyy",
			Age:  29,
		}
		ctx.XML(200, user)
	})

	userRg.Get("/excel", func(ctx *msgo.Context) {
		ctx.File("template/Book1.xlsx")
	})

	userRg.Get("/book", func(ctx *msgo.Context) {
		ctx.FileAttachment("template/Book1.xlsx", "a?aa.xlsx")
	})

	userRg.Get("/fs", func(ctx *msgo.Context) {
		ctx.FileFromFS("Book1.xlsx", http.Dir("template"))
	})

	userRg.Get("/redirect", func(ctx *msgo.Context) {
		ctx.Redirect(302, "/user/xml")
	})

	userRg.Get("/string", func(ctx *msgo.Context) {
		ctx.String(http.StatusOK, "%v fuck  %v", "ff", "dyy")
	})

	userRg.Get("/query", func(ctx *msgo.Context) {
		fmt.Println(ctx.GetQuery("id"))
	})

	userRg.Get("/queryDefault", func(ctx *msgo.Context) {
		fmt.Println(ctx.GetDefaultQuery("id", "fkdyy"))
	})

	userRg.Get("/queryMap", func(ctx *msgo.Context) {
		tmp, _ := ctx.GetQueryMap("name")
		ctx.JSON(200, tmp)
	})

	userRg.Get("/queryArray", func(ctx *msgo.Context) {
		values, ok := ctx.GetQueryArray("id")
		if ok {
			fmt.Println(values)
		} else {
			fmt.Println("No values found")
		}
	})

	userRg.Post("/formPost", func(ctx *msgo.Context) {
		m, _ := ctx.GetPostFormMap("dyy")
		ctx.JSON(200, m)
	})

	userRg.Post("/formFile", func(ctx *msgo.Context) {
		fileHeader := ctx.FormFile("dyy")
		ctx.SaveUploadFile(fileHeader, "fkdyy.xml")
		ctx.JSON(200, "ok")
	})

	userRg.Post("/formFiles", func(ctx *msgo.Context) {
		fileHeaders := ctx.FormFiles("dyy")
		for _, fileHeader := range fileHeaders {
			err := ctx.SaveUploadFile(fileHeader, fileHeader.Filename)
			if err != nil {
				fmt.Println(err)
			}
		}
		ctx.JSON(200, "ok")
	})

	userRg.Post("/dealJson", func(ctx *msgo.Context) {
		//user := &User{}
		user := make([]User, 0)
		ctx.DisallowUnknownFields = true
		ctx.IsValidate = false
		err := ctx.BindJson(&user)
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, err)
			return
		}
		ctx.JSON(200, "ok")
	})

	userRg.Post("/dealXML", func(ctx *msgo.Context) {
		user := &User{}
		err := ctx.BindXML(&user)
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, err)
			return
		}
		ctx.JSON(200, "ok")
	})

	//e.Run()
	e.RunTLS(":8118", "key/server.pem", "key/server.key")
}
