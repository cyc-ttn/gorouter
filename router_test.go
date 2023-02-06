package gorouter

import (
	"testing"
)

func TestRouterNode_AddRoute(t *testing.T) {

	routes := []*DefaultRoute[any]{
		{
			Method: "GET",
			Path:   "/users",
		},
		{
			Method: "GET",
			Path:   "/user",
		},
		{
			Method: "GET",
			Path:   "/user/:id",
		},
		{
			Method: "POST",
			Path:   "/user/:id/delete",
		},
		{
			Method: "GET",
			Path:   "/users/:location/:id",
		},
	}

	// This should make a tree with two children (users and user) and under user, there should be one more
	// leaf. A total of three nodes.
	router := NewRouter[any]()
	for _, r := range routes {
		if err := router.AddRoute(r); err != nil {
			t.Fatal(err)
		}
	}
	if len(router.Children) != 2 {
		t.Error("Expecting two children")
	}

	// Test the router for different things using the route context.
	paths := []struct {
		Method       string
		Path         string
		RouteContext *RouteContext

		ExpectError bool
		CtxTests    map[string]string
	}{
		{
			Method: "GET",
			Path:   "/users",
		},
		{
			Method: "GET",
			Path:   "/user",
		},
		{
			Method: "GET",
			Path:   "/user/12345",
			CtxTests: map[string]string{
				"id": "12345",
			},
		},
		{
			Method:      "POST",
			Path:        "/user/12345",
			ExpectError: true,
		},
		{
			Method: "POST",
			Path:   "/user/12345/delete",
			CtxTests: map[string]string{
				"id": "12345",
			},
		},
		{
			Method: "GET",
			Path:   "/users/railtown/12345",
			CtxTests: map[string]string{
				"location": "railtown",
				"id":       "12345",
			},
		},
		{
			Method: "GET",
			Path:   "/users/railtown/12345?aoijwegjaoiwjf#09jflikja",
			CtxTests: map[string]string{
				"location": "railtown",
				"id":       "12345",
			},
		},
	}

	for _, p := range paths {
		p.RouteContext = &RouteContext{}
		route, err := router.Match(p.Method, p.Path, p.RouteContext)
		if err != nil {
			if p.ExpectError {
				continue
			}
			t.Errorf("There is an error with matching route %s %s: %s", p.Method, p.Path, err)
			continue
		}

		if route == nil {
			t.Errorf("Route for %s %s returned nil", p.Method, p.Path)
		}

		if len(p.CtxTests) == 0 {
			continue
		}

		for key, value := range p.CtxTests {
			res, ok := p.RouteContext.Params[key]
			if !ok {
				t.Errorf("Expect param %s to be present, but it isn't", key)
			}
			if value != res {
				t.Errorf("Expect param %s to be %s but got %s", key, value, res)
			}
		}
	}
}
