package gorouter

type Route interface {
	GetMethod() string
	GetPath() string
	GetHandler() HandlerFunc
	GetParamNames() []string
	AddParamName(string)
}

type HandlerFunc func(ctx interface{})

type DefaultRoute struct {
	Method      string
	Path        string
	HandlerFunc HandlerFunc
	ParamNames  []string
}

func (r *DefaultRoute) GetMethod() string {
	return r.Method
}

func (r *DefaultRoute) GetPath() string {
	return r.Path
}

func (r *DefaultRoute) GetHandler() HandlerFunc {
	return r.HandlerFunc
}

func (r *DefaultRoute) GetParamNames() []string {
	return r.ParamNames
}

func (r *DefaultRoute) AddParamName(name string) {
	r.ParamNames = append(r.ParamNames, name)
}
