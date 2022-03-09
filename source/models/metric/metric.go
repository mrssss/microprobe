package metric

import (
	"encoding/json"
	"fmt"
	"github.com/mrssss/microprobe/blueprint"
	"github.com/mrssss/microprobe/source"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Metric struct {
	Devicename          string    `json:"devicename"`
	Region              string    `json:"region"`
	City                string    `json:"city"`
	Version             string    `json:"version"`
	Lat                 float32   `json:"lat"`
	Lon                 float32   `json:"lon"`
	Battery             float32   `json:"battery"`
	Humidity            uint16    `json:"humidity"`
	Temperature         int16     `json:"temperature"`
	HydraulicPressure   float32   `json:"hydraulic_pressure"`
	AtmosphericPressure float32   `json:"atmospheric_pressure"`
	Timestamp           time.Time `json:"timestamp"`
}

var last_alert = time.Now()

func generateMetric(ts time.Time, devIndex int, region string, location LatLon) Metric {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	humidity := uint16(r.Uint32()) % uint16(100)
	if time.Now().Sub(last_alert).Seconds() > 1 {
		humidity = 200
		last_alert = time.Now()
	}

	return Metric{
		Devicename:          fmt.Sprintf("%s-BHSH-%05d", location.City, devIndex),
		Region:              region,
		City:                location.City,
		Version:             "1.0",
		Lat:                 location.Lat,
		Lon:                 location.Lon,
		Battery:             r.Float32() * 100,
		Humidity:            humidity,
		Temperature:         int16(r.Int31()) % int16(100),
		HydraulicPressure:   1000 + r.Float32()*1000,
		AtmosphericPressure: 101.3 + r.Float32()*100,
		Timestamp:           ts,
	}
}

func GenerateMetrics(totalDevices uint32, locations map[string][]LatLon) []Metric {
	ts := time.Now()

	records := make([]Metric, 0, totalDevices)
	for k := range regionMap {
		for i := 0; i < int(totalDevices)/len(regionMap); i++ {
			records = append(records, generateMetric(ts, i, k, locations[k][i]))
		}
	}

	return records
}

type MetricStorage struct {
	event         chan string
	done          *sync.WaitGroup
	running       bool
	totalEntities uint32
}

func (ms *MetricStorage) Process() {
	ms.done.Add(1)
	defer ms.done.Done()
	ms.running = true
	count := 0
	//cur_time := time.Now()
	for {
		locs := GenerateLocations(ms.totalEntities, true)
		mts := GenerateMetrics(ms.totalEntities, locs)
		count += len(mts)
		for _, mt := range mts {
			data, err := json.Marshal(&mt)
			if err != nil {
				log.Println("json marshal failed")
			}
			ms.event <- string(data)
		}
		//if time.Now().Sub(cur_time).Seconds() >= 1 {
		//	cur_time = time.Now()
		//	//log.Printf("Generate %d events.\n", count)
		//}
	}
}

func (ms *MetricStorage) Stop() {
	ms.running = false
}

func init() {
	source.RegisterSource("metric", NewMetricStorage)
}

func NewMetricStorage(sc *blueprint.SourceConfig, event chan string, done *sync.WaitGroup) (source.Source, error) {
	context, ok := sc.Settings.(map[interface{}]interface{})
	var totalEntities uint32 = 10000

	if ok {
		for k, v := range context {
			kk, okk := k.(string)
			vv, okv := v.(int)
			if okk && okv && kk == "total_entities" {
				totalEntities = uint32(vv)
			}
		}
	}
	return &MetricStorage{totalEntities: totalEntities, event: event, done: done, running: true}, nil
}
