package file

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mrssss/microprobe/blueprint"
	"github.com/mrssss/microprobe/client"
)

type ElasticSearchClient struct {
	addresses []string
	result    string
	done      *sync.WaitGroup
	running   bool
}

var latestDocId = 0

func toDatetime(str string) (*time.Time, error) {
	//layout := "2006-01-02T15:04:05.999999-07:00"
	t, err := time.Parse(time.RFC3339, str)

	if err != nil {
		return nil, err
	}
	return &t, nil
}

var start_time = time.Now().Add(-1 * time.Minute)

func (esc *ElasticSearchClient) Process() {
	esc.running = true
	esc.done.Add(1)
	defer esc.done.Done()

	f, err := os.Create(esc.result)
	if err != nil {
		fmt.Printf("failed to create file %+v", err)
	}
	defer f.Close()

	out := bufio.NewWriter(f)

	defer out.Flush()

	//cert, _ := ioutil.ReadFile(es.ca_cert)
	cfg := elasticsearch.Config{
		Addresses: esc.addresses,
	}

	cli, err := elasticsearch.NewClient(cfg)

	if err != nil {
		fmt.Printf("")
		os.Exit(1)
	}

	for {
		var buf bytes.Buffer
		var r map[string]interface{}

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": map[string]interface{}{
						"range": map[string]interface{}{
							"timestamp": map[string]string{
								"gte": start_time.Format(time.RFC3339),
								"lt":  "now",
							},
						},
					},
					"filter": map[string]interface{}{
						"range": map[string]interface{}{
							"humidity": map[string]interface{}{
								"gte": 100,
							},
						},
					},
				},
			},
		}
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			log.Fatalf("Error encoding query: %s", err)
		}
		// Perform the search request.

		query_start := time.Now()
		res, err := cli.Search(
			cli.Search.WithContext(context.Background()),
			cli.Search.WithIndex("test"),
			cli.Search.WithBody(&buf),
			cli.Search.WithTrackTotalHits(true),
			cli.Search.WithPretty(),
		)
		query_response_time := time.Now().Sub(query_start)
		if err != nil {
			log.Fatalf("Error getting response: %s", err)
		}

		if res.IsError() {
			var e map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
				log.Fatalf("Error parsing the response body: %s", err)
			} else {
				// Print the response status and error information.
				log.Fatalf("[%s] %s: %s",
					res.Status(),
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"],
				)
			}
		}

		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		}
		res.Body.Close()
		io.Copy(ioutil.Discard, res.Body)
		// Print the response status, number of results, and request duration.
		//log.Printf(
		//	"[%s] %d hits; took: %dms",
		//	res.Status(),
		//	int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		//	int(r["took"].(float64)),
		//)
		// Print the ID and document source for each hit.
		for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
			docId, _ := hit.(map[string]interface{})["_id"].(string)
			iDocId, _ := strconv.Atoi(docId)
			if iDocId > latestDocId {
				t0, err := toDatetime(hit.(map[string]interface{})["_source"].(map[string]interface{})["timestamp"].(string))
				if err != nil {
					log.Printf("failed to convert ", err)
					return
				}

				if start_time.Before(*t0) {
					start_time = *t0
				}

				t1 := time.Now()

				l := fmt.Sprintf("%s,%s,%s,%s,%s,%f\n",
					query_response_time,
					t1.Sub(*t0),
					time.Now().String(),
					hit.(map[string]interface{})["_source"].(map[string]interface{})["timestamp"],
					hit.(map[string]interface{})["_source"].(map[string]interface{})["devicename"],
					hit.(map[string]interface{})["_source"].(map[string]interface{})["humidity"].(float64))
				out.WriteString(l)
				out.Flush()
				log.Printf(l)
				latestDocId = iDocId
			}
		}
	}
	//for {
	//	cli.Search
	//}

}

func (esc *ElasticSearchClient) Stop() {
	esc.running = false
}

func init() {
	client.RegisterClient("elastic_search", NewElasticSearchClient)
}

func NewElasticSearchClient(es *blueprint.ClientConfig, done *sync.WaitGroup) (client.Client, error) {
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

	return &ElasticSearchClient{
		addresses: addresses,
		result:    es.Result,
		done:      done,
		running:   true,
	}, nil
}
