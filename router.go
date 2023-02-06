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

type RouterNode[R any] struct {
	Children      []*RouterNode[R]
	Matcher       RouteMatcher[R]
	Route         map[string]Route[R]
	RouteMatchers func() []RouteMatcher[R]
}

func NewRouter[R any]() *RouterNode[R] {
	return &RouterNode[R]{
		Children: make([]*RouterNode[R], 0, 1),
		Matcher:  &RouteMatcherRoot[R]{},
		RouteMatchers: func() []RouteMatcher[R] {
			return []RouteMatcher[R]{
				&RouteMatcherPlaceholder[R]{},
				&RouteMatcherString[R]{},
			}
		},
	}
}

func (r *RouterNode[R]) AddRoute(route Route[R]) error {
	return r.add(route, route.GetPath())
}

// Adds a route to RouterNode
func (r *RouterNode[R]) add(route Route[R], path string) error {
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
		var matcher RouteMatcher[R]
		var err error

		// Add the route to myself by splitting into tokens!
		matcher, rem, err = MatchPathToMatcher[R](rem, route, r.RouteMatchers())
		if err != nil {
			return err
		}

		newNode := &RouterNode[R]{
			Children:      make([]*RouterNode[R], 0, 1),
			Matcher:       matcher,
			RouteMatchers: r.RouteMatchers,
		}
		node.Children = append(node.Children, newNode)
		node = newNode
	}
	node.addLeaf(route)

	return nil
}

func (r *RouterNode[R]) addLeaf(route Route[R]) {
	if r.Route == nil {
		r.Route = make(map[string]Route[R])
	}
	r.Route[route.GetMethod()] = route
}

func (r *RouterNode[R]) getLeaf(method string) Route[R] {
	if r.Route == nil {
		return nil
	}
	return r.Route[method]
}

type RouteParamList []string

func (r *RouteParamList) Add(param string) {
	*r = append(*r, param)
}

func (r *RouterNode[R]) Match(method, path string, ctx *RouteContext) (Route[R], error) {

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
		name := route.GetParamNames()[i]
		ctx.Params[name] = p
	}

	return route, nil
}

func (r *RouterNode[R]) match(method, path string, params *RouteParamList) (Route[R], error) {
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
