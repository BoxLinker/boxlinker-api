package application

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"fmt"
	"github.com/BoxLinker/boxlinker-api/modules/logs/es"
	"io/ioutil"
	"github.com/cabernety/gopkg/httplib"
	"time"
	"io"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"errors"
)

type Result struct {
	Hits struct{
		Hits []Hit `json:"hits"`
	} `json:"hits"`
}

type Hit struct {
	ID string `json:"_id"`
	Source struct{
		Log string `json:"log"`
		Timestamp string `json:"@timestamp"`
	} `json:"_source"`
}

type lreader struct {
	done chan struct{}
	reader io.Reader
	bufCh chan []byte
	errCh chan error
}

func (r *lreader) start(){
	buf := make([]byte, 32*1024)
	for {
		nr, er := r.reader.Read(buf)
		if nr > 0 {
			r.bufCh <- buf[0:nr]
		}
		if er != nil {
			r.errCh <- er
			break
		}
	}
}

/**
 *	@param {string} startTime 日志的起始时间，格式为 `2017-11-11T05:22:37.000882442Z` 或者不传
 */
func (a *Api) Log(w http.ResponseWriter, r *http.Request){

	containerID := mux.Vars(r)["containerID"]

	startTime := boxlinker.GetQueryParam(r, "start_time")

	if startTime == "" {
		startTime = "now-5m" // 默认获取 5 分钟以内的
	}

	logger := es.NewLogger(&es.LoggerOptions{
		SearchFunc: func() (string, error) {
			logrus.Debugf("request elasticsearch with containerID(%s) and startTime(%s)", containerID, startTime)
			res, err := httplib.Get(fmt.Sprintf(
				"https://es.boxlinker.com/%s/fluentd/_search?filter_path=took,hits.hits._id,hits.hits._source.log,hits.hits._source.@timestamp",
				"logstash-2017.11.11",
			)).Body(fmt.Sprintf(
				`
						{
						  "query": {
							"bool": {
							  "filter": [{
								"term": {
								  "docker.container_id": "%s"
								}
							  },{
								"range": {
								  "@timestamp": {
									"gte": "%s",
									"lte": "now"
								  }
								}
							  }]
							}
						  }
						}
					`,
				containerID,
				startTime,
			)).SetTimeout(time.Second*10, time.Second*10).Response()
			if err != nil {
				return "", err
			}
			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return "", err
			}

			// 解析结果，并获取最后一条的时间戳
			var result Result
			if err := json.Unmarshal(b, &result); err != nil {
				return "", err
			}
			hits := result.Hits.Hits
			if len(hits) <= 0 {
				return "", nil
			}
			hit := hits[len(hits) - 1]
			startTime = hit.Source.Timestamp

			return string(b), nil
		},
	})

	rr, err := logger.Open("")
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	// Chrome won't show data if we don't set this. See
	// http://stackoverflow.com/questions/26164705/chrome-not-handling-chunked-responses-like-firefox-safari.
	w.Header().Set("X-Content-Type-Options", "nosniff")

	disconnectNotify := w.(http.CloseNotifier).CloseNotify()
	bufCh := make(chan []byte)
	errCh := make(chan error)
	exitCh := make(chan error)

	go func(){
		buf := make([]byte, 32*1024)
		defer logrus.Debug("===>>>>>")
		for {
			nr, er := rr.Read(buf)
			logrus.Debugf("read %d", nr)
			if nr > 0 {
				bufCh <- buf[0:nr]
			}
			if er != nil {
				errCh <- er
				break
			}
		}
	}()

	select {
	case buf := <-bufCh:
		logrus.Debugf("buf: %s", string(buf))
		w.Write(buf)
	case <-disconnectNotify:
		logrus.Debug("disconnectNotify")
		rr.(*es.Reader).Close()
		exitCh <- errors.New("disconnectNotify")
		break
	case err := <-errCh:
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		break
	}


	go func(){

	}()


	//rw := streamhttp.StreamingResponseWriter(w)
	//defer close(stream.Heartbeat(w, time.Second*25)) // Send a null character every 25 seconds.
	//
	//go func(){
	//	select {
	//	case <-disconnectNotify:
	//	}
	//}()
	//
	//if _, err := io.Copy(rw, rr); err != nil {
	//	boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
	//	return
	//}


}
