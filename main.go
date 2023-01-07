package main

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const SCROLL_DOWN_UNIT int = 2000

type Wishitem struct {
	index            uint
	name             string
	discount_price   uint
	discount_percent uint
}
type Combination struct {
	total_price     uint
	wishitems_index []uint
}

var wishlist_page = ""
var wishitems []Wishitem
var wishitems_with_selected []Wishitem
var wishitems_without_selected []Wishitem
var combinations [][]Combination
var diff_binding binding.String = binding.NewString()

// GUI
var filtered_result *fyne.Container

func main() {
	new_app := app.New()
	new_app.Settings().SetTheme(&new_theme{})
	window := new_app.NewWindow("Steam願望清單最佳組合程式")
	window.SetMaster()

	var acceptable_combination_list []Combination
	var file_save = dialog.NewFileSave(
		func(writer fyne.URIWriteCloser, err error) {
			if writer != nil {
				write_data(writer, acceptable_combination_list)
			}
		},
		window)
	file_save.SetFileName("steam願望清單組合")

	var status = widget.NewLabel("無願望清單")
	var main_box = container.NewBorder(widget.NewSeparator(), widget.NewSeparator(), nil, nil, status)
	var up = container.NewVBox()
	var url = widget.NewEntry()
	var up_0 = widget.NewForm(widget.NewFormItem("願望清單網址", url))
	var progress = widget.NewProgressBar()
	progress.Max = float64(scroll_times_max)
	var scroll_times_binding = binding.NewFloat()
	progress.Bind(scroll_times_binding)
	var unselected_max int
	var unselected_limit = widget.NewSelect([]string{"0", "1", "2", "3", "4", "5"}, func(data string) {
		unselected_max, _ = strconv.Atoi(data)
	})
	unselected_limit.SetSelected("5")
	go func() {
		for {
			scroll_times_binding.Set(float64(<-scroll_channel))
		}
	}()
	go func() {
		for {
			progress.Max = float64(<-scroll_max_channel)
		}
	}()
	var up_1 = widget.NewForm(widget.NewFormItem("抓取願望清單進度", progress))
	var up_2 = widget.NewForm(widget.NewFormItem("金額與信用卡折扣的可容忍差額", widget.NewEntryWithData(diff_binding)))
	var up_3 = widget.NewForm(widget.NewFormItem("搭配非勾選的遊戲上限數量", unselected_limit))
	up.Add(up_0)
	up.Add(up_1)
	up.Add(up_2)
	up.Add(up_3)
	up.Add(widget.NewForm(widget.NewFormItem("願望清單越多，「搭配非勾選的遊戲上限數量」數值設定越高，產出組合的時間越長", widget.NewLabel(""))))
	var down = container.NewHBox()
	box := container.NewBorder(up, down, nil, nil, main_box)
	window.SetContent(box)

	var check_list []binding.Bool
	down.Add(widget.NewButton("從網址抓取資料", func() {
		main_box.RemoveAll()
		status = widget.NewLabel("抓取資料中......")
		main_box = container.NewBorder(widget.NewSeparator(), widget.NewSeparator(), nil, nil, status)
		box = container.NewBorder(up, down, nil, nil, main_box)
		window.SetContent(box)
		wishlist_page = url.Text
		wishitems = get_wishlist()
		main_box.RemoveAll()
		status = widget.NewLabel("可勾選必列入組合結果的遊戲")
		for index := 0; index < len(wishitems); index++ {
			check_list = append(check_list, binding.NewBool())
		}
		var new_box_for_scroll = container.NewVBox()
		for index, ele := range wishitems {
			var wishitems_info = ele.name + "    " + get_price_string(ele) + "    " + get_discount_string(ele)
			var check = widget.NewCheckWithData(wishitems_info, check_list[index])
			new_box_for_scroll.Add(check)
		}
		var scroll = container.NewVScroll(new_box_for_scroll)
		main_box = container.NewBorder(widget.NewSeparator(), widget.NewSeparator(), nil, nil, container.NewBorder(status, nil, nil, nil, scroll))
		box = container.NewBorder(up, down, nil, nil, main_box)
		window.SetContent(box)
	}))

	down.Add(widget.NewButton("產生組合結果並存檔", func() {
		wishitems_with_selected = []Wishitem{}
		wishitems_without_selected = []Wishitem{}
		var without_selected_index = 0
		var selected_index = 0
		for index, ele := range check_list {
			selected, _ := ele.Get()
			if selected {
				wishitems_with_selected = append(wishitems_with_selected, wishitems[index])
				wishitems_with_selected[selected_index].index = uint(selected_index)
				selected_index++
			} else {
				wishitems_without_selected = append(wishitems_without_selected, wishitems[index])
				wishitems_without_selected[without_selected_index].index = uint(without_selected_index)
				without_selected_index++
			}
		}
		var limit int
		if unselected_max < len(wishitems_without_selected) {
			limit = unselected_max
		} else {
			limit = len(wishitems_without_selected)
		}
		if limit > 5 {
			limit = 5
		}
		combinations = generate_all_combination(limit, wishitems_without_selected)
		acceptable_combination_list = get_acceptable_combination(combinations)
		file_save.Show()
	}))

	down.Add(widget.NewLabel(
		"注意: 請確保願望清單的網址正確，或是願望清單有被設定成公開(即無痕視窗也可以觀看)，以及有安裝 google chrome 瀏覽器，否則程式會卡住/閃退"))
	window.ShowAndRun()
}

func generate_all_combination(unselected_count int, wishitems []Wishitem) [][]Combination {
	var result [][]Combination
	for index := 0; index <= len(wishitems); index++ {
		result = append(result, []Combination{})
	}
	// Total item in combination = 1
	for _, ele := range wishitems {
		var combination Combination = Combination{ele.discount_price, []uint{ele.index}}
		result[1] = append(result[1], combination)
	}
	// Total items in combination > 1
	for index := 2; index <= unselected_count; index++ {
		for _, prev_combination := range result[index-1] {
			var last_item_index uint = prev_combination.wishitems_index[len(prev_combination.wishitems_index)-1]
			if last_item_index != uint(len(wishitems))-1 {
				for _, new_ele := range wishitems[last_item_index+1:] {
					new_wishitems_index := make([]uint, len(prev_combination.wishitems_index))
					copy(new_wishitems_index, prev_combination.wishitems_index)
					var new_combination Combination = Combination{prev_combination.total_price + new_ele.discount_price, append(new_wishitems_index, new_ele.index)}
					result[index] = append(result[index], new_combination)
				}
			}
		}
	}
	return result
}

func get_acceptable_combination(combinations [][]Combination) []Combination {
	var acceptable_combination []Combination
	var diff, _ = diff_binding.Get()
	var diff_val, _ = strconv.Atoi(diff)
	var selected_total_price uint = 0
	for _, ele := range wishitems_with_selected {
		selected_total_price += ele.discount_price
	}
	for _, outer_ele := range combinations {
		for _, ele := range outer_ele {
			if selected_total_price+ele.total_price >= 100 && (selected_total_price+ele.total_price)%100 <= uint(diff_val) {
				acceptable_combination = append(acceptable_combination, ele)
			}
		}
	}
	return acceptable_combination
}

func write_data(writer fyne.URIWriteCloser, combination_list []Combination) {
	for _, com := range combination_list {
		var info string = "組合:\n"
		var selected_total_price uint = 0
		for _, ele := range wishitems_with_selected {
			info += ele.name
			info += "    "
			info += get_price_string(ele)
			info += "    "
			info += get_discount_string(ele)
			selected_total_price += ele.discount_price
			info += "\n"
		}
		for _, ele := range com.wishitems_index {
			info += wishitems_without_selected[ele].name
			info += "    "
			info += get_price_string(wishitems_without_selected[ele])
			info += "    "
			info += get_discount_string(wishitems_without_selected[ele])
			info += "\n"
		}
		info += strconv.Itoa(int(selected_total_price+com.total_price)) + "元"
		info += "\n\n"
		writer.Write([]byte(info))
	}
}

func get_price_string(wishitem Wishitem) string {
	return strconv.Itoa(int(wishitem.discount_price)) + "元"
}

func get_discount_string(wishitem Wishitem) string {
	if wishitem.discount_percent == 100 {
		return ""
	}
	return "(" + strconv.FormatFloat(float64(wishitem.discount_percent)/10, 'f', -1, 64) + "折)"
}
