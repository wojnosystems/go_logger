package go_logger

// Base log supports msg and data fields. This is an
// easy way to bootstrap the log message or extend it
// for other purposes. This supports setting a specific
// message field and a data field that can take arbitrary
// data from the log writer.
type Base struct {
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// NewBase creates a new logger.Base object initialized with
// a msg and with a user-data map ready for use.
// @param msg the message for this log. You can localize this if you wish
func NewBase(msg string) *Base {
	return &Base{
		Msg:  msg,
		Data: make(map[string]interface{}),
	}
}

// SetMsg changes the Msg field
func (b *Base) SetMsg(m string) {
	b.Msg = m
}

// SetData adds/overwrites a key and value to the data field
func (b *Base) SetData(key string, value interface{}) {
	b.Data[key] = value
}

// DeleteData removes a key-value from the data field
func (b *Base) DeleteData(key string) {
	delete(b.Data, key)
}

// msgFull is the actual message object written to the logs.
// This includes everything required to log the messages properly
// with enough context to be useful
type msgFull struct {
	Msg
	Name     string `json:"n,omitempty"`
	Tag      string `json:"l,omitempty"`
	Time     string `json:"ts,omitempty"`
	FileLine string `json:"src,omitempty"`
}
