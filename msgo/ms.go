package msgo

import (
	"fanfan926.icu/msgo/v2/render"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)

const ANY = "ANY"

type HandleFunc func(context *Context)
type MiddlewareHandleFunc func(handle HandleFunc) HandleFunc

var total int64

// eg: /user/login /user/register , user is a routeGroup
type routeGroup struct {
	name              string                           // eg: user
	handleFuncMap     map[string]map[string]HandleFunc // eg: login get:loginFunc
	middlewareFuncMap map[string]map[string][]MiddlewareHandleFunc
	handleMethodMap   map[string][]string    // eg: get {/login, /register} TODO: delete it
	middlewareFunc    []MiddlewareHandleFunc // eg:
	routeTree         *treeNode
}

func (rg *routeGroup) registerHandleFuncMap(path, method string, f HandleFunc, mf ...MiddlewareHandleFunc) {
	_, ok := rg.handleFuncMap[path]
	if !ok {
		rg.handleFuncMap[path] = make(map[string]HandleFunc)
		rg.middlewareFuncMap[path] = make(map[string][]MiddlewareHandleFunc)
	}
	_, ok = rg.handleFuncMap[path][method]
	if ok {
		panic("Not allowed repeated route")
	}
	rg.handleFuncMap[path][method] = f
	rg.handleMethodMap[method] = append(rg.handleMethodMap[method], path)
	rg.middlewareFuncMap[path][method] = append(rg.middlewareFuncMap[path][method], mf...)

	rg.routeTree.Put("/" + rg.name + path) // rootPath + subPath
}

// Use register a preMiddlewareFunc
func (rg *routeGroup) Use(middleFunc MiddlewareHandleFunc) {
	rg.middlewareFunc = append(rg.middlewareFunc, middleFunc)
}

// it will execute all the middlewares registered
func (rg *routeGroup) methodMiddlewareHandler(handleFunc HandleFunc, ctx *Context, path string, method string) {
	if rg.middlewareFunc != nil {
		for _, middleFunc := range rg.middlewareFunc {
			handleFunc = middleFunc(handleFunc)
		}
	}

	middleFuncs := rg.middlewareFuncMap[path][method]
	for _, middleFunc := range middleFuncs {
		handleFunc = middleFunc(handleFunc)
	}
	handleFunc(ctx)
}

// Any register a method
// eg: routeGroup{
//		"/login",
//		"{"/login": func xxx()}",
//      "{"ANY":{"/login"}}"
//}
func (rg *routeGroup) Any(path string, f HandleFunc, mf ...MiddlewareHandleFunc) {
	rg.registerHandleFuncMap(path, ANY, f, mf...)
}
func (rg *routeGroup) Get(path string, f HandleFunc, mf ...MiddlewareHandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodGet, f, mf...)
}
func (rg *routeGroup) Post(path string, f HandleFunc, mf ...MiddlewareHandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodPost, f, mf...)
}
func (rg *routeGroup) Delete(path string, f HandleFunc, mf ...MiddlewareHandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodPost, f, mf...)
}
func (rg *routeGroup) Put(path string, f HandleFunc, mf ...MiddlewareHandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodPost, f, mf...)
}
func (rg *routeGroup) Patch(path string, f HandleFunc, mf ...MiddlewareHandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodPost, f, mf...)
}
func (rg *routeGroup) Options(path string, f HandleFunc, mf ...MiddlewareHandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodPost, f, mf...)
}
func (rg *routeGroup) Head(path string, f HandleFunc, mf ...MiddlewareHandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodPost, f, mf...)
}

type route struct {
	routeGroups []*routeGroup
}

// Group register a new route group
func (r *route) Group(name string) *routeGroup {
	rg := &routeGroup{
		name:              name,
		handleFuncMap:     make(map[string]map[string]HandleFunc),
		middlewareFuncMap: make(map[string]map[string][]MiddlewareHandleFunc),
		handleMethodMap:   make(map[string][]string),
		routeTree: &treeNode{
			name:     "/",
			children: make([]*treeNode, 0),
		},
	}

	r.routeGroups = append(r.routeGroups, rg)
	return rg
}

type Engine struct {
	Route   *route
	funcMap template.FuncMap
	render  *render.HTMLRender
	pool    sync.Pool
}

func New() *Engine {
	engine := &Engine{
		Route: &route{},
	}
	engine.pool.New = func() any {
		return engine.allocateContext()
	}
	return engine
}

func (e *Engine) allocateContext() any {
	atomic.AddInt64(&total, 1)
	return &Context{engine: e}
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) SetHTMLTemplate(t *template.Template) {
	e.render = &render.HTMLRender{
		Template: t,
	}
}

func (e *Engine) LoadTemplate(pattern string) {
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.SetHTMLTemplate(t)
}

// main function
func (e *Engine) httpRequestHandle(w http.ResponseWriter, r *http.Request, ctx *Context) {
	method := r.Method
	for _, rg := range e.Route.routeGroups {
		node := rg.routeTree.Get(r.URL.Path)
		if node != nil && node.isEnd == true {
			uri := SubStringLast(node.rootPath, rg.name) // name must contain a slash
			handle, ok := rg.handleFuncMap[uri]["ANY"]
			if ok {
				rg.methodMiddlewareHandler(handle, ctx, uri, "ANY")
				return
			}

			handle, ok = rg.handleFuncMap[uri][method]
			if ok {
				rg.methodMiddlewareHandler(handle, ctx, uri, method)
				return
			}
			w.WriteHeader(405)
			fmt.Fprintf(w, "%s not supported", method)
			return
		}
	}

	w.WriteHeader(404) // If not found, return 404
	fmt.Fprintf(w, "%s not found", r.RequestURI)
	return
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := e.pool.Get().(*Context)
	ctx.W = w
	ctx.R = r
	e.httpRequestHandle(w, r, ctx)
	e.pool.Put(ctx)

	fmt.Println(total)
}

func (e *Engine) Run() {
	http.Handle("/", e)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
