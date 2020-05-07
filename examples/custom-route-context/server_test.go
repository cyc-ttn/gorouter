package main

import (
	"net/http"

	. "github.com/cyc-ttn/gorouter"
)

type Services struct{}

type Server struct {
	S *Services
	R *RouterNode
}

type CustomRouteContext struct {
	RouteContext
	S *Services
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &RouteContext{
		W:      w,
		R:      r,
		Method: r.Method,
		Path:   r.URL.Path,
	}
	route, err := s.R.Match(r.Method, r.URL.Path, ctx)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	route.GetHandler()(&CustomRouteContext{
		RouteContext: *ctx,
		S:            s.S,
	})
}

type CustomHandlerFunc func(ctx *CustomRouteContext)

func getTestRoute(ctx *CustomRouteContext) {
	ctx.String(http.StatusOK, "Nice!")
}

func NewGET(path string, handler CustomHandlerFunc) Route {
	return &DefaultRoute{
		Method: "GET",
		Path:   path,
		HandlerFunc: func(ctx interface{}) {
			handler(ctx.(*CustomRouteContext))
		},
	}
}

func main() {
	// Instead of NewRouter, you can create your own route matching algorithm by
	// r := NewRouter()
	// r.RouteMatchers = []RouteMatchers{&MyCustomMatcher{}, &RouteMatcherString{}}
	// A matcher just needs to follow the RouteMatcher interface.

	s := &Server{
		S: &Services{},
		R: NewRouter(),
	}
	s.R.AddRoute(NewGET("/test-route", getTestRoute))
	http.ListenAndServe(":8080", s)
}
