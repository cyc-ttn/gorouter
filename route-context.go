package gorouter

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
)

type RouteContext struct {
	W http.ResponseWriter
	R *http.Request

	Method string
	Path   string
	UserId string // LoggedOn User
	Params map[string]string
	Query  url.Values
}

func (c *RouteContext) Status(status int) {
	c.W.WriteHeader(status)
}

func (c *RouteContext) JSON(status int, v interface{}) error {
	c.W.Header().Add("Content-Type", "application/json")
	c.Status(status)
	return json.NewEncoder(c.W).Encode(v)
}

func (c *RouteContext) Redirect(status int, url string) {
	http.Redirect(c.W, c.R, url, status)
}

func (c *RouteContext) String(status int, format string, data ...interface{}) {
	c.W.Header().Add("Content-Type", "text/plain; charset=utf-8")
	c.Status(status)
	if _, err := io.WriteString(c.W, fmt.Sprintf(format, data...)); err != nil {
		log.Println(err)
	}
}

func (c *RouteContext) ShouldBindJSON(v interface{}) error {
	return json.NewDecoder(c.R.Body).Decode(v)
}

func (c *RouteContext) FormFile(name string, maxMemory int64) (*multipart.FileHeader, error) {
	if c.R.MultipartForm == nil {
		if err := c.R.ParseMultipartForm(maxMemory); err != nil {
			return nil, err
		}
	}
	file, fileHeader, err := c.R.FormFile(name)
	if err != nil {
		return nil, err
	}
	file.Close()
	return fileHeader, nil
}
