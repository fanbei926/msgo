package msgo

import (
	"fmt"
	"log"
	"net/http"
)

type HandleFunc func(context *Context)

type Context struct {
	W http.ResponseWriter
	R *http.Request
}

// eg: /user/login /user/register , user is a routeGroup
type routeGroup struct {
	name            string                           // eg: user
	handleFuncMap   map[string]map[string]HandleFunc // eg: login loginFunc
	handleMethodMap map[string][]string              // eg: get {/login, /register}
}

func (rg *routeGroup) registerHandleFuncMap(name, method string, f HandleFunc) {
	_, ok := rg.handleFuncMap[name]
	if !ok {
		rg.handleFuncMap[name] = make(map[string]HandleFunc)
	}
	_, ok = rg.handleFuncMap[name][method]
	if ok {
		panic("Not allowed repeated route")
	}
	rg.handleFuncMap[name][method] = f
	rg.handleMethodMap[method] = append(rg.handleMethodMap[method], name)
}

// Any register a method
// eg: routeGroup{
//		"/login",
//		"{"/login": func xxx()}",
//      "{"ANY":{"/login"}}"
//}
func (rg *routeGroup) Any(name string, f HandleFunc) {
	rg.registerHandleFuncMap(name, "ANY", f)
}

func (rg *routeGroup) Get(name string, f HandleFunc) {
	rg.registerHandleFuncMap(name, http.MethodGet, f)
}

func (rg *routeGroup) Post(name string, f HandleFunc) {
	rg.registerHandleFuncMap(name, http.MethodPost, f)
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
		for name, methodHandleMap := range rg.handleFuncMap { // name: "/login", handle: func
			url := fmt.Sprintf("/%s%s", rg.name, name)
			if url == r.RequestURI { // please don't use else, beacuse u must let it loops
				ctx := &Context{
					W: w,
					R: r,
				}
				handle, ok := methodHandleMap["ANY"]
				if ok {
					handle(ctx)
					return
				}

				handle, ok = methodHandleMap[method]
				if ok {
					handle(ctx)
					return
				}
				w.WriteHeader(405)
				fmt.Fprintf(w, "%s not supported", method)
				return
			}
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
