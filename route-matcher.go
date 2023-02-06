package gorouter

import (
	"errors"
)

var (
	ErrNoMatchersSatisfied = errors.New("no matchers satisfied")
)

type RouteMatcher[R any] interface {
	// Match tries to match the URL path.
	// Will return any remainder as well as a bool indicating whether a match was found or not.
	// Route Context is for the matcher to add a value to the context. It may be null (e.g., if it is used in add route).
	Match(method, path string, params *RouteParamList) (string, bool)

	// TokenMatch tries to match the template strings for a token.
	TokenMatch(token string, route Route[R]) (string, bool)

	// GetToken attempts to retrieve the next token. Returns if it was successful or not.
	// Router assumes the function will prefill the variables with the correct data
	// and then return true.
	GetToken(path string, route Route[R]) (string, bool)
}

// MatchPathToMatcher tries to match a path to an effective RouteMatcher
// through its GetToken method. As it goes through a list of matchers
// sequentially and will accept the first acceptable matcher, the lower
// priority matchers should be at the bottom.
func MatchPathToMatcher[R any](path string, route Route[R], tests []RouteMatcher[R]) (RouteMatcher[R], string, error) {
	for _, t := range tests {
		if rem, ok := t.GetToken(path, route); ok {
			return t, rem, nil
		}
	}
	return nil, "", ErrNoMatchersSatisfied
}

// Matches the root.
type RouteMatcherRoot[R any] struct{}

func (r *RouteMatcherRoot[R]) Match(method, path string, params *RouteParamList) (string, bool) {
	if len(path) == 0 {
		return "", true
	}
	return path[1:], path[0] == '/'
}
func (r *RouteMatcherRoot[R]) TokenMatch(path string, route Route[R]) (string, bool) {
	return path[1:], path[0] == '/'
}
func (r *RouteMatcherRoot[R]) GetToken(path string, route Route[R]) (string, bool) {
	return r.Match("", path, nil)
}

// Matches a static string
type RouteMatcherString[R any] struct {
	Path string
}

func (r *RouteMatcherString[R]) Match(method, path string, params *RouteParamList) (string, bool) {
	return r.TokenMatch(path, nil)
}

func (r *RouteMatcherString[R]) TokenMatch(path string, route Route[R]) (string, bool) {
	rlen := len(r.Path)
	if rlen > len(path) {
		return "", false
	}
	matches := r.Path == path[:rlen]
	if !matches {
		return "", false
	}
	if len(path) == rlen {
		return "", true
	}
	if path[rlen-1] == '/' {
		return path[rlen:], true
	}
	switch path[rlen] {
	case '/':
		// Return everything without the /
		return path[rlen+1:], true
	case '?', '#':
		// These are normally for frontend. We are ignoring them
		return "", true
	default:
		// Anything else, it should still be part of the token.
		// Therefore, we return false
		return "", false
	}
}

func (r *RouteMatcherString[R]) GetToken(path string, route Route[R]) (string, bool) {
	for i, p := range path {
		switch p {
		case '/':
			r.Path = path[:i+1]
			return path[i+1:], true
		case '?', '#':
			r.Path = path[i:]
			return "", true
		}
	}
	r.Path = path
	return "", true
}

type RouteMatcherPlaceholder[R any] struct{}

func (r *RouteMatcherPlaceholder[R]) Match(method, path string, params *RouteParamList) (string, bool) {
	// Use the string matcher's code to get the next path.
	// we want the *path* not the remainder.
	s := &RouteMatcherString[R]{}
	rem, _ := s.GetToken(path, nil)

	// Add the actual parameter value to the params!
	if s.Path[len(s.Path)-1] == '/' {
		s.Path = s.Path[:len(s.Path)-1]
	}
	params.Add(s.Path)
	return rem, true
}

func (r *RouteMatcherPlaceholder[R]) getTokenIndex(path string) int {
	if len(path) == 0 || path[0] != ':' {
		return -1
	}
	for i, c := range path {
		switch c {
		case '/', '?', '#':
			return i
		}
	}
	return len(path)
}

// TokenMatch is called when we are adding a route. It checks if the token
// matches. Therefore, we need to add params into the route as well. In this sense,
// it does the same thing as GetToken.
func (r *RouteMatcherPlaceholder[R]) TokenMatch(path string, route Route[R]) (string, bool) {
	return r.GetToken(path, route)
}

func (r *RouteMatcherPlaceholder[R]) GetToken(path string, route Route[R]) (string, bool) {
	idx := r.getTokenIndex(path)
	if idx == -1 {
		return "", false
	}

	if len(path) < idx+1 {
		route.AddParamName(path[1:])
		return "", true
	}

	route.AddParamName(path[1:idx])
	if path[idx] == '/' {
		return path[idx+1:], true
	}
	return "", true
}
