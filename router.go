package gorouter

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidMatcher = errors.New("invalid matcher")
	ErrPathNotFound   = errors.New("path not found")
)

type RouterNode struct {
	Children      []*RouterNode
	Matcher       RouteMatcher
	Route         map[string]*Route
	RouteMatchers []RouteMatcher
}

func NewRouter() *RouterNode {
	return &RouterNode{
		Children: make([]*RouterNode, 0, 1),
		Matcher:  &RouteMatcherRoot{},
		RouteMatchers: []RouteMatcher{
			&RouteMatcherPlaceholder{},
			&RouteMatcherString{},
		},
	}
}

func (r *RouterNode) AddRoute(route *Route) error {
	return r.add(route, route.Path)
}

// Adds a route to RouterNode
func (r *RouterNode) add(route *Route, path string) error {
	if r.Matcher == nil {
		return ErrInvalidMatcher
	}

	// Make sure that we can get the token
	// for the current matcher. "TokenMatch should
	// return false when the token doesn't fit.
	rem, ok := r.Matcher.TokenMatch(path, route)
	if !ok {
		return ErrInvalidMatcher
	}

	for _, c := range r.Children {
		// If a child node is able to add then the
		// route has been added. Otherwise, we need to add the route!
		if err := c.add(route, rem); err == nil {
			return nil
		}
	}

	// Create matchers until there are no matchers left to create!
	node := r
	for rem != "" {
		var matcher RouteMatcher
		var err error

		// Add the route to myself by splitting into tokens!
		matcher, rem, err = MatchPathToMatcher(rem, route, r.RouteMatchers)
		if err != nil {
			return err
		}

		newNode := &RouterNode{
			Children:      make([]*RouterNode, 0, 1),
			Matcher:       matcher,
			RouteMatchers: r.RouteMatchers,
		}
		node.Children = append(node.Children, newNode)
		node = newNode
	}
	node.addLeaf(route)

	return nil
}

func (r *RouterNode) addLeaf(route *Route) {
	if r.Route == nil {
		r.Route = make(map[string]*Route)
	}
	r.Route[route.Method] = route
}

func (r *RouterNode) getLeaf(method string) *Route {
	if r.Route == nil {
		return nil
	}
	return r.Route[method]
}

type RouteParamList []string

func (r *RouteParamList) Add(param string) {
	*r = append(*r, param)
}

func (r *RouterNode) Match(method, path string, ctx *RouteContext) (*Route, error) {

	// If there is a # in the path, completely ignore it.
	hashIdx := strings.Index(path, "#")
	if hashIdx > -1 {
		path = path[:hashIdx]
	}

	// If there is a ? in the path, parse separately.
	queryIdx := strings.Index(path, "?")
	if queryIdx > -1 {
		//Ignore erroneous query strings.
		ctx.Query, _ = url.ParseQuery(path[queryIdx+1:])
		path = path[:queryIdx]
	}

	params := &RouteParamList{}
	route, err := r.match(method, path, params)
	if err != nil {
		return nil, err
	}

	ctx.Params = make(map[string]string)
	for i, p := range *params {
		name := route.ParamNames[i]
		ctx.Params[name] = p
	}

	return route, nil
}

func (r *RouterNode) match(method, path string, params *RouteParamList) (*Route, error) {
	if r.Matcher == nil {
		return nil, ErrInvalidMatcher
	}
	rem, ok := r.Matcher.Match(method, path, params)
	if !ok {
		return nil, ErrPathNotFound
	}

	if rem == "" {
		route := r.getLeaf(method)
		if route == nil {
			return nil, ErrPathNotFound
		}
		return route, nil
	}

	for _, c := range r.Children {
		r, err := c.match(method, rem, params)
		if err == ErrInvalidMatcher {
			return nil, ErrInvalidMatcher
		}
		if err == nil {
			return r, nil
		}
	}
	return nil, ErrPathNotFound
}
