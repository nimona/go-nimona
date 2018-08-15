package telemetry

import (
	"fmt"
	"sync"
	"time"

	"context"

	"github.com/influxdata/influxdb/client/v2"
)

type InfluxCollector struct {
	client client.Client
	points client.BatchPoints
	input  chan Collectable
	lock   *sync.RWMutex
}

const databaseName = "nimona_metrics"

func NewInfluxCollector(user, pass, addr string) (Collector, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: user,
		Password: pass,
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	input := make(chan Collectable)

	// TODO search for database if not there create it?

	ic := &InfluxCollector{
		client: c,
		lock:   &sync.RWMutex{},
		input:  input,
	}

	go ic.processor(context.Background(), input)

	return ic, nil
}

func (ic *InfluxCollector) Collect(event Collectable) error {
	ic.input <- event
	return nil
}

func (ic *InfluxCollector) processor(ctx context.Context,
	input <-chan Collectable) {
	// Size of the write batch
	batchSize := 1000
	batchConfig := client.BatchPointsConfig{
		Database: databaseName,
	}

	// Create a new point batch
	bp, err := client.NewBatchPoints(batchConfig)
	if err != nil {
		// TODO handle error
		// close channel
	}

	timeout := make(chan bool)

	// Goroutine that send timeout signal every xx ms
	go func(timeout chan bool) {
		for {
			time.Sleep(1000 * time.Millisecond)
			timeout <- true
		}
	}(timeout)

	for {
		select {
		case event := <-input:
			tags := map[string]string{}
			pt, err := client.NewPoint(event.Collection(), tags,
				event.Measurements())
			if err != nil {
				// TODO log error
				fmt.Println(err)
			}
			bp.AddPoint(pt)
		case <-timeout:
			err := ic.client.Write(bp)
			if err != nil {
				// TODO log error
				fmt.Println(err)
			}
			fmt.Println("Wrote points from timeout: ", len(bp.Points()))
			bp, _ = client.NewBatchPoints(batchConfig)
		case <-ctx.Done():
			err := ic.client.Write(bp)
			if err != nil {
				// TODO log error
				fmt.Println(err)
			}
			break
			// TODO dump write
		}

		if len(bp.Points()) >= batchSize {
			err := ic.client.Write(bp)
			if err != nil {
				// TODO log error
				fmt.Println(err)
			}
			fmt.Println("Wrote points from size: ", len(bp.Points()))

			bp, _ = client.NewBatchPoints(batchConfig)
			// TODO check error
		}
	}
}

// collect just adds points to a struct
// a go routine checks if there are
