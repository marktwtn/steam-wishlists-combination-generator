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
	index          uint
	name           string
	discount_price uint
}
type Combination struct {
	total_price     uint
	wishitems_index []uint
}

var wishlist_page = ""
var wishitems []Wishitem
var combinations [][]Combination
var diff_binding binding.String = binding.NewString()

// GUI
var filtered_result *fyne.Container
var wishlist *widget.List = widget.NewList(
	func() int {
		return 1
	},
	func() fyne.CanvasObject {
		return widget.NewLabel("目前暫無資料")
	},
	func(index widget.ListItemID, obj fyne.CanvasObject) {
		obj.(*widget.Label).SetText("目前暫無資料")
	})

func main() {
	new_app := app.New()
	new_app.Settings().SetTheme(&new_theme{})
	window := new_app.NewWindow("Steam願望清單最佳組合程式")
	window.SetMaster()
	main_box := container.NewGridWithColumns(1)

	var acceptable_combination_list []Combination
	var file_save = dialog.NewFileSave(
		func(writer fyne.URIWriteCloser, err error) {
			if writer != nil {
				write_data(writer, acceptable_combination_list)
			}
		},
		window)
	file_save.SetFileName("steam願望清單組合")

	main_box.Add(wishlist)
	var up = container.NewVBox()
	var url = widget.NewEntry()
	var up_0 = widget.NewForm(widget.NewFormItem("願望清單網址", url))
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
	var up_2 = widget.NewForm(widget.NewFormItem("金額與信用卡折扣的可容忍差額", widget.NewEntryWithData(diff_binding)))
	up.Add(up_0)
	up.Add(up_1)
	up.Add(up_2)
	var down = container.NewHBox()
	box := container.NewBorder(up, down, nil, nil, main_box)
	window.SetContent(box)

	go func() {
		down.Add(widget.NewButton("從網址抓取資料", func() {
			wishlist_page = url.Text
			main_box.Remove(wishlist)
			wishitems = get_wishlist()
			wishlist = widget.NewList(
				func() int {
					return len(wishitems)
				},
				func() fyne.CanvasObject {
					label := widget.NewLabel("Default")
					return label
				},
				func(index widget.ListItemID, obj fyne.CanvasObject) {
					obj.(*widget.Label).SetText(wishitems[index].name)
				})
			main_box.Add(wishlist)
			window.SetContent(box)
		}))

		down.Add(widget.NewButton("產生組合結果並存檔", func() {
			combinations = generate_all_combination(wishitems)
			acceptable_combination_list = get_acceptable_combination(combinations)
			file_save.Show()
		}))

		down.Add(widget.NewLabel(
			"注意: 請確保願望清單的網址正確，或是願望清單有被設定成公開(即無痕視窗也可以觀看)，以及有安裝 google chrome 瀏覽器，否則程式會卡住/閃退"))
	}()
	window.ShowAndRun()
}

func generate_all_combination(wishitems []Wishitem) [][]Combination {
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
	for index := 2; index <= len(wishitems); index++ {
		for _, prev_combination := range result[index-1] {
			var last_item_index uint = prev_combination.wishitems_index[len(prev_combination.wishitems_index)-1]
			if last_item_index != uint(len(wishitems))-1 {
				for _, new_ele := range wishitems[last_item_index+1:] {
					var new_combination Combination = Combination{prev_combination.total_price + new_ele.discount_price, append(prev_combination.wishitems_index, new_ele.index)}
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
	for _, outer_ele := range combinations {
		for _, ele := range outer_ele {
			if ele.total_price >= 100 && ele.total_price%100 <= uint(diff_val) {
				acceptable_combination = append(acceptable_combination, ele)
			}
		}
	}
	return acceptable_combination
}

func write_data(writer fyne.URIWriteCloser, combination_list []Combination) {
	for _, com := range combination_list {
		var info string = "組合:\n"
		for _, ele := range com.wishitems_index {
			info += wishitems[ele].name
			info += "\n"
		}
		info += strconv.Itoa(int(com.total_price)) + " 元"
		info += "\n\n"
		writer.Write([]byte(info))
	}
}
