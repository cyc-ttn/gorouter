package gorouter

type Route[R any] interface {
	GetMethod() string
	GetPath() string
	GetHandler() func(R)
	GetParamNames() []string
	AddParamName(string)
}

type DefaultRoute[R any] struct {
	Method      string
	Path        string
	HandlerFunc func(R)
	ParamNames  []string
}

func (r *DefaultRoute[R]) GetMethod() string {
	return r.Method
}

func (r *DefaultRoute[R]) GetPath() string {
	return r.Path
}

func (r *DefaultRoute[R]) GetHandler() func(R) {
	return r.HandlerFunc
}

func (r *DefaultRoute[R]) GetParamNames() []string {
	return r.ParamNames
}

func (r *DefaultRoute[R]) AddParamName(name string) {
	r.ParamNames = append(r.ParamNames, name)
}
