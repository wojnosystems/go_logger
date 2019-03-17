package go_logger

import (
	"bytes"
	"encoding/json"
	"github.com/wojnosystems/go_reopen"
	"github.com/wojnosystems/go_routine"
	"io"
	"log"
	"runtime"
	"sync"
	"time"
)

const (
	MsgBacklogFull = "ERROR_BACKLOG_FULL"
)

// ServiceAgent is an exported structure to facilitate extension and re-use.
// To create, call NewServiceAgent. Do not instantiate yourself.
type ServiceAgent struct {
	// file, where we write logs to
	file go_reopen.WriteCloser

	// name is what's used when identifying this ServiceAgent in
	// the Log file (in cause multiple systems Log to the
	// same file or you want to aggregate logs later)
	name string

	// timeFormat is how time is rendered in the Log files
	timeFormat string

	// recordSeparator is appended after each Log message
	// in the file
	recordSeparator string

	// loggerRoutine is the routine controller that runs
	// the logging agent
	loggerRoutine go_routine.StopJoiner

	// logBuffers is a list of buffers to be recycled
	// between logs. After every write, the buffer is
	// returned so that other goroutines can use them
	// instead of allocating memory
	logBuffers sync.Pool

	// logMessages is the channel containing buffers
	// to write to the logs
	logMessages chan *bytes.Buffer

	// reOpenChan is the signal to our loggerRoutine
	// that it needs to close and re-open the Log file.
	reOpenChan chan bool

	// timeNowFactory generates the current time
	timeNowFactory func() time.Time

	// logOverflowChan is a smaller channel that is prioritized
	// to notify the user of the library that the log has filled up
	logOverflowChan chan *bytes.Buffer

	// serializerFactory is how users of this library can specify a
	// different log format for their messages. The default is json.Encoder
	serializerFactory SerializerFactory
}

// newService creates a ServiceAgent for testing
func newService(f go_reopen.WriteCloser,
	name,
	timeFormat,
	recordSeparator string,
	maxOutstandingLogMessages int,
	timeNowFactory func() time.Time,
	serializerFactory SerializerFactory) *ServiceAgent {
	return &ServiceAgent{
		file:            f,
		name:            name,
		timeFormat:      timeFormat,
		recordSeparator: recordSeparator,
		logBuffers: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		logMessages:       make(chan *bytes.Buffer, maxOutstandingLogMessages),
		reOpenChan:        make(chan bool, 1),
		timeNowFactory:    timeNowFactory,
		logOverflowChan:   make(chan *bytes.Buffer, 1),
		serializerFactory: serializerFactory,
	}
}

// NewServiceAgent creates a new logger intended for use with services
//
// @param f the file to use to write logs into
// @param name the name of this ServiceAgent. This will be included
//   in every message as the `name` field
// @param timeFormat is the format that will be written to the logs.
//   This is when the log was generated See "time" package for a list
//   of valid formats
// @param recordSeparator is appended to each log message. This is
//   usually a new-line
// @param maxOutstandingLogMessage is the number of buffers (not the
//   size of the buffers) to allocate to the channel. Once this many
//   items is in the channel buffer, log messages will drop. If your
//   routines will produce, at max, 4 log messages, and you cap the
//   number of routines to 10,000, this can be set to 40,000 and
//   operate safely in most circumstances. If this value is exceeded,
//   you'll see log messages with tag: "ERROR_BACKLOG_FULL"
// @param timeNowFactory is a function that generates the current time.
//   This is provided so you can localize time, if you so desire
// @param serializerFactory override the log message format. if set to nil
//   Msg objects are converted to json-strings.
// @return a ServiceAgent logger ready for log messages
func NewServiceAgent(f go_reopen.WriteCloser,
	name,
	timeFormat,
	recordSeparator string,
	maxOutstandingLogMessages int,
	nowFactory func() time.Time,
	serializerFactory SerializerFactory) ServiceLogger {
	s := newService(f, name, timeFormat, recordSeparator, maxOutstandingLogMessages, nowFactory, serializerFactory)
	s.loggerRoutine = go_routine.Go(func(stop <-chan bool) (err error) {
		for {
			select {
			// Log rotation has top priority
			case <-s.reOpenChan:
				err := s.file.ReOpen()
				if err != nil {
					return err
				}

			// Log when there is no more room for log messages in the buffer left
			// this has higher priority than regular messages because we want to
			// try to place these messages as close to the time as possible
			// also, this channel only has 1 slot, so any overflows will be discarded
			// this self-limits this to avoid excessive duplicate messages
			case lm := <-s.logOverflowChan:
				writeBufferToWriter(lm, s.file, s.name)
				// put the buffer back into the pool
				s.logBuffers.Put(lm)

			// Log all messages before closing
			case lm := <-s.logMessages:
				writeBufferToWriter(lm, s.file, s.name)
				// put the buffer back into the pool
				s.logBuffers.Put(lm)

			// stop indicates that this routine needs to stop
			case <-stop:
				return
			}
		}
	})
	return s
}

// writeBufferToWriter actually writes the buffer to the file and handles the error by printing a log message
func writeBufferToWriter(lm *bytes.Buffer, f io.Writer, serviceName string) {
	_, err := f.Write(lm.Bytes())
	if err != nil {
		log.Printf(`unable to write to Log for ServiceAgent: "%s" got error: "%v"`, serviceName, err)
	}
}

// defaultTimeFormat is the intelligent default
var defaultTimeFormat = time.RFC3339Nano

// defaultRecordSeparator is a new-line
var defaultRecordSeparator = "\n"

// defaultNowFactory is time.Now().UTC()
var defaultNowFactory = func() time.Time {
	return time.Now().UTC()
}

// NewServiceAgentDefaults creates a ServiceAgent with some sensible defaults
// defaultTimeFormat: time.RFC3339Nano
// defaultRecordSeparator: "\n"
// defaultNowFactory: time.Now().UTC()
func NewServiceAgentDefaults(f go_reopen.WriteCloser,
	name string,
	maxOutstandingLogMessages int) ServiceLogger {
	return NewServiceAgent(f,
		name,
		defaultTimeFormat,
		defaultRecordSeparator,
		maxOutstandingLogMessages,
		nil,
		nil)
}

var defaultSerializerFactory = func(w io.Writer) Serializer {
	return json.NewEncoder(w)
}

// Log creates a log entry and submits it to the log writing routine
// This method does not block (even if maxOutstandingLogMessages is
// exceeded because the file writer is unable to keep pace with new
// log messages) and will return immediately. If the log buffer fills up
// a special log message will be sent on another channel to record
// in the logs the message that the log buffer has overflowed.
//
// @param tag is the "tag" field on the entry
// @param m is the message to write. You can customize this
//   as needed, but use Base for a convenient default
// @param skip is the number of call stacks to skip. You
//   should pass in zero (0) if you're calling Log directly.
//   If Log is inside another function, pass in 1 to whatever
//   depth is needed. See runtime.Caller for more information
func (s *ServiceAgent) Log(tag string, m Msg, skip int) {
	nowFactory := s.timeNowFactory
	if nowFactory == nil {
		// if no factory set, use the default
		nowFactory = defaultNowFactory
	}
	fl := msgFull{
		Msg:  m,
		Name: s.name,
		Time: nowFactory().Format(s.timeFormat),
		Tag:  tag,
	}

	// Get the backtrace to the caller of Log. Use skip in case the calls are nested
	_, fl.FilePath, fl.Line, _ = runtime.Caller(1 + skip)

	// Get/create a new buffer for messages
	buf := s.logBuffers.Get().(*bytes.Buffer)
	buf.Reset()

	// Take the Msg objects and serialize them
	var enc Serializer
	if s.serializerFactory != nil {
		enc = s.serializerFactory(buf)
	} else {
		enc = defaultSerializerFactory(buf)
	}
	err := enc.Encode(fl)
	if err != nil {
		// error encoding
		log.Printf("error encoding Log entry: %v", err)
		buf.Reset()
		s.logBuffers.Put(buf)
		return
	}

	buf.WriteString(s.recordSeparator)

	// Try to send messages. NEVER block. Always fallback to the default log handler if messages fail to send
	// Be noisy on failure
	select {
	case s.logMessages <- buf:
		// Log message sent, all is well
	default:
		// cannot write the log due to the logger being full
		// We're going to append a second message that will
		// also indicate that the buffer has overflowed
		// Re-using the full message for memory usage/laziness
		fl.Tag = MsgBacklogFull
		fl.Msg = nil
		enc := json.NewEncoder(buf)
		err = enc.Encode(fl)
		if err != nil {
			// error encoding
			log.Printf("error encoding Log entry: %v", err)
			s.logBuffers.Put(buf)
			return
		}
		buf.WriteString(s.recordSeparator)
		select {
		case s.logOverflowChan <- buf:
			// still cannot log the overflow, fallback to the generic logger
			log.Printf(`Error log with ServiceName: "%s" has overflowed`, s.name)
		}
	}

}

// ReOpen causes the log file to be closed and reopened
// @return nil, always.
func (s *ServiceAgent) ReOpen() error {
	s.reOpenChan <- true
	return nil
}

// Close shuts down the logging agent go-routine and closes the file
// @return err the first error encountered in the logger agent routine
func (s *ServiceAgent) Close() (err error) {
	err = s.stopAndJoinError()
	_ = s.file.Close()
	return
}

// stopAndJoinError stops the logging agent go-routine and returns any errors
func (s *ServiceAgent) stopAndJoinError() (err error) {
	return s.loggerRoutine.StopAndJoinError()
}

// Name returns the service agent's name. This is the value of the "name" field in the log messages
func (s *ServiceAgent) Name() string {
	return s.name
}

// TimeFormat returns the currently configured time format, for reference
func (s *ServiceAgent) TimeFormat() string {
	return s.timeFormat
}

// RecordSeparator returns the currently configured record separator, for reference
func (s *ServiceAgent) RecordSeparator() string {
	return s.recordSeparator
}
