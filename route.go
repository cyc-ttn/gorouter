package gorouter

type Route[R any] interface {
	GetMethod() string
	GetPath() string
	GetHandler() HandlerFunc[R]
	GetParamNames() []string
	AddParamName(string)
}

type HandlerFunc[R any] func(ctx R)

type DefaultRoute[R any] struct {
	Method      string
	Path        string
	HandlerFunc HandlerFunc[R]
	ParamNames  []string
}

func (r *DefaultRoute[R]) GetMethod() string {
	return r.Method
}

func (r *DefaultRoute[R]) GetPath() string {
	return r.Path
}

func (r *DefaultRoute[R]) GetHandler() HandlerFunc[R] {
	return r.HandlerFunc
}

func (r *DefaultRoute[R]) GetParamNames() []string {
	return r.ParamNames
}

func (r *DefaultRoute[R]) AddParamName(name string) {
	r.ParamNames = append(r.ParamNames, name)
}
