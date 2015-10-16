package sdatcrm

/*
C Post("/api/orders/") -> Create a new order
R Get("/api/orders/id/") -> Get the order with this id
U Put("/api/orders/id/") -> Resave this order with id
D Delete("/api/orders/id/") -> Delete order having id
Q Get("/api/orders/") -> Get all orders
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

const ORDERS_API = "/api/orders/"

type OrderId int64

func init() {
	http.Handle(ORDERS_API, gaeHandler(orderHandler))
	http.HandleFunc("/order/new/", newOrderPageHandler)
	http.HandleFunc("/order/", editOrderPageHandler)
	http.HandleFunc("/orders/", allOrdersPageHandler)
	return
}

func orderHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	oid := r.URL.Path[len(ORDERS_API):]
	c.Infof("Received oid %v", oid)
	if len(oid) > 0 {
		switch r.Method {
		case "GET":
			oid64, err := strconv.ParseInt(oid, 10, 64)
			if err != nil {
				return nil, err
			}
			order := new(Order)
			order.Id = OrderId(oid64)
			return order.get(c)
		default:
			return nil, fmt.Errorf(r.Method + " on " + r.URL.Path + " not implemented")
		}
	} else {
		switch r.Method {
		case "POST":
			return orderSaveEntryPoint(c, r)
		case "GET":
			return getAllOrders(c)
		default:
			return nil, fmt.Errorf(r.Method + " on " + r.URL.Path + " not implemented")
		}
	}
	return nil, nil
}

func decodeOrder(r io.ReadCloser) (*Order, error) {
	defer r.Close()
	var order Order
	err := json.NewDecoder(r).Decode(&order)
	return &order, err
}

func (o *Order) get(c appengine.Context) (*Order, error) {
	err := datastore.Get(c, o.key(c), o)
	if err != nil {
		return nil, err
	}
	return o, nil
}
func (o *Order) save(c appengine.Context) (*Order, error) {
	k, err := datastore.Put(c, o.key(c), o)
	if err != nil {
		return nil, err
	}
	o.Id = OrderId(k.IntID())
	return o, nil
}

func defaultOrderList(c appengine.Context) *datastore.Key {
	ancestorKey := datastore.NewKey(c, "ANCESTOR_KEY", BranchName(c), 0, nil)
	return datastore.NewKey(c, "OrderList", "default", 0, ancestorKey)
}

func (o *Order) key(c appengine.Context) *datastore.Key {
	if o.Id == 0 {
		o.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Order", defaultOrderList(c))
	}
	return datastore.NewKey(c, "Order", "", int64(o.Id), defaultOrderList(c))
}

func getAllOrders(c appengine.Context) ([]Order, error) {
	orders := []Order{}
	ks, err := datastore.NewQuery("Order").Ancestor(defaultOrderList(c)).Order("Created").GetAll(c, &orders)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(orders); i++ {
		orders[i].Id = OrderId(ks[i].IntID())
	}
	return orders, nil
}

func newOrderPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/order.html"))
	var data interface{}
	data = struct{ Nature string }{"NEW"}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func editOrderPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/order.html"))
	var data interface{}
	data = struct{ Nature string }{"EDIT"}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func allOrdersPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/orders.html"))
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func orderSaveEntryPoint(c appengine.Context, r *http.Request) (*Order, error) {
	order, err := decodeOrder(r.Body)
	if err != nil {
		return nil, err
	}
	if order.Id == 0 {
		for _, oi := range order.OrderedItems {
			order.PendingItems = append(order.PendingItems, oi)
		}
	}
	return order.save(c)
}
