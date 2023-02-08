package crawler

import (
	"strconv"
)

type Wishitem struct {
	index            uint
	name             string
	discount_price   uint
	discount_percent uint
}

func New_wishitem(new_index uint, new_name string, new_discount_price uint, new_discount_percent uint) Wishitem {
	var wishitem Wishitem
	wishitem.Set(new_index, new_name, new_discount_price, new_discount_percent)
	return wishitem
}

func (w Wishitem) Set(new_index uint, new_name string, new_discount_price uint, new_discount_percent uint) {
	w.index = new_index
	w.name = new_name
	w.discount_price = new_discount_price
	w.discount_percent = new_discount_percent
}

func (w Wishitem) Get_index() uint {
	return w.index
}

func (w *Wishitem) Set_index(new_index uint) {
	w.index = new_index
}

func (w Wishitem) Get_name() string {
	return w.name
}

func (w Wishitem) Get_discount_price() uint {
	return w.discount_price
}

func (w Wishitem) Get_discount_price_str() string {
	return strconv.Itoa(int(w.discount_price)) + "元"
}

func (w Wishitem) Get_discount_percent() uint {
	return w.discount_percent
}

func (w Wishitem) Get_discount_percent_str() string {
	if w.discount_percent == 100 {
		return ""
	}
	return "(" + strconv.FormatFloat(float64(w.discount_percent)/10, 'f', -1, 64) + "折)"
}
