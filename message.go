package go_logger

// Base log supports msg and data fields. This is an
// easy way to bootstrap the log message or extend it
// for other purposes. This supports setting a specific
// message field and a data field that can take arbitrary
// data from the log writer.
type Base struct {
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
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

func (b *Base) SetMsg(m string) {
	b.Msg = m
}
func (b *Base) SetData(key string, value interface{}) {
	b.Data[key] = value
}
func (b *Base) DeleteData(key string) {
	delete(b.Data, key)
}

// msgFull is the actual message object written to the logs.
// This includes everything required to log the messages properly
// with enough context to be useful
type msgFull struct {
	Msg
	Name     string `json:"name"`
	Tag      string `json:"lvl"`
	Time     string `json:"ts"`
	FilePath string `json:"srcf"`
	Line     int    `json:"srcl"`
}
