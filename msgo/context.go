package msgo

import (
	"errors"
	"fanfan926.icu/msgo/v2/binding"
	"fanfan926.icu/msgo/v2/render"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const defaultMaxMemory = 32 << 20 // 32 mb

type Context struct {
	W                     http.ResponseWriter
	R                     *http.Request
	engine                *Engine
	StatusCode            int
	queryCache            url.Values
	formCache             url.Values
	DisallowUnknownFields bool
	IsValidate            bool
}

func (ctx *Context) FormFile(key string) *multipart.FileHeader {
	file, header, err := ctx.R.FormFile(key)
	if err != nil {
		fmt.Fprintln(ctx.W, err)
	}
	defer file.Close()
	return header
}

func (ctx *Context) FormFiles(key string) []*multipart.FileHeader {
	form, err := ctx.MultipartForm()
	if err != nil {
		fmt.Println(err)
	}
	return form.File[key]
}

func (ctx *Context) MultipartForm() (*multipart.Form, error) {
	err := ctx.R.ParseMultipartForm(defaultMaxMemory)
	return ctx.R.MultipartForm, err
}

func (ctx *Context) initPostFormCache() {
	if ctx.R != nil {
		if err := ctx.R.ParseMultipartForm(defaultMaxMemory); err != nil {
			if errors.Is(err, http.ErrNotMultipart) {
				fmt.Println(err)
			}
		}
		ctx.formCache = ctx.R.PostForm
	} else {
		ctx.formCache = url.Values{}
	}
}

func (ctx *Context) GetPostForm(key string) (string, bool) {
	values, ok := ctx.GetPostFormArray(key)
	if ok {
		return values[0], ok
	}
	return "", false
}

func (ctx *Context) GetPostFormArray(key string) ([]string, bool) {
	ctx.initPostFormCache()
	values, ok := ctx.formCache[key]
	return values, ok
}

func (ctx *Context) PostFormMap(key string) map[string]string {
	dicts, _ := ctx.GetPostFormMap(key)
	return dicts

}

func (ctx *Context) GetPostFormMap(key string) (map[string]string, bool) {
	ctx.initPostFormCache()
	return ctx.get(ctx.formCache, key)

}

func (ctx *Context) initQueryCache() {
	if ctx.R != nil {
		ctx.queryCache = ctx.R.URL.Query()
	} else {
		ctx.queryCache = url.Values{}
	}
}

func (ctx *Context) GetDefaultQuery(key string, defaultValue string) string {
	values, ok := ctx.GetQueryArray(key)
	if ok {
		return values[0]
	}
	return defaultValue
}

func (ctx *Context) GetQuery(key string) string {
	ctx.initQueryCache()
	return ctx.queryCache.Get(key)
}

func (ctx *Context) GetQueryArray(key string) ([]string, bool) {
	ctx.initQueryCache()
	values, ok := ctx.queryCache[key]
	return values, ok
}

func (ctx *Context) QueryMap(key string) map[string]string {
	dicts, _ := ctx.GetQueryMap(key)
	return dicts

}

func (ctx *Context) GetQueryMap(key string) (map[string]string, bool) {
	ctx.initQueryCache()
	return ctx.get(ctx.queryCache, key)

}

func (ctx *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	dists := make(map[string]string)
	exist := false
	for k, value := range m {
		if i := strings.IndexByte(k, '['); i >= 1 && k[:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dists[k[i+1:][:j]] = value[0]
			}
		}
	}

	return dists, exist
}

func (ctx *Context) Html(data string, status int) error {
	return ctx.Render(status, &render.MyTemplate{Data: data, Name: "", IsTemplate: false})
}

func (ctx *Context) HtmlTemplate(name string, data interface{}, filenames ...string) error {
	ctx.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.New(name)
	t, err := t.ParseFiles(filenames...)
	if err != nil {
		return err
	}
	err = t.Execute(ctx.W, data)
	return err
}

func (ctx *Context) HtmlTemplateGlob(name string, data interface{}, pattern string) error {
	ctx.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.New(name)
	t, err := t.ParseGlob(pattern)
	if err != nil {
		return err
	}
	err = t.Execute(ctx.W, data)
	return err
}

func (ctx *Context) Template(name string, data interface{}, status int) error {
	return ctx.Render(status, &render.MyTemplate{Data: data, Name: name, Template: ctx.engine.render.Template, IsTemplate: true})
}

func (ctx *Context) JSON(status int, data interface{}) error {
	return ctx.Render(status, &render.MyJson{Data: data})
}

func (ctx *Context) XML(status int, data any) error {
	return ctx.Render(status, &render.MyXML{Data: data})
}

func (ctx *Context) File(filename string) {
	http.ServeFile(ctx.W, ctx.R, filename)
}

func (ctx *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		fmt.Println(filename)
		ctx.W.Header().Set("Content-Disposition", "attachment; filename="+filename)
	} else {
		fmt.Println(filename)
		ctx.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8`+url.QueryEscape(filename))
	}
	http.ServeFile(ctx.W, ctx.R, filepath)
}

func (ctx *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		ctx.R.URL.Path = old
	}(ctx.R.URL.Path)

	ctx.R.URL.Path = filepath
	http.FileServer(fs).ServeHTTP(ctx.W, ctx.R)
}

func (ctx *Context) Redirect(status int, url string) error {
	return ctx.Render(status, &render.MyRedirect{StatusCode: status, Url: url, Request: ctx.R})
}

func (ctx *Context) String(status int, format string, values ...any) error {
	return ctx.Render(status, &render.MyString{Data: values, Format: format})
}

// Render all render will use it
func (ctx *Context) Render(status int, r render.Render) error {
	ctx.StatusCode = status
	return r.Render(ctx.W, status)
}

func (ctx *Context) SaveUploadFile(fileHeader *multipart.FileHeader, dstFileName string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create("./" + dstFileName)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	return nil
}

func (ctx *Context) BindJson(object any) error {
	json := binding.JSONBinding
	json.DisallowUnknownFields = false
	json.IsValidate = false

	return ctx.MustBindWith(object, json)
}

func (ctx *Context) BindXML(object any) error {
	return ctx.MustBindWith(object, binding.XMLBinding)
}

func (ctx *Context) MustBindWith(object any, b binding.Binding) error {
	if err := ctx.ShouldBind(object, b); err != nil {
		//ctx.W.WriteHeader(http.StatusBadRequest)
		return err
	}
	return nil
}

func (ctx *Context) ShouldBind(object any, b binding.Binding) error {
	return b.Bind(ctx.R, object)
}
