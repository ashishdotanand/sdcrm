package sdatcrm

/*
R Get("/api/pendingorders/purchasers/id") -> Get all pending orders for this purchaserId
Q Get("/api/pendingorders/purchasers/") -> Get all pending orders
*/

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"appengine"
	"appengine/datastore"
)

const PENDING_ORDERS_API = "/api/pendingorders/purchasers/"

func init() {
	http.Handle(PENDING_ORDERS_API, gaeHandler(pendingOrdersHandler))
	http.HandleFunc("/pendingorders/", allPendingOrdersPageHandler)
	return
}

func pendingOrdersHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	pid := r.URL.Path[len(PENDING_ORDERS_API):]
	if len(pid) > 0 {
		switch r.Method {
		case "GET":
			pid64, err := strconv.ParseInt(pid, 10, 64)
			if err != nil {
				return nil, err
			}
			return getPendingOrdersForPurchaser(c, PurchaserId(pid64))
		default:
			return nil, fmt.Errorf(r.Method + " on " + r.URL.Path + " not implemented")
		}
	} else {
		switch r.Method {
		case "POST":
			order, err := decodeOrder(r.Body)
			if err != nil {
				return nil, err
			}
			return order.save(c)
		case "GET":
			return getAllPendingOrders(c)
		default:
			return nil, fmt.Errorf(r.Method + " on " + r.URL.Path + " not implemented")
		}
	}
	return nil, nil
}

func getAllPendingOrders(c appengine.Context) ([]Order, error) {
	pendingorders := []Order{}
	ks, err := datastore.NewQuery("Order").Ancestor(defaultOrderList(c)).Order("Created").Filter("IsComplete=", false).GetAll(c, &pendingorders)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(pendingorders); i++ {
		pendingorders[i].Id = OrderId(ks[i].IntID())
	}
	return pendingorders, nil
}

func getPendingOrdersForPurchaser(c appengine.Context, pid64 PurchaserId) ([]Order, error) {
	pendingorders := []Order{}
	ks, err := datastore.NewQuery("Order").Ancestor(defaultOrderList(c)).Order("Created").Filter("PurchaserId=", pid64).Filter("IsComplete=", false).GetAll(c, &pendingorders)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(pendingorders); i++ {
		pendingorders[i].Id = OrderId(ks[i].IntID())
	}
	return pendingorders, nil
}

func allPendingOrdersPageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/pendingorders.html"))
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
