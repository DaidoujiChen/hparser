package hparser

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/antchfx/xpath"
	"github.com/antchfx/xquery/html"
)

// ids 從一串網址裡面, 把他們的 id 切出來
func idsFrom(urls []string) (ids [][2]string) {
	for _, v := range urls {
		id := [2]string{v[23:30], v[31:41]}
		ids = append(ids, id)
	}
	return
}

// gDataFromURLs 從 gallery 的網址取得 GMetaData 們
func gDataFromURLs(urls []string) (g GMetaDatas, err error) {

	// 準備 post 內容
	ids := idsFrom(urls)
	postBody := map[string]interface{}{
		"method":  "gdata",
		"gidlist": ids,
	}
	jsonData, err := json.Marshal(postBody)
	if err != nil {
		return g, err
	}

	// POST 到解進去 GMetaDatas 裡面
	resp, err := http.Post("https://e-hentai.org/api.php", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return g, err
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return g, err
	}
	err = json.Unmarshal(resBody, &g)
	if err != nil {
		return g, err
	}

	return g, nil
}

// TODO: 這邊目前只有單純 query 首頁, 不含任何 filter

// List 取得 Galleries 列表
func List() (g GMetaDatas, err error) {

	// GET 官網首頁
	resp, err := http.Get("https://e-hentai.org/")
	if err != nil {
		return g, err
	}
	defer resp.Body.Close()

	// parse 列表 URL 們
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return g, err
	}
	expr := xpath.MustCompile("//div [@class='it5']//a/@href")
	iter := expr.Evaluate(htmlquery.CreateXPathNavigator(doc)).(*xpath.NodeIterator)
	urls := []string{}
	for iter.MoveNext() {
		urls = append(urls, iter.Current().Value())
	}
	return gDataFromURLs(urls)
}
