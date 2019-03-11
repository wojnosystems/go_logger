package go_logger

import (
	"bytes"
	"encoding/json"
	reopen "github.com/wojnosystems/go_reopen"
	routine "github.com/wojnosystems/go_routine"
	"log"
	"runtime"
	"sync"
	"time"
)

type service struct {
	// file, where we write logs to
	file reopen.WriteCloser

	// name is what's used when identifying this service in
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
	loggerRoutine routine.StopJoiner

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

	// newFactory generates the current time
	nowFactory func() time.Time
}

// newService creates a service for testing
func newService(f reopen.WriteCloser,
	name,
	timeFormat,
	recordSeparator string,
	maxOutstandingLogMessages int,
	nowFactory func() time.Time) *service {
	return &service{
		file:            f,
		name:            name,
		timeFormat:      timeFormat,
		recordSeparator: recordSeparator,
		logBuffers: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		logMessages: make(chan *bytes.Buffer, maxOutstandingLogMessages),
		reOpenChan:  make(chan bool, 1),
		nowFactory:  nowFactory,
	}
}

// defaultLoggerAgent is the default guts of the logging agent routine
// this allows it to be swapped out for testing
var defaultLoggerAgent = func(s *service) routine.StopJoiner {
	return routine.Go(func(stop <-chan bool) error {
		for {
			select {
			case <-stop:
				return nil
			case <-s.reOpenChan:
				err := s.file.ReOpen()
				if err != nil {
					return err
				}
			case lm := <-s.logMessages:
				_, err := s.file.Write(lm.Bytes())
				if err != nil {
					log.Printf(`unable to write to Log for service: "%s" got error: "%v"`, s.name, err)
				}
				// put the buffer back into the pool
				lm.Reset()
				s.logBuffers.Put(lm)
			}
		}
	})
}

// Service creates a new logger intended for use with services
// @param f the file to use to write logs into
// @param name the name of this service. This will be included in every message as the `name` field
// @param timeFormat is the format that will be written to the logs. This is when the log was generated See "time" package for a list of valid formats
// @param recordSeparator is appended to each log message. This is usually a new-line
// @param maxOutstandingLogMessage is the number of buffers (not the size of the buffers) to allocate to the channel. Once this many items is in the channel buffer, log messages will block. Ensure that this value is large enough to prevent any blocking. If your routines will produce, at max, 4 log messages, and you cap the number of routines to 10,000, this can be set to 40,000 and operate safely in most circumstances.
// @param nowFactory is a function that generates the current time. This is provided so you can localize time, if you so desire
// @return a service logger ready for log messages
func Service(f reopen.WriteCloser,
	name,
	timeFormat,
	recordSeparator string,
	maxOutstandingLogMessages int,
	nowFactory func() time.Time) ServiceLogger {
	s := newService(f, name, timeFormat, recordSeparator, maxOutstandingLogMessages, nowFactory)
	s.loggerRoutine = defaultLoggerAgent(s)
	return s
}

// defaultTimeFormat is the intelligent default
var defaultTimeFormat = time.RFC3339Nano

// defaultRecordSeparator is a new-line
var defaultRecordSeparator = "\n"

// defaultNowFactory is time.Now().UTC()
var defaultNowFactory = func() time.Time {
	return time.Now().UTC()
}

func ServiceDefaults(f reopen.WriteCloser, name string, maxOutstandingLogMessages int) ServiceLogger {
	return Service(f, name, defaultTimeFormat, defaultRecordSeparator, maxOutstandingLogMessages, defaultNowFactory)
}

// Log creates a log entry and submits it to the log writing routine
// This method does not block (unless maxOutstandingLogMessages is
// exceeded because the file writer is unable to keep pace with new
// log messages) and will return immediately.
//
// @param tag is the "tag" field on the entry
// @param m is the message to write. You can customize this
//   as needed, but use Base for a convenient default
// @param skip is the number of call stacks to skip. You
//   should pass in zero (0) if you're calling Log directly.
//   If Log is inside another function, pass in 1 to whatever
//   depth is needed. See runtime.Caller for more information
func (s *service) Log(tag string, m Msg, skip int) {
	fl := msgFull{
		Msg:  m,
		Name: s.name,
		Time: s.nowFactory().Format(s.timeFormat),
		Tag:  tag,
	}

	_, fl.FilePath, fl.Line, _ = runtime.Caller(skip + 1)

	buf := s.logBuffers.Get().(*bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(fl)
	if err != nil {
		// error encoding
		log.Printf("error encoding Log entry: %v", err)
		buf.Reset()
		s.logBuffers.Put(buf)
		return
	}

	buf.WriteString(s.recordSeparator)

	s.logMessages <- buf
}

// ReOpen causes the log file to be closed and reopened
// @return nil, always.
func (s *service) ReOpen() error {
	s.reOpenChan <- true
	return nil
}

// Close shuts down the logging agent go-routine and closes the file
// @return err the first error encountered in the logger agent routine
func (s *service) Close() (err error) {
	err = s.stopAndJoinError()
	_ = s.file.Close()
	return
}

// stopAndJoinError stops the logging agent go-routine and returns any errors
func (s *service) stopAndJoinError() (err error) {
	return s.loggerRoutine.StopAndJoinError()
}
