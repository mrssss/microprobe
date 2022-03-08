package file

import (
	"bufio"
	"fmt"
	"github.com/mrssss/microprobe/blueprint"
	"github.com/mrssss/microprobe/source"
	"log"
	"os"
	"sync"
)

type LocalStorage struct {
	filename string
	event    chan string
	done     *sync.WaitGroup
	running  bool
}

func (sc *LocalStorage) Process() {
	sc.done.Add(1)
	defer sc.done.Done()
	sc.running = true
	f, err := os.Open(sc.filename)
	if err != nil {
		fmt.Printf("failed to create file %+v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		log.Println("read line")
		sc.event <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (sc *LocalStorage) Stop() {
	sc.running = false
}

func init() {
	source.RegisterSource("file", NewLocalStorage)
}

func NewLocalStorage(sc *blueprint.SourceConfig, event chan string, done *sync.WaitGroup) (source.Source, error) {
	context, ok := sc.Settings.(map[interface{}]interface{})
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
