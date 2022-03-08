package file

import (
	"bufio"
	"fmt"
	"github.com/mrssss/microprobe/blueprint"
	"github.com/mrssss/microprobe/sink"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type LocalStorage struct {
	filename string
	event    chan string
	done     *sync.WaitGroup
	running  bool
}

func (sc *LocalStorage) Process() {
	sc.running = true
	sc.done.Add(1)
	defer sc.done.Done()

	f, err := os.Create(sc.filename)
	if err != nil {
		fmt.Printf("failed to create file %+v", err)
	}
	defer f.Close()

	out := bufio.NewWriter(f)

	defer out.Flush()

Stop:
	for {
		select {
		case res := <-sc.event:
			event := strings.TrimSpace(res)
			if event != "" {
				out.WriteString(strings.TrimSpace(event))
				out.WriteString("\n")
				out.Flush()
				log.Println("write line.")
			}
			if event == "EOF" {
				break Stop
			}
		case <-time.After(1 * time.Second):
			if !sc.running {
				break Stop
			}
		}
	}
}

func (sc *LocalStorage) Stop() {
	sc.running = false
}

func init() {
	sink.RegisterSink("file", NewLocalStorage)
}

func NewLocalStorage(sc *blueprint.SinkConfig, event chan string, done *sync.WaitGroup) (sink.Sink, error) {
	context, ok := sc.Ctx.(map[interface{}]interface{})
	var filename string = "data.txt"

	if ok {
		for k, v := range context {
			kk, okk := k.(string)
			vv, okv := v.(string)
			if okk && okv && kk == "filename" {
				filename = vv
			}
		}
	}

	return &LocalStorage{filename: filename, event: event, done: done, running: true}, nil
}
