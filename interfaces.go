package go_logger

import (
	"github.com/wojnosystems/go_reopen"
	"io"
)

// Msg interface defines the generic log message type
// These are the bare minimum methods to implement so
// that the log message can be JSON marshaled to the log file
type Msg interface {
	// yes, messages can be anything, as long as they can be serialized
}

// MsgBaser is the interface that all Base messages provide
type MsgBaser interface {
	// SetMsg sets a string message for this log message
	SetMsg(m string)

	// SetData allows an arbitrary key-value pair to be set on the log message
	// You can submit arbitrary types as long as they are json.Marshal-able
	SetData(key string, value interface{})

	// DeleteData allows you to remove a key-value object set by SetData
	DeleteData(key string)
}

// Serializer allows you to provide a custom serializerFactory for your log messages
type Serializer interface {
	Encode(interface{}) error
}

// SerializerFactory is how you define your own, custom serializers
type SerializerFactory func(io.Writer) Serializer

// ServiceLogger is a logger that can be used with a
// ServiceAgent, where delays writing to log files is not
// acceptable and the log entry needs to be written
// out of band
type ServiceLogger interface {
	// Send the log message to the logger routine
	// @param tag is a string that tags the message,
	//   like ERROR or whatever you want
	// @param msg is the LogFunc used by the
	//   implementor to create the log message
	// @param skip is the number of stack frames
	//   to skip. 0 is the caller's frame, 1 is
	//   the frame above the caller, and so forth
	Log(tag string, msg Msg, skip int)

	// ReOpen causes the file handle to be closed and opened again
	go_reopen.ReOpener

	// Close will cause the files to be closed and the
	// go routine to stop
	io.Closer
}
