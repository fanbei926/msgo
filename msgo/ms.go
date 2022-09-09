package msgo

import (
	"fmt"
	"log"
	"net/http"
)

const ANY = "ANY"

type HandleFunc func(context *Context)

type Context struct {
	W http.ResponseWriter
	R *http.Request
}

// eg: /user/login /user/register , user is a routeGroup
type routeGroup struct {
	name            string                           // eg: user
	handleFuncMap   map[string]map[string]HandleFunc // eg: login get:loginFunc
	handleMethodMap map[string][]string              // eg: get {/login, /register}
	routeTree       *treeNode
}

func (rg *routeGroup) registerHandleFuncMap(path, method string, f HandleFunc) {
	_, ok := rg.handleFuncMap[path]
	if !ok {
		rg.handleFuncMap[path] = make(map[string]HandleFunc)
	}
	_, ok = rg.handleFuncMap[path][method]
	if ok {
		panic("Not allowed repeated route")
	}
	rg.handleFuncMap[path][method] = f
	rg.handleMethodMap[method] = append(rg.handleMethodMap[method], path)

	rg.routeTree.Put("/" + rg.name + path) // rootPath + subPath
}

// Any register a method
// eg: routeGroup{
//		"/login",
//		"{"/login": func xxx()}",
//      "{"ANY":{"/login"}}"
//}
func (rg *routeGroup) Any(path string, f HandleFunc) {
	rg.registerHandleFuncMap(path, ANY, f)
}
func (rg *routeGroup) Get(path string, f HandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodGet, f)
}
func (rg *routeGroup) Post(path string, f HandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodPost, f)
}
func (rg *routeGroup) Delete(path string, f HandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodPost, f)
}
func (rg *routeGroup) Put(path string, f HandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodPost, f)
}
func (rg *routeGroup) Patch(path string, f HandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodPost, f)
}
func (rg *routeGroup) Options(path string, f HandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodPost, f)
}
func (rg *routeGroup) Head(path string, f HandleFunc) {
	rg.registerHandleFuncMap(path, http.MethodPost, f)
}

type route struct {
	routeGroups []*routeGroup
}

// Group register a new route group
func (r *route) Group(name string) *routeGroup {
	rg := &routeGroup{
		name:            name,
		handleFuncMap:   make(map[string]map[string]HandleFunc),
		handleMethodMap: make(map[string][]string),
		routeTree: &treeNode{
			name:     "/",
			children: make([]*treeNode, 0),
		},
	}

	r.routeGroups = append(r.routeGroups, rg)
	return rg
}

type Engine struct {
	Route *route
}

func New() *Engine {
	return &Engine{
		Route: &route{},
	}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	for _, rg := range e.Route.routeGroups {
		node := rg.routeTree.Get(r.RequestURI)
		if node != nil && node.isEnd == true {
			ctx := &Context{
				W: w,
				R: r,
			}
			uri := SubStringLast(node.rootPath, rg.name) // name must contain a slash
			handle, ok := rg.handleFuncMap[uri]["ANY"]
			if ok {
				handle(ctx)
				return
			}

			handle, ok = rg.handleFuncMap[uri][method]
			if ok {
				handle(ctx)
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

func (e *Engine) Run() {
	http.Handle("/", e)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
