package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/olivere/elastic"
	streamhttp "github.com/cabernety/gopkg/stream/http"
	"encoding/json"
	"context"
	"fmt"
	"net/http"
	"time"
	"io"
	"github.com/cabernety/gopkg/stream"
)

type logResult struct {
	log string
	timestamp string
	id string
}

func getLogs(ctx context.Context) []*logResult {

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
	return output
}

func writeCh(ch chan []byte) {
	i := 0
	for {
		time.Sleep(time.Second*3)
		ch <- []byte(fmt.Sprintf("i :> %d", i))
		i ++
		if i > 20 {
			break
		}
	}
}

func main(){
	logrus.SetLevel(logrus.DebugLevel)
	addr := ":8001"

	server := &http.Server{
		Addr: addr, Handler: nil,
	}

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request){
		ch := make(chan []byte)
		//ctx := context.Background()
		//go getLogs(ctx)
		rw.Header().Set("Content-Type", "text/plain")

		// Chrome won't show data if we don't set this. See
		// http://stackoverflow.com/questions/26164705/chrome-not-handling-chunked-responses-like-firefox-safari.
		rw.Header().Set("X-Content-Type-Options", "nosniff")
		w := streamhttp.StreamingResponseWriter(rw)
		defer close(stream.Heartbeat(w, time.Second*25)) // Send a null character every 25 seconds.
		go writeCh(ch)
		select {
		case msg := <- ch:
			logrus.Infof("write ch : %s", string(msg))
			io.WriteString(w, string(msg))
			//sw.WriteHeader(http.StatusOK)
		}

	})
	logrus.Infof("Server listen on %s", addr)
	logrus.Fatal(server.ListenAndServe())
}

func main1(){
	logrus.SetLevel(logrus.DebugLevel)
	addr := ":8001"

	server := &http.Server{
		Addr: addr, Handler: nil,
		WriteTimeout: time.Second * 300,
		}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		ch := make(chan []byte)
		//ctx := context.Background()
		//go getLogs(ctx)
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Connection", "keepalive")
		w.Header().Set("Transfer-Encoding", "chunked")
		// We are going to return json no matter what:
		// Don't cache response:
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		w.Header().Set("Expires", "0")
		// Chrome won't show data if we don't set this. See
		// http://stackoverflow.com/questions/26164705/chrome-not-handling-chunked-responses-like-firefox-safari.
		w.Header().Set("X-Content-Type-Options", "nosniff")
		disconnectNotify := w.(http.CloseNotifier).CloseNotify()
		go writeCh(ch)
		select {
		case msg := <- ch:
			logrus.Infof("write ch : %s", string(msg))
			io.WriteString(w, string(msg))
			//sw.WriteHeader(http.StatusOK)
		case <-disconnectNotify:
			logrus.Info("disconnectNotify")
			break
		}

	})
	logrus.Infof("Server listen on %s", addr)
	logrus.Fatal(server.ListenAndServe())
}
