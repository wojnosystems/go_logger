package go_logger

import "encoding/json"

// Msg interface defines the generic log message type
// These are the bare minimum methods to implement so
// that the log message can be marshaled to the log file
type Msg interface {
	json.Marshaler
}

// Base log supports msg and data fields. This is an
// easy way to bootstrap the log message or extend it
// for other purposes. This supports setting a specific
// message field and a data field that can take arbitrary
// data from the log writer.
type Base struct {
	msg  string
	data map[string]interface{}
}

// baseExport is used to export the data as JSON.
// This is me being lazy and using a struct instead
// of manually writing an encoder.
type baseExport struct {
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

// NewMsg creates a new logger.Base object initialized with
// a msg and with a user-data map ready for use.
// @param msg the message for this log. You can localize this if you wish
func NewMsg(msg string) Msg {
	return &Base{
		msg:  msg,
		data: make(map[string]interface{}),
	}
}

func (b *Base) SetMsg(m string) {
	b.msg = m
}
func (b Base) Msg() string {
	return b.msg
}

func (b *Base) SetData(key string, value interface{}) {
	b.data[key] = value
}
func (b *Base) StreamData(key string, value interface{}) *Base {
	b.SetData(key, value)
	return b
}
func (b *Base) DeleteData(key string) {
	delete(b.data, key)
}
func (b Base) Data() map[string]interface{} {
	return b.data
}
func (b Base) MarshalJSON() (d []byte, err error) {
	be := baseExport{
		Msg:  b.msg,
		Data: b.data,
	}
	return json.Marshal(be)
}

type msgFull struct {
	Msg
	Name     string `json:"name"`
	Tag      string `json:"lvl"`
	Time     string `json:"ts"`
	FilePath string `json:"srcf"`
	Line     int    `json:"srcl"`
}
