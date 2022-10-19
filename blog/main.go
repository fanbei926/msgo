package main

import (
	"fanfan926.icu/msgo/v2"
	msLog "fanfan926.icu/msgo/v2/log"
	fmt "fmt"
	"net/http"
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
	e := msgo.New()
	logger := msLog.Default()
	userRg := e.Route.Group("user")
	userRg.Use(msgo.Logging)

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
	logger.Level = msLog.LevelDebug
	userRg.Get("/info", func(ctx *msgo.Context) {
		logger.Debug("Debug")
		logger.Info("Info")
		logger.Error("Error")
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

	userRg.Post("/login", func(ctx *msgo.Context) {
		fmt.Fprintln(ctx.W, "post login")
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

	e.Run()

}
