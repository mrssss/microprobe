package file

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
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

	if err != nil {
		fmt.Printf("")
		os.Exit(1)
	}

	var ind uint64 = 0
	var wg sync.WaitGroup
Stop:
	for {
		select {
		case res := <-es.event:
			event := strings.TrimSpace(res)
			if event != "" {

				go func(docId uint64, e string) {
					wg.Add(1)
					defer wg.Done()

					log.Printf("indexing document ID=%d", docId)
					req := esapi.IndexRequest{
						Index:      "test",
						DocumentID: strconv.FormatUint(docId, 10),
						Body:       strings.NewReader(e),
						Refresh:    "true",
					}
					// Perform the request with the client.
					res, err := req.Do(context.Background(), client)
					if err != nil {
						log.Fatalf("Error getting response: %s", err)
					}
					defer res.Body.Close()

					if res.IsError() {
						log.Printf("[%s] Error indexing document ID=%d", res.Status(), ind)
					} else {
						// Deserialize the response into a map.
						var r map[string]interface{}
						if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
							log.Printf("Error parsing the response body: %s", err)
						} else {
							// Print the response status and indexed document version.
							//log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
						}
					}
				}(atomic.AddUint64(&ind, 1), event)
			}
		case <-time.After(1 * time.Second):
			if !es.running {
				wg.Wait()
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
