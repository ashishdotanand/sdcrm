package sdatcrm

/*
C Post("/api/purchasers/") -> Create a new purchaser
R Get("/api/purchasers/id/") -> Get the purchaser with this id
U Put("/api/purchasers/id/") -> Resave this purchaser with id
D Delete("/api/purchasers/id/") -> Delete purchaser having id
Q Get("/api/purchasers/") -> Get all purchasers
*/

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"appengine"
	"appengine/datastore"
)

const PURCHASERS_API = "/api/purchasers/"

type CSL string //Comma or Semicolon separated list
type PurchaserId int64

func init() {
	http.Handle(PURCHASERS_API, gaeHandler(purchaserHandler))
	http.HandleFunc("/purchaser/new/", newPurchaserPageHandler)
	http.HandleFunc("/purchaser/", editPurchaserPageHandler)
	http.HandleFunc("/purchasers/", allPurchaserPageHandler)
}

func purchaserHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := r.URL.Path[len(PURCHASERS_API):]
	if len(id) > 0 {
		switch r.Method {
		case "GET":
			id64, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return nil, err
			}
			purchaser := new(Purchaser)
			purchaser.Id = PurchaserId(id64)
			return purchaser.get(c)
		default:
			return nil, fmt.Errorf(r.Method + " on " + r.URL.Path + " not implemented")
		}
	} else {
		switch r.Method {
		case "POST":
			purchaser, err := decodePurchaser(r.Body)
			if err != nil {
				return nil, err
			}
			return purchaser.save(c)
		case "GET":
			return getAllPurchasers(c)
		default:
			return nil, fmt.Errorf(r.Method + " on " + r.URL.Path + " not implemented")
		}
	}
	return nil, nil
}

func decodePurchaser(r io.ReadCloser) (*Purchaser, error) {
	defer r.Close()
	var purchaser Purchaser
	err := json.NewDecoder(r).Decode(&purchaser)
	return &purchaser, err
}

func (o *Purchaser) get(c appengine.Context) (*Purchaser, error) {
	err := datastore.Get(c, o.key(c), o)
	if err != nil {
		return nil, err
	}
	return o, nil
}
func (o *Purchaser) save(c appengine.Context) (*Purchaser, error) {
	k, err := datastore.Put(c, o.key(c), o)
	if err != nil {
		return nil, err
	}
	o.Id = PurchaserId(k.IntID())
	return o, nil
}

func defaultPurchaserList(c appengine.Context) *datastore.Key {
	ancestorKey := datastore.NewKey(c, "ANCESTOR_KEY", BranchName(c), 0, nil)
	return datastore.NewKey(c, "PurchaserList", "default", 0, ancestorKey)
}

func (o *Purchaser) key(c appengine.Context) *datastore.Key {
	if o.Id == 0 {
		o.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Purchaser", defaultPurchaserList(c))
	}
	return datastore.NewKey(c, "Purchaser", "", int64(o.Id), defaultPurchaserList(c))
}

func getAllPurchasers(c appengine.Context) ([]Purchaser, error) {
	//TODO: If performance is slow we should consider projection queries here. But not solving the problem now as pre-optimization is the root cause of all evil.
	purchasers := []Purchaser{}
	ks, err := datastore.NewQuery("Purchaser").Ancestor(defaultPurchaserList(c)).Order("Created").GetAll(c, &purchasers)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(purchasers); i++ {
		purchasers[i].Id = PurchaserId(ks[i].IntID())
	}
	return purchasers, nil
}

func newPurchaserPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/purchaser.html"))
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

func editPurchaserPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/purchaser.html"))
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

func allPurchaserPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/purchasers.html"))
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
