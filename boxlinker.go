package boxlinker

import (
	"strconv"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

func ParsePageConfig(r *http.Request) PageConfig {
	return ParseHTTPQuery(r)
}

func ParseHTTPQuery(r *http.Request) PageConfig {
	q := r.URL.Query()
	current_pageS := q.Get("current_page")

	current_page, err := strconv.ParseInt(current_pageS, 10, 0)
	if err != nil || current_page <= 0 {
		current_page = 1
	}

	page_countS := q.Get("page_count")

	page_count, err := strconv.ParseInt(page_countS, 10, 0)
	if err != nil || page_count <= 0 {
		page_count = 10
	}
	pc := PageConfig{}
	pc.CurrentPage = int(current_page)
	pc.PageCount = int(page_count)
	return pc
}

func ReadRequestBody(r *http.Request, bean interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, bean); err != nil {
		return err
	}
	return nil
}