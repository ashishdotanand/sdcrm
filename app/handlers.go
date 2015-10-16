package sdatcrm

import (
	"appengine"
	"encoding/json"
	"html/template"
	"net/http"
)

type gaeHandler func(c appengine.Context, w http.ResponseWriter, r *http.Request) (interface{}, error)

func (h gaeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	val, err := h(c, w, r)
	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}
	if err != nil {
		c.Errorf("%v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type urlStruct struct {
	handler      func(w http.ResponseWriter, r *http.Request)
	templatePath string
	data         interface{}
}

type apiStruct struct {
	handler func(w http.ResponseWriter, r *http.Request)
}

var urlMaps map[string]urlStruct
var apiMaps map[string]apiStruct

type TEMPLATE_AND_DATA struct {
	t    *template.Template
	data interface{}
}

var PRE_BUILT_TEMPLATES_WITH_DATA = make(map[string]TEMPLATE_AND_DATA)
var PAGE_NOT_FOUND_TEMPLATE = template.Must(template.ParseFiles("templates/pageNotFound.html"))

func initDynamicHTMLUrlMaps() {
	http.HandleFunc("/dynamic", apiNotImplementedHandler)
}

func initStaticHTMLUrlMaps() {
	urlMaps := map[string]urlStruct{"/": {generalPageHandler, "templates/control-panel.html", nil}}

	for path, urlBlob := range urlMaps {
		templatePath := urlBlob.templatePath
		data := urlBlob.data
		PRE_BUILT_TEMPLATES_WITH_DATA[path] = TEMPLATE_AND_DATA{template.Must(template.ParseFiles(templatePath)), data}
	}

	for path, urlBlob := range urlMaps {
		http.HandleFunc(path, urlBlob.handler)
	}
	return
}

func initRootApiMaps() {
	apiMaps := map[string]apiStruct{
		"/a/api/echo": {echoAPIHandler},
		"/a/api/":     {apiNotImplementedHandler},
		"/api/":       {apiNotImplementedHandler},
	}

	for path, apiBlob := range apiMaps {
		http.HandleFunc(path, apiBlob.handler)
	}

	return
}

func init() {
	initRootApiMaps()
	initStaticHTMLUrlMaps()
	initDynamicHTMLUrlMaps()
	return
}

func generalPageHandler(w http.ResponseWriter, r *http.Request) {
	t := PRE_BUILT_TEMPLATES_WITH_DATA[r.URL.Path].t
	data := PRE_BUILT_TEMPLATES_WITH_DATA[r.URL.Path].data
	if t == nil {
		t = PAGE_NOT_FOUND_TEMPLATE
		data = nil
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func apiNotImplementedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, r.URL.Path+" not implemented", http.StatusNotImplemented)
	return
}

func echoAPIHandler(w http.ResponseWriter, r *http.Request) {
	WriteJson(w, "test")
	return
}
