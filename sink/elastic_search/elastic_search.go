package file

import (
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/mrssss/microprobe/blueprint"
	"github.com/mrssss/microprobe/sink"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ElasticSearchStorage struct {
	addresses []string
	event     chan string
	done      *sync.WaitGroup
	running   bool
}

func (es *ElasticSearchStorage) Process() {
	es.running = true
	es.done.Add(1)
	defer es.done.Done()

	//cert, _ := ioutil.ReadFile(es.ca_cert)
	cfg := elasticsearch.Config{
		Addresses: es.addresses,
	}

	client, err := elasticsearch.NewClient(cfg)

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         "test",                // The default index name
		Client:        client,                // The Elasticsearch client
		NumWorkers:    10,                    // The number of worker goroutines
		FlushBytes:    int(1024),             // The flush threshold in bytes
		FlushInterval: 10 * time.Millisecond, // The periodic flush interval
	})
	defer bi.Close(context.Background())

	if err != nil {
		fmt.Printf("")
		os.Exit(1)
	}

	var ind int = 0
	var countSuccessful uint64 = 0
	cur_time := time.Now()
	start_time := time.Now()
Stop:
	for {
		select {
		case res := <-es.event:
			event := strings.TrimSpace(res)
			if event != "" {
				err = bi.Add(
					context.Background(),
					esutil.BulkIndexerItem{
						// Action field configures the operation to perform (index, create, delete, update)
						Action: "index",

						// DocumentID is the (optional) document ID
						DocumentID: strconv.Itoa(ind),

						// Body is an `io.Reader` with the payload
						Body: strings.NewReader(event),

						// OnSuccess is called for each successful operation
						OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
							atomic.AddUint64(&countSuccessful, 1)
						},

						// OnFailure is called for each failed operation
						OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
							if err != nil {
								log.Printf("ERROR: %s", err)
							} else {
								log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
							}
						},
					},
				)
				ind++
			}
			if time.Now().Sub(cur_time).Seconds() >= 1 {
				cur_time = time.Now()
				log.Printf("index %d events in %f milliseconds", atomic.LoadUint64(&countSuccessful), float64(time.Now().Sub(start_time).Milliseconds()))
			}
		case <-time.After(1 * time.Second):
			if !es.running {
				break Stop
			}
		}
	}

	log.Printf("indexing document ID=%d", ind)
}

func (es *ElasticSearchStorage) Stop() {
	es.running = false
}

func init() {
	sink.RegisterSink("elastic_search", NewElasticSearchStorage)
}

func NewElasticSearchStorage(es *blueprint.SinkConfig, event chan string, done *sync.WaitGroup) (sink.Sink, error) {
	ctx, ok := es.Ctx.(map[interface{}]interface{})
	addresses := []string{}

	if ok {
		for k, v := range ctx {
			kk, okk := k.(string)
			if !okk {
				fmt.Printf("Invalid key")
				os.Exit(1)
			}
			if kk == "addresses" {
				vv, okv := v.([]interface{})
				if okv {
					for _, vaddr := range vv {
						addr, oka := vaddr.(string)
						if oka {
							addresses = append(addresses, addr)
						}
					}
				}
			}
		}
	}

	if len(addresses) == 0 {
		fmt.Printf("Invalid addresses")
		os.Exit(1)
	}

	return &ElasticSearchStorage{
		addresses: addresses,
		event:     event,
		done:      done,
		running:   true,
	}, nil
}
