package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/olivere/elastic"
	"encoding/json"
	"context"
	"time"
)

func main(){
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()

	client, err := elastic.NewClient(elastic.SetURL("https://es.boxlinker.com"))
	if err != nil {
		panic(err)
	}
	// Search with a term query
	termQuery := elastic.NewTermQuery("docker.container_id", "d73e02f27f3fa070c53a3408c8f6f25d30cc6138636b937d22bf21c8474b1754")
	elastic.NewMultiTermvectorItem()
	searchResult, err := client.Search().
		Index("logstash-2017.11.07").   // search in index "twitter"
		Query(termQuery).   // specify the query
		Sort("@timestamp", false). // sort by "user" field, ascending
		From(0).Size(9).   // take documents 0-9
		Pretty(true).       // pretty print request and response JSON
		Do(ctx)             // execute

	if err != nil {
		panic(err)
	}
	//output := make([]*logResult, 0)

	for _, hit := range searchResult.Hits.Hits {
		var item map[string]interface{}
		b, err := hit.Source.MarshalJSON()
		if err != nil {
			panic(err)
			continue
		}
		if err := json.Unmarshal(b, &item); err != nil {
			panic(err)
			continue
		}
		logrus.Debugf("%s", hit.Uid)
		logrus.Debugf("%s", item["@timestamp"])
		logrus.Debugf("%s", item["log"])
		logrus.Debug("=======")
	}
}
