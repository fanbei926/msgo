package msgo

import (
	"fmt"
	"log"
	"net/http"
)

type HandleFunc func(w http.ResponseWriter, r *http.Request)

// eg: /user/login /user/register , user is a routeGroup
type routeGroup struct {
	name            string                // eg: user
	handleFuncMap   map[string]HandleFunc // eg: login loginFunc
	handleMethodMap map[string][]string   // eg: get {/login, /register}
}

// Any register a method
// eg: routeGroup{
//		"/login",
//		"{"/login": func xxx()}",
//      "{"ANY":{"/login"}}"
//}
func (rg *routeGroup) Any(name string, f HandleFunc) {
	rg.handleFuncMap[name] = f
	rg.handleMethodMap["ANY"] = append(rg.handleMethodMap["ANY"], name)
}

func (rg *routeGroup) Get(name string, f HandleFunc) {
	rg.handleFuncMap[name] = f
	rg.handleMethodMap[http.MethodGet] = append(rg.handleMethodMap[http.MethodGet], name)
}

func (rg *routeGroup) Post(name string, f HandleFunc) {
	rg.handleFuncMap[name] = f
	rg.handleMethodMap[http.MethodPost] = append(rg.handleMethodMap[http.MethodPost], name)
}

type route struct {
	routeGroups []*routeGroup
}

// Group register a new route group
func (r *route) Group(name string) *routeGroup {
	rg := &routeGroup{
		name:            name,
		handleFuncMap:   make(map[string]HandleFunc),
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
		for name, methodHandle := range rg.handleFuncMap { // name: "/login", handle: func
			url := fmt.Sprintf("/%s%s", rg.name, name)
			if url == r.RequestURI { // please don't use else, beacuse u must let it loops
				routes, ok := rg.handleMethodMap["ANY"]
				if ok {
					for _, routeName := range routes {
						if routeName == name {
							methodHandle(w, r)
							return
						}
					}
				}

				routes, ok = rg.handleMethodMap[method]
				if ok {
					for _, routeName := range routes {
						if routeName == name {
							methodHandle(w, r)
							return
						}
					}
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
