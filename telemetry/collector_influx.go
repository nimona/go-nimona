package telemetry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"go.uber.org/zap"

	"nimona.io/go/log"
)

// InfluxCollector implements the Collector interface with a InlfuxDB storage
type InfluxCollector struct {
	client      client.Client
	input       chan Collectable
	lock        *sync.RWMutex
	logger      *zap.Logger
	batchSize   int
	timeout     time.Duration
	batchConfig client.BatchPointsConfig
}

const databaseName = "metrics"

// NewInfluxCollector connects to an inlfuxdb, starts listenings for
// Collectables and returns a Collector
func NewInfluxCollector(user, pass, addr string) (Collector, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: user,
		Password: pass,
	})
	if err != nil {
		return nil, err
	}

	input := make(chan Collectable)

	ic := &InfluxCollector{
		client: c,
		lock:   &sync.RWMutex{},
		input:  input,
		logger: log.Logger(context.Background()).Named("collector_influx"),
		// TODO fix the point aggregation and find a sane batch size
		batchSize: 1,
		batchConfig: client.BatchPointsConfig{
			Database:  databaseName,
			Precision: "ns",
		},
		timeout: 1000 * time.Millisecond,
	}

	// Create the database
	if err := ic.createDB(databaseName); err != nil {
		return nil, err
	}

	// Initialize the processor
	go ic.processor(context.Background(), input)

	return ic, nil
}

// Collect send any collectable event to the db
func (ic *InfluxCollector) Collect(event Collectable) error {
	ic.input <- event
	return nil
}

func (ic *InfluxCollector) processor(ctx context.Context,
	input <-chan Collectable) {

	// Create a new point batch
	bp, err := client.NewBatchPoints(ic.batchConfig)
	if err != nil {
		ic.logger.Error("Failed to create new batch points", zap.Error(err))
		return
	}

	timeout := make(chan bool)

	// Goroutine that send timeout signal every xx ms
	go func(timeout chan bool) {
		for {
			time.Sleep(ic.timeout)
			timeout <- true
		}
	}(timeout)

	for {
		select {
		case event := <-input:
			tags := map[string]string{}
			fields := map[string]interface{}{}
			for k, v := range event.Measurements() {
				switch k {
				case "direction", "content_type":
					sv, ok := v.(string)
					if !ok {
						continue
					}
					tags[k] = sv
				default:
					fields[k] = v
				}
			}
			pt, err := client.NewPoint(event.Collection(), tags, fields)
			if err != nil {
				ic.logger.Error("Failed to create point", zap.Error(err))
			}
			bp.AddPoint(pt)
		case <-timeout:
			bp, err = ic.writePoints(bp)
			if err != nil {
				ic.logger.Error("Failed to write points", zap.Error(err))
			}
		case <-ctx.Done():
			bp, err = ic.writePoints(bp)
			if err != nil {
				ic.logger.Error("Failed to write points", zap.Error(err))
			}
			break
		}

		if len(bp.Points()) >= ic.batchSize {
			bp, err = ic.writePoints(bp)
			if err != nil {
				ic.logger.Error("Failed to write points", zap.Error(err))
			}
		}
	}
}

func (ic *InfluxCollector) createDB(db string) error {
	_, err := ic.client.Query(client.Query{
		Command: fmt.Sprintf("CREATE DATABASE %s", db),
	})
	if err != nil {
		ic.logger.Error("Failed to create database", zap.Error(err))
		return err
	}

	return nil
}

func (ic *InfluxCollector) writePoints(bp client.BatchPoints) (client.BatchPoints, error) {
	nbp, _ := client.NewBatchPoints(ic.batchConfig)
	err := ic.client.Write(bp)
	return nbp, err
}
