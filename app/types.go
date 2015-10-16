package sdatcrm

import (
	"sort"
	"time"
)

type TCDieItem struct {
	PelletSize string `json:"PelletSize"`
	BoreSize   string `json:"BoreSize"`
	CaseType   string `json:"CaseType"`
	CaseSize   string `json:"CaseSize"`
}

type MiscItem struct {
	Name string `json:"Name"`
	Unit string `json:"Unit"`
}

type SKU struct {
	TCDieItem
	MiscItem
	Rate     float64 `json:"Rate"`
	Type     string  `json:"Type"` //TCD or MSC
	CRemarks string  `json:"CRemarks"`
}

func (a Item) skuEquals(b Item) bool {
	if a.PelletSize != b.PelletSize {
		return false
	}
	if a.BoreSize != b.BoreSize {
		return false
	}
	if a.CaseType != b.CaseType {
		return false
	}
	if a.CaseSize != b.CaseSize {
		return false
	}
	if a.Name != b.Name {
		return false
	}
	if a.Unit != b.Unit {
		return false
	}
	if a.Rate != b.Rate {
		return false
	}
	if a.Type != b.Type {
		return false
	}
	if a.CRemarks != b.CRemarks {
		return false
	}

	return true
}

type Item struct {
	SKU
	Qty int64 `json:"Qty"`
}

type Items []Item

type FilterItemsFunc func(Item) bool

func (p Items) Filter(f FilterItemsFunc) Items {
	//Returns a copy
	var newItems Items
	for _, x := range p {
		if !f(x) {
			newItems = append(newItems, x)
		}
	}
	return newItems
}

type Address struct {
	DeliveryAddress      string `json:"DeliveryAddress"`
	City                 string `json:"City"`
	State                string `json:"State"`
	Pincode              string `json:"Pincode"`
	EnvelopePhoneNumbers CSL    `json:"EnvelopePhoneNumbers"`
}

type Purchaser struct {
	Address
	Id                   PurchaserId `json:"Id" datastore:"-"`
	SKUs                 []SKU       `json:"SKUs"`
	Created              time.Time   `json:"Created"`
	Name                 string      `json:"Name"`
	DispatchEmails       CSL         `json:"DispatchEmails"`
	DefaultTaxPercentage float64     `json:"DefaultTaxPercentage"`
	FORMCEmails          CSL         `json:"FORMCEmails"`
	TinNumber            string      `json:"TinNumber"`
	BillingAddress       string      `json:"BillingAddress"`
	SMSPhoneNumbers      CSL         `json:"SMSPhoneNumbers"`
	MDPhoneNumbers       CSL         `json:"MDPhoneNumbers"`
	CreditDays           int64       `json:"CreditDays"`
	CRemarks             string      `json:"CRemarks"`
}

type Order struct {
	Id                    OrderId                `json:"Id" datastore:"-"`
	Created               time.Time              `json:"Created"`
	Date                  time.Time              `json:"Date"`
	TotalQty              int64                  `json:"TotalQty"`
	PurchaserId           PurchaserId            `json:"PurchaserId"`
	SupplierName          string                 `json:"SupplierName"`
	Number                string                 `json:"Number"`
	IsComplete            bool                   `json:"IsComplete"`
	OrderedItems          Items                  `json:"OrderedItems"`
	PendingItems          Items                  `json:"PendingItems"`
	ShippedItemsWithInvId []ShippedItemWithInvId `json:"ShippedItemsWithInvId"`
	PuntedItems           Items                  `json:"PuntedItems"`
}

type Orders []Order

func (p Orders) Len() int           { return len(p) }
func (p Orders) Less(i, j int) bool { return p[i].Date.Before(p[j].Date) }
func (p Orders) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Orders) Clone() Orders {
	var newOrders Orders
	for _, x := range p {
		newOrders = append(newOrders, x)
	}
	return newOrders
}

// Sort is a convenience method.
func (p Orders) Sort() { sort.Sort(p) }

type ShippedItemWithInvId struct {
	//This will be contained by the order. All the smart stuff is contained inside the order. Because and invoice is representing a piece of truth in real world. And it should be left as is.
	Item      Item
	InvoiceId InvoiceId `json:"InvoiceId"`
}

type Invoice struct {
	Id             InvoiceId   `json:"Id" datastore:"-"`
	Items          Items       `json:"Items"`
	Created        time.Time   `json:"Created"`
	Date           time.Time   `json:"Date"`
	TotalQty       int64       `json:"TotalQty"`
	PurchaserId    PurchaserId `json:"PurchaserId"`
	SupplierName   string      `json:"SupplierName"`
	Number         string      `json:"Number"`
	PRemarks       string      `json:"PRemarks"`
	OrdersId       []OrderId   `json:"OrdersId"`
	DoNotMoveStock bool        `json:"DoNotMoveStock"`
	GoodsValue     int64       `json:"GoodsValue"`
	DiscountAmount int64       `json:"DiscountAmount"`
	TaxPercentage  float64     `json:"TaxPercentage"`
	TaxAmount      int64       `json:"TaxAmount"`
	CourierCharges int64       `json:"CourierCharges"`
	InvoiceAmount  int64       `json:"InvoiceAmount"`
}
