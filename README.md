# steam-wishlists-combination-generator
產生總價格適合信用卡滿百回饋的Steam願望清單組合

## 前置準備
- 安裝 google chrome 瀏覽器
- 願望清單設置成公開

## 執行方式
- 使用執行檔
    - 到 [Release](https://github.com/marktwtn/steam-wishlists-combination-generator/releases) 頁面下載最新版本執行檔

- 使用原始碼
  - 下載專案並使用指令 `go build` 生成執行檔執行

## 執行檔使用方式
- 滑鼠左鍵雙點擊使用執行檔
- 填入願望清單網址
- 設定信用卡滿百的差額允許值。
  例如設定為 10 ，則遊戲組合總價格為 100 ~ 110、200 ~ 210、...... 皆可
- 按下「從網址抓取資料」按鈕並等候，若進度持續為 0%，則有可能是網址錯誤或是願望清單沒有設定成公開
- 抓取結果會顯示在中間，抓取完畢按下「產生組合結果並存檔」來把組合結果存成檔案
