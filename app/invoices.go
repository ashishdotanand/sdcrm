package sdatcrm

/*
C Post("/api/invoices/") -> Create a new invoice
R Get("/api/invoices/id/") -> Get the invoice with this id
U Put("/api/invoices/id/") -> Resave this invoice with id
D Delete("/api/invoices/id/") -> Delete invoice having id
Q Get("/api/invoices/") -> Get all invoices
*/

import (
	"appengine/delay"
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

const INVOICES_API = "/api/invoices/"

type InvoiceId int64

func init() {
	http.Handle(INVOICES_API, gaeHandler(invoiceHandler))
	http.HandleFunc("/invoice/new/", newInvoicePageHandler)
	http.HandleFunc("/invoice/", editInvoicePageHandler)
	http.HandleFunc("/invoices/", allInvoicePageHandler)
}

func invoiceHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := r.URL.Path[len(INVOICES_API):]
	if len(id) > 0 {
		switch r.Method {
		case "GET":
			id64, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return nil, err
			}
			invoice := new(Invoice)
			invoice.Id = InvoiceId(id64)
			return invoice.get(c)
		default:
			return nil, fmt.Errorf(r.Method + " on " + r.URL.Path + " not implemented")
		}
	} else {
		switch r.Method {
		case "POST":
			return invoiceSaveEntryPoint(c, r)
		case "GET":
			return getAllInvoices(c)
		default:
			return nil, fmt.Errorf(r.Method + " on " + r.URL.Path + " not implemented")
		}
	}
	return nil, nil
}

func decodeInvoice(r io.ReadCloser) (*Invoice, error) {
	defer r.Close()
	var invoice Invoice
	err := json.NewDecoder(r).Decode(&invoice)
	return &invoice, err
}

func (o *Invoice) get(c appengine.Context) (*Invoice, error) {
	err := datastore.Get(c, o.key(c), o)
	if err != nil {
		return nil, err
	}
	return o, nil
}
func (o *Invoice) save(c appengine.Context) (*Invoice, error) {
	k, err := datastore.Put(c, o.key(c), o)
	if err != nil {
		return nil, err
	}
	o.Id = InvoiceId(k.IntID())
	return o, nil
}

func defaultInvoiceList(c appengine.Context) *datastore.Key {
	ancestorKey := datastore.NewKey(c, "ANCESTOR_KEY", BranchName(c), 0, nil)
	return datastore.NewKey(c, "InvoiceList", "default", 0, ancestorKey)
}

func (o *Invoice) key(c appengine.Context) *datastore.Key {
	if o.Id == 0 {
		o.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Invoice", defaultInvoiceList(c))
	}
	return datastore.NewKey(c, "Invoice", "", int64(o.Id), defaultInvoiceList(c))
}

func getAllInvoices(c appengine.Context) ([]Invoice, error) {
	invoices := []Invoice{}
	ks, err := datastore.NewQuery("Invoice").Ancestor(defaultInvoiceList(c)).Order("Created").GetAll(c, &invoices)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(invoices); i++ {
		invoices[i].Id = InvoiceId(ks[i].IntID())
	}
	return invoices, nil
}

func newInvoicePageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/invoice.html"))
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

func editInvoicePageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/invoice.html"))
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

func allInvoicePageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/invoices.html"))
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

var laterSaveInvoiceRippleEffects = delay.Func("later-save-invoice-ripple-effects", func(c appengine.Context, invoice *Invoice) error {
	// 1. Find pending orders
	// 2. Sort them by date
	// 3. Start filling the ShippedItemWithInvId in Order
	// 4. Mark if they are complete
	// 5. Repeat until all the invoice items have been exhausted.
	// 6. Report error if you have items left in invoice but no corresponding order for it.

	pendingOrders, err := getPendingOrdersForPurchaser(c, invoice.PurchaserId)
	if err != nil {
		return err
	}
	Orders(pendingOrders).Sort()
	unmappedInvoiceItems := append(Items(nil), invoice.Items...)

	for oi := range pendingOrders {
		pendingOrder := &pendingOrders[oi]
		for ii := range pendingOrder.PendingItems {
			pendingItem := &pendingOrder.PendingItems[ii]
			for umi := range unmappedInvoiceItems {
				unmappedInvoiceItem := &unmappedInvoiceItems[umi]

				if pendingItem.Qty == 0 {
					continue
				}
				if pendingItem.Qty == 0 {
					continue
				}
				if (*pendingItem).skuEquals(*unmappedInvoiceItem) {
					//c.Infof("__________________________________")
					//c.Infof("unmappedInvoiceItem %+v", unmappedInvoiceItem)
					//c.Infof("pendingItem%+v", pendingItem)
					//Find minimum of the two quantities
					minQty := min(pendingItem.Qty, unmappedInvoiceItem.Qty)

					//Prepare siwi
					var siwi ShippedItemWithInvId
					siwi.Item = *unmappedInvoiceItem
					siwi.Item.Qty = minQty
					siwi.InvoiceId = invoice.Id
					//c.Infof("Created siwi %+v", siwi)

					//Reduce pending
					pendingItem.Qty -= minQty
					pendingOrder.ShippedItemsWithInvId = append(pendingOrder.ShippedItemsWithInvId, siwi)
					//c.Infof("pendingOrder.ShippedItemsWithInvId became %+v", pendingOrder.ShippedItemsWithInvId)

					//Reduce unmappedInvoiceItem.Qty
					unmappedInvoiceItem.Qty -= minQty

					//Store Order id in Invoice for future reference.
					invoice.OrdersId = append(invoice.OrdersId, pendingOrder.Id)

				}
			}
		}
	}

	for i := range pendingOrders {
		pendingOrder := &pendingOrders[i]
		c.Infof("___________________________________________")
		c.Infof("\nBefore Filtering pendingOrder.PendingItems = %#v", pendingOrder.PendingItems)
		EmptyItems := func(item Item) bool { return item.Qty == 0 }
		c.Infof("\nafter Filtering pendingOrder.PendingItems = %#v", pendingOrder.PendingItems)
		pendingOrder.PendingItems = pendingOrder.PendingItems.Filter(EmptyItems)
		if len(pendingOrder.PendingItems) == 0 {
			pendingOrder.IsComplete = true
		}
		if _, err := pendingOrder.save(c); err != nil {
			//TODO: If there are too many pending orders for a purchaser, you might want to set a dirty flag or multi save the entities. But this is unlikely.
			return err
		}
	}

	if _, err := invoice.save(c); err != nil {
		return err
	}

	return nil
})

func invoiceSaveEntryPoint(c appengine.Context, r *http.Request) (*Invoice, error) {
	c.Infof("_______________________")
	c.Infof("Raw Body of request having invoice: %#v", r.Body)
	invoice, err := decodeInvoice(r.Body)
	c.Infof("Decoded invoice: %#v", invoice)
	if err != nil {
		return nil, err
	}
	if invoice.Id == 0 {
		return ProcessBrandNewInvoice(c, invoice)
	} else {
		return ProcessBrandNewInvoice(c, invoice)
		//return ProcessExistingInvoice(c, invoice)
	}
}

func ProcessBrandNewInvoice(c appengine.Context, invoice *Invoice) (*Invoice, error) {
	var err error
	extraItems, err := FindExtraItemsInInvoice(c, invoice)
	if err != nil {
		return nil, err
	}
	if len(extraItems) > 0 {
		return nil, fmt.Errorf("The invoice has extraItems and cannot be punched as is. Please create the PO for these extra items first. The items are: %#v", extraItems)
	}

	invoice, err = invoice.save(c)
	if err != nil {
		return nil, err
	}

	laterSaveInvoiceRippleEffects.Call(c, invoice)

	return invoice, nil
}

func ProcessExistingInvoice(c appengine.Context, invoice *Invoice) (*Invoice, error) {
	//TODO:
	//Delete if not requred
	return invoice.save(c)
}

//func CreateAdHocOrderWithTheseItems(c appengine.Context, items []Item, invoice *Invoice) (*Order, error) {
//	order := new(Order)
//	order.PurchaserId = invoice.PurchaserId
//	order.SupplierName = invoice.SupplierName
//	order.Date = time.Now()
//	order.Number = "Telephonic"
//	for _, i := range items {
//		order.OrderedItems = append(order.OrderedItems, i)
//		order.PendingItems = append(order.PendingItems, i)
//	}
//	c.Infof("About to save order:%#v", order)
//	return order.save(c)
//}
//
//func CreateAdHocOrderIfRequired(c appengine.Context, invoice *Invoice) error {
//	// 1. Check if the invoice is being created for some extra items.
//	// 2. Create an adHoc order for extra items.
//	// 3. Recheck if invoice is being created for extra items. This time it should not be. Defensive Programming.
//	extraItems, err := FindExtraItemsInInvoice(c, invoice)
//	if err != nil {
//		return err
//	}
//	c.Infof("Extra items in invoice of %v: %#v", invoice.PurchaserId, extraItems)
//
//	if len(extraItems) > 0 {
//		o, err := CreateAdHocOrderWithTheseItems(c, extraItems, invoice)
//		if err != nil {
//			return err
//		}
//		c.Infof("created teh extra order: %#v", o)
//
//	}
//	return nil
//}
