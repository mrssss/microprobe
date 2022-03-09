package main

import (
	"fmt"
	"github.com/mrssss/microprobe/blueprint"
	"github.com/mrssss/microprobe/client"
	"github.com/mrssss/microprobe/sink"
	"github.com/mrssss/microprobe/source"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"sync"
	"time"

	// side effects
	_ "github.com/mrssss/microprobe/sink/file"
	// side effects
	_ "github.com/mrssss/microprobe/sink/elastic_search"
	// side effects
	_ "github.com/mrssss/microprobe/source/file"
	// side effects
	_ "github.com/mrssss/microprobe/source/models/metric"
	// side effects
	_ "github.com/mrssss/microprobe/client/elastic_search"
)

func main() {
	app := kingpin.New("dataloader", "load data to target systems")

	configFile := app.Flag("config", "data loader config file").Default("config/config.yml").String()

	kingpin.Version("1.0.0")
	kingpin.MustParse(app.Parse(os.Args[1:]))

	bp, err := blueprint.NewBluePrintFromFile(*configFile)

	if err != nil {
		fmt.Printf("failed to read configuration file %+v", err)
		os.Exit(1)
	}

	event := make(chan string, 10000)
	defer close(event)

	var wg sync.WaitGroup

	dest, err := sink.NewSink(&bp.Sink, event, &wg)
	src, err := source.NewSource(&bp.Source, event, &wg)
	cli, err := client.NewClient(&bp.Client, &wg)

	go dest.Process()

	for i := 0; i < 10; i++ {
		go src.Process()
	}
	// For avoiding index is not created issue
	time.Sleep(5 * time.Second)
	go cli.Process()
	wg.Wait()
}
