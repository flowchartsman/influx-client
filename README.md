# influx-client [![GoDoc](https://godoc.org/github.com/flowchartsman/influx-client?status.svg)](https://godoc.org/github.com/flowchartsman/influx-client)
--
    import "github.com/flowchartsman/influx-client"


## Usage

#### func  Marshal

```go
func Marshal(v interface{}, measurement string) (influx.Point, error)
```
Marshal returns an *influx.Point for v.

Marshal traverses the first level of v. If an encountered value implements the
InfluxValuer or fmt.Stringer interfaces and is not a nil pointer, Marshal will
use the returned value to render the tag or field. Nil pointers are skipped.

Otherwise, Marshal supports encoding integers, floats, strings and booleans.

The encoding of each struct field can be customized by the format string stored
under the "influx" key in the struct field's tag. The format string gives the
name of the field, possibly followed by a comma-separated list of options. The
name may be empty in order to specify options without overriding the default
field name.

The "omitzero" option specifies that the field should be omitted from the
encoding if the field has an zero value as defined by reflect.IsZero() Note:
implemented internally until it lands in tip. (ref:
https://go-review.googlesource.com/c/go/+/171337/ )

The "tag" option specifies that the field is a tag, and the value will be
converted to a string, following InfluxDB specifications. (ref:
https://docs.influxdata.com/influxdb/v1.7/concepts/key_concepts/#tag-value) If
the "tag" option is not present the field will be treated as an InfluxDB field.
(ref:
https://docs.influxdata.com/influxdb/v1.7/concepts/key_concepts/#field-value)

As a special case, if the field tag is "-", the field is always omitted. Note
that a field with name "-" can still be generated using the tag "-,".

Examples of struct field tags and their meanings:

      // Value appears in InfluxDB as field with key "myName".
      Value int `influx:"myName"`

      // Value appears in InfluxDB as tag with key "myName" and stringified
      // integer representation
      Value int `influx:"myname,tag"`

      // Value appears in InfluxDB as field with key "myName" but will be
      // ommitted if it has a zero value as defined above.
      Value int `influx:"myName,omitzero"`

      // Value appears in InfluxDB as field with key "Value" (the default), but
    	 // will be ommitted if it has a zero value.
      Value int `influx:",omitzero"`

      // Value is ignored by this package.
      Value int `influx:"-"`

      // Value appears in InfluxDB with field key "-".
      Value int `influx:"-,"`

Anonymous struct fields will be marshaled with their package-local type name
unless specified otherwise via tags.

Pointer values encode as the value pointed to.

#### type BatchingClient

```go
type BatchingClient struct {
}
```

A BatchingClient batches influx data points and sends them off at a prescribed
interval

#### func  NewBatchingClient

```go
func NewBatchingClient(config BatchingClientConfig) (*BatchingClient, error)
```
NewBatchingClient creates a BatchingClient for influxDB

#### func (*BatchingClient) Send

```go
func (bc *BatchingClient) Send(v interface{}, measurement string) error
```
Send sends a metric to the influx server

#### func (*BatchingClient) Stop

```go
func (bc *BatchingClient) Stop()
```
Stop stops the client and flushes remaining datapoints

#### type BatchingClientConfig

```go
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
	// Logger is the logger to use. Default: nop
	Logger Logger
}
```

BatchingClientConfig is the configuration for a BatchingClient

#### type InfluxValuer

```go
type InfluxValuer interface {
	InfluxValue() (value interface{})
}
```

InfluxValuer is the interface for your type to return a tag or field value

#### type Logger

```go
type Logger interface {
	// Log is a simple logging method your type must have so the influx client can asynchronously log errors
	Log(message string)
}
```

Logger is an interface you can pass to the client where it will log error
messages. NOTE: This will only log error messages
