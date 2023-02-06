package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/marktwtn/steam-wishlists-combination-generator/crawler"
)

type Combination struct {
	total_price     uint
	wishitems_index []uint
}

type Config struct {
	key           string
	default_value int
}

const UNSELECTED_MAX int = 10

var wishitems_with_selected []crawler.Wishitem
var wishitems_without_selected []crawler.Wishitem
var configs = map[string]Config{
	"diff":        {"diff", 20},
	"lower_bound": {"lower_bound", 100},
	"upper_bound": {"upper_bound", 2000},
}

func main() {
	new_app := app.NewWithID("steam-wishlists-combination-generator")
	new_app.Settings().SetTheme(&new_theme{})
	window := new_app.NewWindow("Steam願望清單最佳組合程式")
	window.SetMaster()

	var up = container.NewVBox()
	var up_crawler = widget.NewForm()
	var url_binding = binding.NewString()
	up_crawler.AppendItem(create_url_widget(new_app, &url_binding))
	var scroll_times_binding = binding.NewFloat()
	var scroll_progress_channel = make(chan int, 10)
	var scroll_max_channel = make(chan int, 1)
	up_crawler.AppendItem(create_progress_widget(&scroll_times_binding, scroll_progress_channel, scroll_max_channel))
	up.Add(up_crawler)
	up.Add(widget.NewSeparator())
	var up_config = widget.NewForm()
	var diff_binding = set_default_and_bind_value(configs["diff"], new_app.Preferences())
	up_config.AppendItem(create_diff_widget(&diff_binding))
	var lower_bound_binding = set_default_and_bind_value(configs["lower_bound"], new_app.Preferences())
	var upper_bound_binding = set_default_and_bind_value(configs["upper_bound"], new_app.Preferences())
	up_config.AppendItem(create_budget_widget(&lower_bound_binding, &upper_bound_binding))
	var unselected_number int
	up_config.AppendItem(create_select_limit_widget(new_app, &unselected_number))
	up.Add(up_config)
	var down = container.NewHBox()
	var status = widget.NewLabel("無願望清單")
	var main_box = container.NewBorder(widget.NewSeparator(), widget.NewSeparator(), nil, nil, status)
	var box = container.NewBorder(up, down, nil, nil, main_box)

	var wishitems []crawler.Wishitem
	var check_list []binding.Bool
	var combination_count_binding = binding.NewFloat()
	var combination_channel = make(chan int, 100)
	var combination_progress = widget.NewProgressBar()
	go func() {
		for {
			combination_count_binding.Set(float64(<-combination_channel))
		}
	}()
	combination_progress.Bind(combination_count_binding)
	down.Add(widget.NewButton("從網址抓取資料", func() {
		var reset = func() {
			scroll_times_binding.Set(0)
			check_list = nil
			main_box.RemoveAll()
		}
		reset()

		status = widget.NewLabel("抓取資料中......")
		main_box = container.NewBorder(widget.NewSeparator(), widget.NewSeparator(), nil, nil, status)
		box = container.NewBorder(up, down, nil, nil, main_box)
		window.SetContent(box)
		var url, _ = url_binding.Get()
		wishitems = crawler.Get_wishlist(url, scroll_progress_channel, scroll_max_channel)
		sort.SliceStable(wishitems, func(i, j int) bool {
			return wishitems[i].Get_discount_price() < wishitems[j].Get_discount_price()
		})
		main_box.RemoveAll()

		status = widget.NewLabel("可勾選必列入組合結果的遊戲")
		for index := 0; index < len(wishitems); index++ {
			check_list = append(check_list, binding.NewBool())
		}
		var new_box_for_scroll = container.NewVBox()
		for index, wishitem := range wishitems {
			check_and_name := widget.NewCheckWithData(wishitem.Get_name(), check_list[index])
			price := container.NewGridWithColumns(4, widget.NewLabel(wishitem.Get_discount_price_str()), widget.NewLabel(wishitem.Get_discount_percent_str()))
			item := container.NewGridWithColumns(4, check_and_name, price)
			new_box_for_scroll.Add(item)
		}
		var scroll = container.NewVScroll(new_box_for_scroll)
		main_box = container.NewBorder(widget.NewSeparator(), widget.NewSeparator(), nil, nil, container.NewBorder(container.NewVBox(status, container.NewGridWithColumns(1, widget.NewLabel("組合結果處理進度: "), combination_progress)), nil, nil, nil, scroll))
		box = container.NewBorder(up, down, nil, nil, main_box)
		window.SetContent(box)
	}))
	down.Add(widget.NewButton("產生組合結果並存檔", func() {
		wishitems_with_selected = []crawler.Wishitem{}
		wishitems_without_selected = []crawler.Wishitem{}
		var without_selected_index = 0
		var selected_index = 0
		for index, ele := range check_list {
			selected, _ := ele.Get()
			if selected {
				wishitems_with_selected = append(wishitems_with_selected, wishitems[index])
				wishitems_with_selected[selected_index].Set_index(uint(selected_index))
				selected_index++
			} else {
				wishitems_without_selected = append(wishitems_without_selected, wishitems[index])
				wishitems_without_selected[without_selected_index].Set_index(uint(without_selected_index))
				without_selected_index++
			}
		}
		limit := min(unselected_number, len(wishitems_without_selected), UNSELECTED_MAX)
		combination_progress.Max = float64(limit)
		diff, _ := diff_binding.Get()
		lower_bound, _ := lower_bound_binding.Get()
		upper_bound, _ := upper_bound_binding.Get()
		combinations := generate_filtered_combination(limit, upper_bound, wishitems_without_selected, combination_channel)
		acceptable_combination_list := get_acceptable_combination(uint(diff), lower_bound, upper_bound, combinations)
		file_save := dialog.NewFileSave(
			func(writer fyne.URIWriteCloser, err error) {
				if writer != nil {
					new_app.Preferences().SetString("save_uri", writer.URI().String())
					write_data(writer, acceptable_combination_list)
				}
			},
			window)
		file_save.SetFileName("steam願望清單組合")
		if new_app.Preferences().StringWithFallback("save_uri", "") != "" {
			save_uri := remove_file_in_uri(new_app.Preferences().String("save_uri"))
			uri, _ := storage.ParseURI(save_uri)
			location, _ := storage.ListerForURI(uri)
			file_save.SetLocation(location)
		}
		file_save.Show()
	}))
	down.Add(widget.NewLabel(
		"注意: 請確保願望清單的網址正確，或是願望清單有被設定成公開(即無痕視窗也可以觀看)，以及有安裝 google chrome 瀏覽器，否則程式會卡住/閃退"))

	window.SetContent(box)
	window.ShowAndRun()
}

func create_url_widget(app fyne.App, url_binding *binding.String) *widget.FormItem {
	(*url_binding).Set(app.Preferences().String("url"))
	var url_entry = widget.NewEntryWithData(*url_binding)
	url_entry.OnCursorChanged = func() {
		var url, _ = (*url_binding).Get()
		app.Preferences().SetString("url", url)
	}

	return widget.NewFormItem("願望清單網址", url_entry)
}

func create_progress_widget(scroll_times_binding *binding.Float, scroll_progress_channel chan int, scroll_max_channel chan int) *widget.FormItem {
	var progress = widget.NewProgressBar()
	progress.Bind(*scroll_times_binding)
	go func() {
		for {
			(*scroll_times_binding).Set(float64(<-scroll_progress_channel))
		}
	}()
	go func() {
		for {
			progress.Max = float64(<-scroll_max_channel)
		}
	}()

	return widget.NewFormItem("抓取願望清單進度", progress)
}

func set_default_and_bind_value(config Config, preference fyne.Preferences) binding.Int {
	var val = preference.IntWithFallback(config.key, config.default_value)
	var bind = binding.BindPreferenceInt(config.key, preference)
	bind.Set(val)
	return bind
}

func create_diff_widget(diff_binding *binding.Int) *widget.FormItem {
	var diff_entry = widget.NewEntryWithData(binding.IntToString(*diff_binding))
	diff_entry.Validator = validation.NewRegexp("^[0-9]{0,2}$", "請輸入介於 0 ~ 99 的數字")

	return widget.NewFormItem("金額與信用卡折扣的可容忍差額", diff_entry)
}

func create_budget_widget(lower_bound_binding *binding.Int, upper_bound_binding *binding.Int) *widget.FormItem {
	lower_bound, _ := (*lower_bound_binding).Get()
	if lower_bound <= 100 {
		(*lower_bound_binding).Set(100)
	}
	var lower_bound_widget = widget.NewEntryWithData(binding.IntToString(*lower_bound_binding))
	lower_bound_widget.Validator = validation.NewRegexp("^[0-9]*$", "請輸入大於 0 的數字")

	var tilde = widget.NewLabel("~")
	tilde.Alignment = fyne.TextAlignCenter

	var upper_bound_widget = widget.NewEntryWithData(binding.IntToString(*upper_bound_binding))
	upper_bound_widget.Validator = validation.NewRegexp("^[0-9]*$", "請輸入大於 0 的數字")

	var budget_info = widget.NewLabel("")

	var check_budget = func() {
		lower_bound, _ := (*lower_bound_binding).Get()
		upper_bound, _ := (*upper_bound_binding).Get()
		if is_budget_valid(lower_bound, upper_bound) {
			budget_info.SetText("")
		} else {
			budget_info.SetText("警告: 不合理的預算範圍")
		}
	}
	lower_bound_widget.OnCursorChanged = check_budget
	upper_bound_widget.OnCursorChanged = check_budget

	return widget.NewFormItem("預算範圍", container.NewGridWithRows(1, lower_bound_widget, tilde, upper_bound_widget, budget_info))
}

func create_select_limit_widget(app fyne.App, unselected_number *int) *widget.FormItem {
	var option = []string{}
	for index := 0; index <= UNSELECTED_MAX; index++ {
		option = append(option, strconv.Itoa(index))
	}
	var unselected_limit = widget.NewSelect(option, func(data string) {
		*unselected_number, _ = strconv.Atoi(data)
		app.Preferences().SetInt("limit", *unselected_number)
	})
	unselected_limit.SetSelected(option[app.Preferences().Int("limit")])

	return widget.NewFormItem("搭配非勾選的遊戲上限數量", container.NewGridWithRows(1, unselected_limit, widget.NewLabel("願望清單越多，「搭配非勾選的遊戲上限數量」數值設定越高，產出組合的時間越長")))
}

func is_budget_valid(lower_bound int, upper_bound int) bool {
	if lower_bound < 0 || upper_bound < 0 {
		return false
	}
	if lower_bound > upper_bound {
		return false
	}
	return true
}

func min(a int, b int, c int) int {
	if a < b {
		if a < c {
			return a
		}
	} else {
		if b < c {
			return b
		}
	}
	return c
}

func generate_filtered_combination(unselected_count int, price_limit int, wishitems []crawler.Wishitem, combination_channel chan int) [][]Combination {
	var result [][]Combination
	for index := 0; index <= len(wishitems); index++ {
		result = append(result, []Combination{})
	}
	// Total item in combination = 1
	for _, wishitem := range wishitems {
		var combination Combination = Combination{wishitem.Get_discount_price(), []uint{wishitem.Get_index()}}
		if combination.total_price < uint(price_limit) {
			result[1] = append(result[1], combination)
		}
	}
	// Total items in combination > 1
	for index := 2; index <= unselected_count; index++ {
		for _, prev_combination := range result[index-1] {
			var last_item_index uint = prev_combination.wishitems_index[len(prev_combination.wishitems_index)-1]
			if last_item_index != uint(len(wishitems))-1 {
				for _, wishitem := range wishitems[last_item_index+1:] {
					new_wishitems_index := make([]uint, len(prev_combination.wishitems_index))
					copy(new_wishitems_index, prev_combination.wishitems_index)
					var new_combination Combination = Combination{prev_combination.total_price + wishitem.Get_discount_price(), append(new_wishitems_index, wishitem.Get_index())}
					if new_combination.total_price < uint(price_limit) {
						result[index] = append(result[index], new_combination)
					}
				}
			}
		}
		combination_channel <- index
	}
	return result
}

func get_acceptable_combination(diff uint, lower_bound int, upper_bound int, combinations_list [][]Combination) []Combination {
	var acceptable_combination []Combination
	var selected_total_price uint = 0
	for _, wishitem := range wishitems_with_selected {
		selected_total_price += wishitem.Get_discount_price()
	}
	for _, combinations := range combinations_list {
		for _, combination := range combinations {
			var total = selected_total_price + combination.total_price
			if total >= 100 && total%100 <= diff && total >= uint(lower_bound) && total <= uint(upper_bound) {
				acceptable_combination = append(acceptable_combination, combination)
			}
		}
	}
	return acceptable_combination
}

func remove_file_in_uri(uri string) string {
	last_slash_index := strings.LastIndex(uri, "/")
	return uri[0:last_slash_index]
}

func write_data(writer fyne.URIWriteCloser, combinations []Combination) {
	var selected_info = ""
	var selected_total_price uint = 0
	for _, wishitem := range wishitems_with_selected {
		selected_info += fmt.Sprintf("%-[1]*s %8s %s\n", 50-(len(wishitem.Get_name())-utf8.RuneCountInString(wishitem.Get_name()))/2, wishitem.Get_name(), wishitem.Get_discount_price_str(), wishitem.Get_discount_percent_str())
		selected_total_price += wishitem.Get_discount_price()
	}
	for _, combination := range combinations {
		var info string = "組合:\n"
		info += selected_info
		for _, index := range combination.wishitems_index {
			var wishitem = wishitems_without_selected[index]
			info += fmt.Sprintf("%-[1]*s %8s %s\n", 50-(len(wishitem.Get_name())-utf8.RuneCountInString(wishitem.Get_name()))/2, wishitem.Get_name(), wishitem.Get_discount_price_str(), wishitem.Get_discount_percent_str())
		}
		info += strconv.Itoa(int(selected_total_price+combination.total_price)) + "元"
		info += "\n\n"
		writer.Write([]byte(info))
	}
}
