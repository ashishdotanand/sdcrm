package sdatcrm

/*
C Post("/api/foos/") -> Create a new foo
R Get("/api/foos/id/") -> Get the foo with this id
U Put("/api/foos/id/") -> Resave this foo with id
D Delete("/api/foos/id/") -> Delete foo having id
Q Get("/api/foos/") -> Get all foos
*/

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"
)

const FOOS_API = "/api/foos/"

type FooId int64

type Foo struct {
	Id            FooId     `json:"Id" datastore:"-"`
	Items         []Item    `json:"Items"`
	Created       time.Time `json:"Created"`
	Date          time.Time `json:"Date"`
	TotalQty      int64     `json:"TotalQty"`
	PurchaserName string    `json:"PurchaserName"`
	SupplierName  string    `json:"SupplierName"`
	Number        string    `json:"Number"`
	OrdersId      []OrderId `json:"OrdersId"`
}

func init() {
	http.Handle(FOOS_API, gaeHandler(fooHandler))
	http.HandleFunc("/foo/new/", newFooPageHandler)
	http.HandleFunc("/foo/", editFooPageHandler)
	http.HandleFunc("/foos/", allFooPageHandler)
}

func fooHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := r.URL.Path[len(FOOS_API):]
	c.Errorf("Received foo id %v", id)
	if len(id) > 0 {
		switch r.Method {
		case "GET":
			id64, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return nil, err
			}
			foo := new(Foo)
			foo.Id = FooId(id64)
			return foo.get(c)
		default:
			return nil, fmt.Errorf(r.Method + " on " + r.URL.Path + " not implemented")
		}
	} else {
		switch r.Method {
		case "POST":
			foo, err := decodeFoo(r.Body)
			if err != nil {
				return nil, err
			}
			return foo.save(c)
		case "GET":
			return getAllFoos(c)
		default:
			return nil, fmt.Errorf(r.Method + " on " + r.URL.Path + " not implemented")
		}
	}
	return nil, nil
}

func decodeFoo(r io.ReadCloser) (*Foo, error) {
	defer r.Close()
	var foo Foo
	err := json.NewDecoder(r).Decode(&foo)
	return &foo, err
}

func (o *Foo) get(c appengine.Context) (*Foo, error) {
	err := datastore.Get(c, o.key(c), o)
	if err != nil {
		return nil, err
	}
	return o, nil
}
func (o *Foo) save(c appengine.Context) (*Foo, error) {
	k, err := datastore.Put(c, o.key(c), o)
	if err != nil {
		return nil, err
	}
	o.Id = FooId(k.IntID())
	return o, nil
}

func defaultFooList(c appengine.Context) *datastore.Key {
	ancestorKey := datastore.NewKey(c, "ANCESTOR_KEY", BranchName(c), 0, nil)
	return datastore.NewKey(c, "FooList", "default", 0, ancestorKey)
}

func (o *Foo) key(c appengine.Context) *datastore.Key {
	if o.Id == 0 {
		o.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Foo", defaultFooList(c))
	}
	return datastore.NewKey(c, "Foo", "", int64(o.Id), defaultFooList(c))
}

func getAllFoos(c appengine.Context) ([]Foo, error) {
	foos := []Foo{}
	ks, err := datastore.NewQuery("Foo").Ancestor(defaultFooList(c)).Order("Created").GetAll(c, &foos)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(foos); i++ {
		foos[i].Id = FooId(ks[i].IntID())
	}
	return foos, nil
}

func newFooPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/foo.html"))
	var data interface{}
	data = struct{ Nature string }{"NEW"}
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

func editFooPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/foo.html"))
	var data interface{}
	data = struct{ Nature string }{"EDIT"}
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

func allFooPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/foos.html"))
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
