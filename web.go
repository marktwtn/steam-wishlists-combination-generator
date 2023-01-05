package main

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

var scroll_times int = 0
var scroll_times_max int = 1
var scroll_channel = make(chan int, 3)
var scroll_max_channel = make(chan int, 1)

func get_wishlist() []Wishitem {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	height := detect_webpage_height(ctx)
	scroll_max_channel <- int(height) + SCROLL_DOWN_UNIT

	var data []*cdp.Node
	var tasks chromedp.Tasks
	var wishlist = make(map[string]Wishitem)
	scroll_times_max = int(height) + SCROLL_DOWN_UNIT
	var scroll_count = 0
	for scroll_times = 0; scroll_times*SCROLL_DOWN_UNIT < int(height)+SCROLL_DOWN_UNIT; scroll_times++ {
		// Scroll down
		tasks = append(tasks, chromedp.Evaluate("window.scrollTo(0, "+strconv.Itoa(scroll_times*SCROLL_DOWN_UNIT)+");", nil))
		// Loading
		// TODO: Determine loading time
		tasks = append(tasks,
			chromedp.ActionFunc(func(context.Context) error {
				time.Sleep(1 * time.Second)
				return nil
			}))
		// Find nodes
		tasks = append(tasks, chromedp.Nodes("#wishlist_ctn .content", &data, chromedp.NodeVisible))
		// Get subtree of nodes
		tasks = append(tasks, chromedp.ActionFunc(func(c context.Context) error {
			for _, node := range data {
				// Ask chromedp to populate the subtree of nodes
				// Reference:
				// https://github.com/chromedp/examples/blob/069e33b4da60cf74307681fde4e18475ef19c439/subtree/main.go
				dom.RequestChildNodes(node.NodeID).WithDepth(-1).Do(c)
			}
			return nil
		}))
		// Get required data from nodes
		tasks = append(tasks,
			chromedp.ActionFunc(func(c context.Context) error {
				// TODO: Determine subtree loading time
				time.Sleep(1 * time.Second)
				for _, node := range data {
					name := get_title(node)
					if _, existed := wishlist[name]; !existed {
						if is_released(node) {
							price := get_final_price(node)
							discount_percent := get_discount_percent(node)
							wishlist[name] = Wishitem{0, name, price, discount_percent}
						}
					}
				}
				return nil
			}))
		// Send progress with channel
		tasks = append(tasks,
			chromedp.ActionFunc(func(context.Context) error {
				scroll_count++
				scroll_channel <- scroll_count * SCROLL_DOWN_UNIT
				return nil
			}))
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(wishlist_page),
		tasks,
	)
	if err != nil {
		log.Fatal(err)
	}

	var wishitems []Wishitem
	var count uint = 0
	for _, val := range wishlist {
		val.index = count
		count++
		wishitems = append(wishitems, val)
	}
	return wishitems
}

func detect_webpage_height(ctx context.Context) uint {
	var height uint
	err := chromedp.Run(ctx,
		chromedp.Navigate(wishlist_page),
		chromedp.ActionFunc(func(context.Context) error {
			time.Sleep(5 * time.Second)
			return nil
		}),
		chromedp.Evaluate("document.body.scrollHeight", &height),
	)
	if err != nil {
		log.Fatal(err)
	}
	return height
}

func get_title(node *cdp.Node) string {
	if is_attribute_existed(node, "content") {
		title_node := node.Children[0].Children[0]
		title := strings.TrimSpace(title_node.NodeValue)
		return title
	}
	return ""
}

func get_final_price(node *cdp.Node) uint {
	if is_released(node) {
		var price_node *cdp.Node
		var price int = 0
		var err error
		if is_attribute_existed(get_discount_block(node), "discount_block") {
			if is_discounted(node) {
				price_node = get_purchase_container_node(node).Children[0].Children[0].Children[1].Children[1].Children[0]
			} else {
				price_node = get_purchase_container_node(node).Children[0].Children[0].Children[0].Children[0].Children[0]
			}
			price, err = strconv.Atoi(remove_nt_and_dollar_sign(remove_thousand_comma(price_node.NodeValue)))
		}
		if err == nil {
			return uint(price)
		} else {
			return 0
		}
	}
	return 0
}

func get_discount_percent(node *cdp.Node) uint {
	if is_released(node) {
		var discount_node *cdp.Node
		var discount int = 0
		var err error
		if is_attribute_existed(get_discount_block(node), "discount_block") {
			if is_discounted(node) {
				discount_node = get_discount_block(node).Children[0].Children[0]
				discount, err = strconv.Atoi(remove_neg_and_percent_sign(discount_node.NodeValue))
				if err == nil {
					return 100 - uint(discount)
				}
			}
		}
	}
	return 100
}

func is_discounted(node *cdp.Node) bool {
	if get_discount_block(node) != nil {
		if is_attribute_existed(get_discount_block(node), "no_discount") {
			return false
		} else {
			return true
		}
	}
	return false
}

func is_released(node *cdp.Node) bool {
	return is_attribute_existed(get_purchase_container_node(node).Children[0], "purchase_area")
}

func get_discount_block(node *cdp.Node) *cdp.Node {
	if is_released(node) {
		return get_purchase_container_node(node).Children[0].Children[0]
	}
	return nil
}

func get_purchase_container_node(node *cdp.Node) *cdp.Node {
	if is_attribute_existed(node, "content") {
		return node.Children[1].Children[1]
	}
	return nil
}

func is_attribute_existed(node *cdp.Node, attr string) bool {
	for _, ele := range node.Attributes {
		if strings.Contains(ele, attr) {
			return true
		}
	}
	return false
}

func remove_nt_and_dollar_sign(input string) string {
	return strings.Trim(input, "NT$ ")
}

func remove_neg_and_percent_sign(input string) string {
	return strings.Trim(input, "-%")
}

func remove_thousand_comma(input string) string {
	return strings.ReplaceAll(input, ",", "")
}
