package go_logger

import (
	reopen "github.com/wojnosystems/go_reopen"
	"io"
)

// ServiceLogger is a logger that can be used with a
// service, where delays writing to log files is not
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
	reopen.ReOpener

	// Close will cause the files to be closed and the
	// go routine to stop
	io.Closer
}