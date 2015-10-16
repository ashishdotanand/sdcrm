package sdatcrm

/*
R Get("/api/extraItemsInInvoice/") -> Get all extra items in this invoice for which an adHoc order should be created
*/

import (
	"fmt"
	"net/http"

	"appengine"
)

const EXTRA_ITEMS_IN_INVOICE_API = "/api/extraItemsInInvoice/"

func init() {
	http.Handle(EXTRA_ITEMS_IN_INVOICE_API, gaeHandler(extraItemsInInvoiceHandler))
	return
}

func extraItemsInInvoiceHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	pid := r.URL.Path[len(EXTRA_ITEMS_IN_INVOICE_API):]
	if len(pid) == 0 {
		switch r.Method {
		case "POST":
			invoice, err := decodeInvoice(r.Body)
			if err != nil {
				return nil, err
			}
			extraItems, err := FindExtraItemsInInvoice(c, invoice)
			if err != nil {
				return nil, err
			}
			type ItemGroup struct{ Items Items }
			var itemGroup ItemGroup
			for i := range extraItems {
				itemGroup.Items = append(itemGroup.Items, extraItems[i])

			}
			return itemGroup, nil
		default:
			return nil, fmt.Errorf(r.Method + " on " + r.URL.Path + " not implemented")
		}
	}
	return nil, nil
}

func PrintItem(c appengine.Context, x Item) {
	//c.Infof("\nItem: %#v", x)
}
func PrintSPItem(c appengine.Context, x []*Item) {
	for _, a := range x {
		PrintItem(c, *a)
	}
}
func PrintSItem(c appengine.Context, x []Item) {
	for _, a := range x {
		PrintItem(c, a)
	}
}
func AddToSquashed(c appengine.Context, cpi *Item, spi Items) Items {
	var cpy = *cpi
	//c.Infof("Got the item:")
	PrintItem(c, cpy)

	found := false
	for i, _ := range spi {
		if (spi[i]).skuEquals(cpy) {
			found = true
			//c.Infof("Increasing the quantity from %#v to %#v for following item:", spi[i].Qty, spi[i].Qty+cpi.Qty)
			PrintItem(c, cpy)
			spi[i].Qty += cpi.Qty
		} else {

			//c.Infof("These skus are not equal:")
			PrintItem(c, cpy)
			PrintItem(c, spi[i])
		}

	}
	if found == false {
		//c.Infof("Appending to spi")
		PrintItem(c, cpy)
		spi = append(spi, cpy)
	}
	return spi
}

func FindExtraItemsInInvoice(c appengine.Context, invoice *Invoice) (Items, error) {
	pendingOrders, err := getPendingOrdersForPurchaser(c, invoice.PurchaserId)
	if err != nil {
		return nil, err
	}

	var clubbedPendingItems Items
	for _, o := range pendingOrders {
		for _, i := range o.PendingItems {
			clubbedPendingItems = append(clubbedPendingItems, i)
		}
	}
	//c.Infof("\n\n________\nClubbedItems")
	PrintSItem(c, clubbedPendingItems)

	var squashedPendingItems Items
	for _, cpi := range clubbedPendingItems {
		//c.Infof("\n\n________\n Calling AddToSquashed")
		squashedPendingItems = AddToSquashed(c, &cpi, squashedPendingItems)
	}

	//c.Infof("\n\n________\n squashedPendingItems")
	PrintSItem(c, squashedPendingItems)

	var extraItemsInInvoice = append(Items(nil), invoice.Items...)
	for i := range extraItemsInInvoice {
		for _, cpi := range clubbedPendingItems {
			if extraItemsInInvoice[i].skuEquals(cpi) {
				extraItemsInInvoice[i].Qty -= cpi.Qty
			}
		}
	}
	//c.Infof("\n\n________\nExtraItemsInInvoice")
	PrintSItem(c, extraItemsInInvoice)

	var prunedExtraItems Items
	for _, i := range extraItemsInInvoice {
		if i.Qty != 0 {
			prunedExtraItems = append(prunedExtraItems, i)
		}
	}
	//c.Infof("\n\n________\n prunedExtraItems")
	PrintSItem(c, prunedExtraItems)

	return prunedExtraItems, nil
}
