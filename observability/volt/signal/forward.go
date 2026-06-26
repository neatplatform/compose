package signal

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tinylib/msgp/msgp"
)

func runForward(args []string) {
	fs := flag.NewFlagSet("forward", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "Use JSON encoding")

	if err := fs.Parse(args[1:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	if *asJSON {
		fmt.Println("Using JSON encoding")
	} else {
		fmt.Println("Using MessagePack encoding")
	}

	addr := fs.Arg(0)
	if addr == "" {
		fmt.Println("A Forward endpoint is required.")
		os.Exit(1)
	}

	baseKV := []string{
		"name", "volt",
		"env", "local",
		"tenant", "test",
	}

	client, err := newForwardClient("volt.logs", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer client.Close()

	for i, msg := range messages {
		kvs := append(baseKV,
			"uuid", uuid.NewString(),
		)

		if err := client.Send(time.Now(), "info", msg, kvs...); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Sent log:  #%-2d  message=%s\n", i+1, msg)
		}

		time.Sleep(randMs(500, 1000))
	}
}

func init() {
	msgp.RegisterExtension(eventTimeExtType, func() msgp.Extension {
		return new(eventTime)
	})
}

// eventTime is a nanosecond-precision timestamp encoded as msgpack extension type 0.
// See https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1#eventtime-ext-format
type eventTime time.Time

const (
	eventTimeExtType = 0
	eventTimeLen     = 8
)

func (t *eventTime) ExtensionType() int8 {
	return eventTimeExtType
}

func (t *eventTime) Len() int {
	return eventTimeLen
}

func (t *eventTime) MarshalBinaryTo(b []byte) error {
	utc := time.Time(*t).UTC()

	sec := uint32(utc.Unix())
	nsec := uint32(utc.Nanosecond())

	b[0], b[1], b[2], b[3] = byte(sec>>24), byte(sec>>16), byte(sec>>8), byte(sec)
	b[4], b[5], b[6], b[7] = byte(nsec>>24), byte(nsec>>16), byte(nsec>>8), byte(nsec)

	return nil
}

func (t *eventTime) UnmarshalBinary(b []byte) error {
	if len(b) != eventTimeLen {
		return fmt.Errorf("EventTime: invalid length %d", len(b))
	}

	sec := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	nsec := uint32(b[4])<<24 | uint32(b[5])<<16 | uint32(b[6])<<8 | uint32(b[7])

	*t = eventTime(time.Unix(int64(sec), int64(nsec)))

	return nil
}

// record represents a log record.
type record map[string]string

// entry represents a timestamped log record.
type entry struct {
	t eventTime
	r record
}

// forwardClient is a TCP client for sending events/records to a Fluentd Forward Protocol endpoint.
type forwardClient struct {
	mu sync.Mutex

	tag      string
	endpoint string
	conn     net.Conn
}

func newForwardClient(tag, endpoint string) (*forwardClient, error) {
	c := &forwardClient{
		tag:      tag,
		endpoint: endpoint,
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *forwardClient) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := net.DialTimeout("tcp", c.endpoint, 5*time.Second)
	if err != nil {
		return fmt.Errorf("error dialing connection: %w", err)
	}

	c.conn = conn

	return nil
}

func (c *forwardClient) Send(t time.Time, level, message string, kv ...string) error {
	r := make(record, 3+len(kv)/2)

	// Add the fixed fields: timestamp, level, message
	r["timestamp"] = t.Format(time.RFC3339Nano)
	r["level"] = level
	r["message"] = message

	// Add the fields
	for i := 0; i+1 < len(kv); i += 2 {
		k, v := kv[i], kv[i+1]
		r[k] = v
	}

	// No buffering for demoing purposes
	entries := []entry{
		{eventTime(time.Now()), r},
	}

	w := msgp.NewWriter(c.conn)

	// Outer tuple: [tag, entries-array, option-map]
	if err := w.WriteArrayHeader(3); err != nil {
		return err
	}

	if err := w.WriteString(c.tag); err != nil {
		return err
	}

	// Entries: array of [EventTime, record] tuples
	if err := w.WriteArrayHeader(uint32(len(entries))); err != nil {
		return err
	}

	for _, e := range entries {
		if err := w.WriteArrayHeader(2); err != nil {
			return err
		}

		if err := w.WriteExtension(&e.t); err != nil {
			return err
		}

		if err := w.WriteMapStrStr(e.r); err != nil {
			return err
		}
	}

	// No options
	if err := w.WriteMapHeader(0); err != nil {
		return err
	}

	return w.Flush()
}

func (c *forwardClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil

		if err != nil {
			return fmt.Errorf("error closing connection: %w", err)
		}
	}

	return nil
}
