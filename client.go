package influx

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	client "github.com/influxdata/influxdb1-client"
	"github.com/pkg/errors"
)

// default logger that does nothing
type noplogger struct{}

// Log satisfies Logger
func (n *noplogger) Log(string) {}

// Logger is an interface you can pass to the client where it will log error messages. NOTE: This will only log error messages
type Logger interface {
	// Log is a simple logging method your type must have so the influx client can asynchronously log errors
	Log(message string)
}

// BatchingClientConfig is the configuration for a BatchingClient
type BatchingClientConfig struct {
	// BatchSize is the maximum size of the batch of points to send. If the buffer is full and another event is added it will flush. Conversely, batches of < BatchSize will be sent if the interval passes.
	// Default: 20
	BatchSize uint
	// Interval is the interval at which to send metrics. Each time it ticks, it will send every event in the buffer.
	// Default: 30 * time.Second
	// Minimum: 1 * time.Second
	Interval time.Duration
	// InfluxURL is the URL to the influx host. Required.
	InfluxURL string
	// Database is the influx database to use. Required.
	Database string
	// Logger is the logger to use. Default: nop
	Logger Logger
}

// A BatchingClient batches influx data points and sends them off at a prescribed interval
type BatchingClient struct {
	// TODO: move most of this crap into an embedded config.
	client    *client.Client
	batchSize int // int for len() comparison
	database  string
	logger    Logger
	stopper   sync.Once
	stopChan  chan struct{}
	pointChan chan client.Point
	flush     *time.Ticker
}

// NewBatchingClient creates a BatchingClient for influxDB
func NewBatchingClient(config BatchingClientConfig) (*BatchingClient, error) {
	// TODO: error check
	if config.BatchSize == 0 {
		config.BatchSize = 20
	}
	if config.Interval < time.Second {
		config.Interval = 30 * time.Second
	}
	if config.Logger == nil {
		config.Logger = new(noplogger)
	}

	influxURL, err := url.Parse(config.InfluxURL)
	if err != nil {
		return nil, errors.Wrap(err, "invalid URL")
	}

	influxClient, err := client.NewClient(client.Config{
		URL:     *influxURL,
		Timeout: client.DefaultTimeout,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error creating influx client")
	}

	bc := &BatchingClient{
		client:    influxClient,
		batchSize: int(config.BatchSize),
		database:  config.Database,
		logger:    config.Logger,
		stopChan:  make(chan struct{}),
		pointChan: make(chan client.Point),
		flush:     time.NewTicker(config.Interval),
	}

	go bc.eventLoop()

	return bc, nil
}

func (bc *BatchingClient) eventLoop() {
	var flush bool
	var stop bool
	var buf []client.Point
	for {
		flush = false
		select {
		case pt := <-bc.pointChan:
			buf = append(buf, pt)
		case <-bc.flush.C:
			flush = true
		case <-bc.stopChan:
			stop = true
		}

		if (flush || stop || len(buf) >= bc.batchSize) && len(buf) > 0 {
			var bp client.BatchPoints
			bp.Points = buf
			bp.Database = bc.database
			_, err := bc.client.Write(bp) // don't care about response for now
			// if there's an error, log it, but do not clear the buffer
			if err != nil {
				bc.logger.Log(err.Error())
			} else {
				buf = nil
			}
		}
		if stop {
			// XXX: drain?
			return
		}
	}
}

// Send sends a metric to the influx server
func (bc *BatchingClient) Send(v interface{}, measurement string) error {
	select {
	case <-bc.stopChan:
		return fmt.Errorf("client has been closed")
	default:
	}
	p, err := Marshal(v, measurement)
	if err != nil {
		return err
	}
	bc.pointChan <- p
	return nil
}

// SendRaw sends a raw influx point to the influx server
func (bc *BatchingClient) SendRaw(rp RawPoint, measurement string) error {
	if len(rp.Fields) == 0 {
		return errors.New("cannot send a point with empty fields")
	}
	var p client.Point

	p.Tags = rp.Tags
	p.Fields = rp.Fields
	p.Time = time.Now()
	p.Measurement = measurement
	bc.pointChan <- p

	return nil
}

// Stop stops the client and flushes remaining datapoints
func (bc *BatchingClient) Stop() {
	bc.stopper.Do(func() {
		close(bc.stopChan)
		// wait?
	})
}

// RawPoint is a raw influx point to be batched and sent directly to the database
type RawPoint struct {
	Tags   map[string]string
	Fields map[string]interface{}
}
