package application

import (
	"context"
	"github.com/olivere/elastic"
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"fmt"
)

type logResult struct {
	log string
	timestamp string
	id string
}

func (a *Api) Log(w http.ResponseWriter, r *http.Request){

	pc := boxlinker.ParsePageConfig(r)
	ctx := context.Background()

	client, err := elastic.NewClient(elastic.SetURL("https://es.boxlinker.com"))
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	// Search with a term query
	termQuery := elastic.NewTermQuery("kubernetes.container_name", "application")
	searchResult, err := client.Search().
		Index("logstash-2017.11.07").   // search in index "twitter"
		Query(termQuery).   // specify the query
		Sort("@timestamp", true). // sort by "user" field, ascending
		From(pc.Offset()).Size(pc.Limit()).   // take documents 0-9
		Pretty(true).       // pretty print request and response JSON
		Do(ctx)             // execute

	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	output := make([]*logResult, 0)

	for _, hit := range searchResult.Hits.Hits {
		var item map[string]interface{}
		b, err := hit.Source.MarshalJSON()
		if err != nil {
			continue
		}
		if err := json.Unmarshal(b, &item); err != nil {
			continue
		}
		output = append(output, &logResult{
			id: hit.Id,
			log: fmt.Sprint(item["log"]),
			timestamp: fmt.Sprint(item["@timestamp"]),
		})
	}

}
