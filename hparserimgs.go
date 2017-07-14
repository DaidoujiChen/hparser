package hparser

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/antchfx/xpath"
	"github.com/antchfx/xquery/html"
)

// pageInfo 用來把 url 轉化成可用的資訊
type pageInfo struct {
	Gid    string
	Page   string
	ImgKey string
}

func (p *pageInfo) from(url string) {
	slashSplits := strings.Split(url, "/")
	slashSplitLength := len(slashSplits)
	lastSplitString := slashSplits[slashSplitLength-1]
	minusSplits := strings.Split(lastSplitString, "-")

	p.Gid = minusSplits[0]
	p.Page = minusSplits[1]
	p.ImgKey = slashSplits[slashSplitLength-2]
}

func showPageFromURL(url, showKey string) (imgURL string, err error) {
	p := pageInfo{}
	p.from(url)

	// 準備 post 內容
	postBody := map[string]interface{}{
		"method":  "showpage",
		"gid":     p.Gid,
		"page":    p.Page,
		"imgkey":  p.ImgKey,
		"showkey": showKey,
	}
	jsonData, err := json.Marshal(postBody)
	if err != nil {
		return imgURL, err
	}

	// POST 然後解出醜醜的資料
	resp, err := http.Post("https://e-hentai.org/api.php", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return imgURL, err
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return imgURL, err
	}
	var uglyData map[string]interface{}
	err = json.Unmarshal(resBody, &uglyData)
	if err != nil {
		return imgURL, err
	}

	// uglyData 裡面會有一個 i3 的 key, 裡面會有我們需要的真實圖片網址
	i3, ok := uglyData["i3"].(string)
	if !ok {
		return
	}
	r := regexp.MustCompile(`src="(\S+)"`)
	matchs := r.FindStringSubmatch(i3)
	if len(matchs) == 2 {
		return matchs[1], nil
	}

	// 到這邊就 pa 不出來
	return "", errors.New("ImageURL Not Found")
}

func showKey(url string) (match string, err error) {

	// 取得任何一個頁面先
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return match, err
	}
	req.Header.Set("Referer", url)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return match, err
	}
	defer resp.Body.Close()

	// parse 其中一段 javascript
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return match, err
	}
	expr := xpath.MustCompile("//script [@type='text/javascript']")
	iter := expr.Evaluate(htmlquery.CreateXPathNavigator(doc)).(*xpath.NodeIterator)
	for iter.MoveNext() {
		r := regexp.MustCompile(`showkey="(\w+)"`)
		matchs := r.FindStringSubmatch(iter.Current().Value())
		if len(matchs) == 2 {
			return matchs[1], nil
		}
	}

	// 到這邊的話表示說沒有 pa 到
	return "", errors.New("ShowKey Not Found")
}

// TODO: 目前應該就只有第一頁, 最多 40 張圖片

// ImagesFrom 解析出一個 gallery 中所有的圖片
func ImagesFrom(g GMetaData) (imgURLs []string, err error) {

	// 取得該 gallery 單一頁面內容
	url := "https://e-hentai.org/g/" + strconv.Itoa(g.GID) + "/" + g.Token + "/"
	resp, err := http.Get(url)
	if err != nil {
		return imgURLs, err
	}
	defer resp.Body.Close()

	// parse 該畫面上面所有的圖片連結
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return imgURLs, err
	}
	urls := []string{}
	expr := xpath.MustCompile("//div [@class='gdtm']//a/@href")
	iter := expr.Evaluate(htmlquery.CreateXPathNavigator(doc)).(*xpath.NodeIterator)
	for iter.MoveNext() {
		urls = append(urls, iter.Current().Value())
	}

	if len(urls) == 0 {
		return imgURLs, errors.New("Get Image Links Fail")
	}

	// 開始拿每一張圖片前, 需要先從其中一張拿到整個 gallery 的 showkey
	showKey, err := showKey(urls[0])
	if err != nil {
		return imgURLs, err
	}

	// 分別去取得每個圖片頁面的連結
	var wg sync.WaitGroup
	var lock sync.Mutex
	for _, v := range urls {
		wg.Add(1)
		go func(v string) {
			imgURL, err := showPageFromURL(v, showKey)
			if err == nil {
				lock.Lock()
				imgURLs = append(imgURLs, imgURL)
				lock.Unlock()
			}
			wg.Done()
		}(v)
	}
	wg.Wait()

	return imgURLs, nil
}
