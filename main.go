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

const SCROLL_DOWN_UNIT int = 2000
const UNSELECTED_MAX int = 5

var wishitems []Wishitem
var wishitems_with_selected []Wishitem
var wishitems_without_selected []Wishitem
var combinations [][]Combination
var diff_binding binding.String = binding.NewString()
var combination_channel = make(chan int, 100)

// GUI
var combination_progress *widget.ProgressBar

func main() {
	new_app := app.New()
	new_app.Settings().SetTheme(&new_theme{})
	window := new_app.NewWindow("Steam願望清單最佳組合程式")
	window.SetMaster()

	var up = container.NewVBox()
	var url = widget.NewEntry()
	var up_0 = widget.NewForm(widget.NewFormItem("願望清單網址", url))
	up.Add(up_0)
	var progress = widget.NewProgressBar()
	progress.Max = float64(scroll_times_max)
	var scroll_times_binding = binding.NewFloat()
	progress.Bind(scroll_times_binding)
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
	up.Add(up_1)
	var up_2 = widget.NewForm(widget.NewFormItem("金額與信用卡折扣的可容忍差額", widget.NewEntryWithData(diff_binding)))
	up.Add(up_2)
	var unselected_max int
	var option = []string{}
	for index := 0; index <= UNSELECTED_MAX; index++ {
		option = append(option, strconv.Itoa(index))
	}
	var unselected_limit = widget.NewSelect(option, func(data string) {
		unselected_max, _ = strconv.Atoi(data)
	})
	unselected_limit.SetSelected(option[len(option)-1])
	var up_3 = widget.NewForm(widget.NewFormItem("搭配非勾選的遊戲上限數量", unselected_limit))
	up.Add(up_3)
	var up_4 = widget.NewForm(widget.NewFormItem("願望清單越多，「搭配非勾選的遊戲上限數量」數值設定越高，產出組合的時間越長", widget.NewLabel("")))
	up.Add(up_4)
	var down = container.NewHBox()
	var status = widget.NewLabel("無願望清單")
	var main_box = container.NewBorder(widget.NewSeparator(), widget.NewSeparator(), nil, nil, status)
	var box = container.NewBorder(up, down, nil, nil, main_box)

	var check_list []binding.Bool
	var combination_count_binding = binding.NewFloat()
	combination_progress = widget.NewProgressBar()
	go func() {
		for {
			combination_count_binding.Set(float64(<-combination_channel))
		}
	}()
	combination_progress.Bind(combination_count_binding)
	down.Add(widget.NewButton("從網址抓取資料", func() {
		main_box.RemoveAll()
		status = widget.NewLabel("抓取資料中......")
		main_box = container.NewBorder(widget.NewSeparator(), widget.NewSeparator(), nil, nil, status)
		box = container.NewBorder(up, down, nil, nil, main_box)
		window.SetContent(box)
		wishitems = get_wishlist(url.Text)
		main_box.RemoveAll()
		status = widget.NewLabel("可勾選必列入組合結果的遊戲")
		for index := 0; index < len(wishitems); index++ {
			check_list = append(check_list, binding.NewBool())
		}
		var new_box_for_scroll = container.NewVBox()
		for index, wishitem := range wishitems {
			var wishitem_info = wishitem.name + "    " + get_price_string(wishitem) + "    " + get_discount_string(wishitem)
			var check = widget.NewCheckWithData(wishitem_info, check_list[index])
			new_box_for_scroll.Add(check)
		}
		var scroll = container.NewVScroll(new_box_for_scroll)
		main_box = container.NewBorder(widget.NewSeparator(), widget.NewSeparator(), nil, nil, container.NewBorder(container.NewVBox(status, container.NewGridWithColumns(1, widget.NewLabel("組合結果處理進度: "), combination_progress)), nil, nil, nil, scroll))
		box = container.NewBorder(up, down, nil, nil, main_box)
		window.SetContent(box)
	}))
	var acceptable_combination_list []Combination
	var file_save = dialog.NewFileSave(
		func(writer fyne.URIWriteCloser, err error) {
			if writer != nil {
				write_data(writer, acceptable_combination_list)
			}
		},
		window)
	file_save.SetFileName("steam願望清單組合")
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
		var combination_max = 0
		for index := 1; index <= limit; index++ {
			combination_max += get_combination_count(index, len(wishitems_without_selected))
		}
		combination_progress.Max = float64(combination_max)
		combinations = generate_all_combination(limit, wishitems_without_selected)
		acceptable_combination_list = get_acceptable_combination(combinations)
		file_save.Show()
	}))
	down.Add(widget.NewLabel(
		"注意: 請確保願望清單的網址正確，或是願望清單有被設定成公開(即無痕視窗也可以觀看)，以及有安裝 google chrome 瀏覽器，否則程式會卡住/閃退"))

	window.SetContent(box)
	window.ShowAndRun()
}

func generate_all_combination(unselected_count int, wishitems []Wishitem) [][]Combination {
	var result [][]Combination
	var combination_count = 0
	for index := 0; index <= len(wishitems); index++ {
		result = append(result, []Combination{})
	}
	// Total item in combination = 1
	for _, wishitem := range wishitems {
		var combination Combination = Combination{wishitem.discount_price, []uint{wishitem.index}}
		result[1] = append(result[1], combination)
		combination_count++
	}
	// Total items in combination > 1
	for index := 2; index <= unselected_count; index++ {
		for _, prev_combination := range result[index-1] {
			var last_item_index uint = prev_combination.wishitems_index[len(prev_combination.wishitems_index)-1]
			if last_item_index != uint(len(wishitems))-1 {
				for _, wishitem := range wishitems[last_item_index+1:] {
					new_wishitems_index := make([]uint, len(prev_combination.wishitems_index))
					copy(new_wishitems_index, prev_combination.wishitems_index)
					var new_combination Combination = Combination{prev_combination.total_price + wishitem.discount_price, append(new_wishitems_index, wishitem.index)}
					result[index] = append(result[index], new_combination)
					combination_count++
					combination_channel <- combination_count
				}
			}
		}
	}
	return result
}

func get_combination_count(selected int, total int) int {
	if selected <= 0 || total <= 0 {
		return 0
	}
	var result = 1
	for index := 1; index <= selected; index++ {
		result = result * (total - index + 1) / index
	}
	return result
}

func get_acceptable_combination(combinations [][]Combination) []Combination {
	var acceptable_combination []Combination
	var diff, _ = diff_binding.Get()
	var diff_val, _ = strconv.Atoi(diff)
	var selected_total_price uint = 0
	for _, wishitem := range wishitems_with_selected {
		selected_total_price += wishitem.discount_price
	}
	for _, combination := range combinations {
		for _, ele := range combination {
			if selected_total_price+ele.total_price >= 100 && (selected_total_price+ele.total_price)%100 <= uint(diff_val) {
				acceptable_combination = append(acceptable_combination, ele)
			}
		}
	}
	return acceptable_combination
}

func write_data(writer fyne.URIWriteCloser, combinations []Combination) {
	for _, combination := range combinations {
		var info string = "組合:\n"
		var selected_total_price uint = 0
		for _, wishitem := range wishitems_with_selected {
			info += wishitem.name
			info += "    "
			info += get_price_string(wishitem)
			info += "    "
			info += get_discount_string(wishitem)
			selected_total_price += wishitem.discount_price
			info += "\n"
		}
		for _, index := range combination.wishitems_index {
			var wishitem = wishitems_without_selected[index]
			info += wishitem.name
			info += "    "
			info += get_price_string(wishitem)
			info += "    "
			info += get_discount_string(wishitem)
			info += "\n"
		}
		info += strconv.Itoa(int(selected_total_price+combination.total_price)) + "元"
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
