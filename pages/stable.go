package pages

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/uniswap-auto-gui/services"
	"github.com/uniswap-auto-gui/utils"
)

func stableScreen(_ fyne.Window) fyne.CanvasObject {
	dataList := binding.BindStringList(&[]string{})

	find := widget.NewButton("Find Stable Coins", func() {
		go func() {
			for {
				c1 := make(chan string, 1)
				go utils.Post(c1, "pairs", "")
				trackStables(c1, dataList)
				time.Sleep(time.Minute * 20)
			}
		}()
	})

	list := widget.NewListWithData(dataList,
		func() fyne.CanvasObject {
			leftPane := container.NewHBox(widget.NewHyperlink("DEX", parseURL("https://fyne.io/")), widget.NewLabel("token"), widget.NewLabel("address"), widget.NewLabel("price"), widget.NewLabel("change"), widget.NewLabel("duration"))
			return container.NewBorder(nil, nil, leftPane, widget.NewButton("+", nil))
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			lc := obj.(*fyne.Container).Objects[0].(*fyne.Container)

			dex := lc.Objects[0].(*widget.Hyperlink)

			f := item.(binding.String)
			text := lc.Objects[1].(*widget.Label)
			text.Bind(f)

			label := lc.Objects[2].(*widget.Label)
			price := lc.Objects[3].(*widget.Label)
			change := lc.Objects[4].(*widget.Label)
			duration := lc.Objects[5].(*widget.Label)

			btn := obj.(*fyne.Container).Objects[1].(*widget.Button)
			btn.OnTapped = func() {
				fmt.Println("Ok!")
			}

			go func() {
				for {
					var swaps utils.Swaps
					c1 := make(chan string, 1)
					pair, _ := f.Get()
					utils.Post(c1, "swaps", pair)
					msg := <-c1
					json.Unmarshal([]byte(msg), &swaps)
					n, p, c, d, a := services.SwapsInfo(swaps)
					label.SetText(n)
					price.SetText(fmt.Sprintf("%f", p))
					change.SetText(fmt.Sprintf("%f", c))
					duration.SetText(fmt.Sprintf("%f hours", d))

					url := fmt.Sprintf("https://www.dextools.io/app/ether/pair-explorer/%s", pair)
					dex.SetURL(parseURL(url))

					if a {
						services.Notify("Price Change Alert", n)
					}
				}
			}()
		})

	return container.NewBorder(find, nil, nil, nil, list)
}

func trackStables(pings <-chan string, list binding.ExternalStringList) {
	msg := <-pings
	var pairs utils.Pairs

	json.Unmarshal([]byte(msg), &pairs)

	var wg sync.WaitGroup
	wg.Add(len(pairs.Data.Pairs))
	go services.StableTokens(&wg, pairs, list)
	wg.Wait()
}