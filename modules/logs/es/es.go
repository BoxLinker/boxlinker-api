package es

import (
	"io"
	"time"
	"sync"
	"bytes"
	"fmt"
	"github.com/olivere/elastic"
	"context"
	"errors"
	"github.com/BoxLinker/boxlinker-api/modules/logs"
)

const (
	readThrottle = time.Second * 3
	writeThrottle = time.Second * 5
)

type Entity struct {
	Log 		string 			`json:"log"`
	Kubernetes 	map[string]interface{} 	`json:"kubernetes"`
	Timestamp 	string 			`json:"@timestamp"`
}

type logger struct {
	containerID string
	elasticIndex string
	ctx context.Context
	client *elastic.Client
	searchFunc func() (string, error)
}

type LoggerOptions struct {
	Client *elastic.Client
	ContainerID string
	ElasticIndex string
	Context context.Context
	SearchFunc func() (string, error)
}

func NewLogger(option *LoggerOptions) logs.Logger {
	return &logger{
		client: option.Client,
		containerID: option.ContainerID,
		elasticIndex: option.ElasticIndex,
		ctx: option.Context,
		searchFunc: option.SearchFunc,
	}
}

func (l *logger) Create(name string) (io.Writer, error) {
	return nil, errors.New("es logger not implement writer")
}

func (l *logger) Open(name string) (io.Reader, error) {
	r := &reader{
		l: l,
		throttle: time.Tick(readThrottle),
		closed: false,
		//b: lockingBuffer{},
		client: l.client,
		cid: l.containerID,
		index: l.elasticIndex,
		ctx: l.ctx,
	}
	go r.start()
	return r, nil
}

type writer struct {
	io.WriteCloser
}

func (w *writer) Close() error {
	return w.WriteCloser.Close()
}

type reader struct {
	l *logger
	ctx context.Context
	cid string
	index string
	client *elastic.Client
	io.ReadCloser
	err error
	throttle <-chan time.Time
	b lockingBuffer
	pos int64
	closed bool
}

func (r *reader) Close() error {
	r.closed = true
	return nil
}

func (r *reader) start() {
	for {
		<-r.throttle

		if r.closed {
			break
		}

		if r.err = r.read(); r.err != nil {
			fmt.Errorf("es reader err: %v", r.err)
			return
		}

	}
}

func (r *reader) read() error {
	s, err := r.l.searchFunc()
	if err != nil {
		return err
	}
	r.b.WriteString(s)
	r.pos += int64(len(s))
	return nil
}

func (r *reader) read1() error {
	termQuery := elastic.NewTermQuery("docker.container_id", r.cid)
	results, err := r.client.Search().Index(r.index).
	Query(termQuery).
	Sort("@timestamp", false).
	From(0).Size(10).
	Pretty(true).
	Do(r.ctx)
	if err != nil {
		return err
	}

	for _, hit := range results.Hits.Hits {
		b, err := hit.Source.MarshalJSON()
		if err != nil {
			return err
		}
		r.b.Write(b)
		r.pos += int64(len(b))
	}

	return nil
}
func (r *reader) Read(b []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	if r.b.Len() == 0 {
		return 0, nil
	}
	return r.b.Read(b)
}


type lockingBuffer struct {
	sync.Mutex
	bytes.Buffer
}

func (r *lockingBuffer) Read(b []byte) (int, error) {
	r.Lock()
	defer r.Unlock()

	return r.Buffer.Read(b)
}

func (r *lockingBuffer) Write(b []byte) (int, error) {
	r.Lock()
	defer r.Unlock()

	return r.Buffer.Write(b)
}