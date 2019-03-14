package go_logger

import (
	"github.com/wojnosystems/go_reopen"
	"github.com/wojnosystems/go_snitch"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestService_Log(t *testing.T) {
	originalName := tmpFileName(t)
	defer func() { _ = os.Remove(originalName) }()
	f, err := go_reopen.OpenFile(originalName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	writeChan := make(chan bool, 1)
	reOpenChan := make(chan bool, 1)
	snitcherFile := go_snitch.NewFile(f, nil, func() {
		writeChan <- true
	}, nil, func() {
		reOpenChan <- true
	})
	sl := NewServiceAgent(snitcherFile, "test-ServiceAgent", time.RFC3339Nano, "\n", 10, defaultNowFactory, defaultSerializerFactory)

	sl.Log("ERROR", NewBase(`log message`), 0)

	// wait for the write
	<-writeChan

	// rotate
	rotatedName := tmpFileName(t)
	defer func() { _ = os.Remove(rotatedName) }()
	err = os.Rename(originalName, rotatedName)
	if err != nil {
		t.Fatal(err)
	}

	err = sl.ReOpen()
	if err != nil {
		t.Error(err)
	}
	// wait for the re-open
	<-reOpenChan

	// ensure log message appeared in rotated file
	m, err := ioutil.ReadFile(rotatedName)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(m), "log message") {
		t.Error(`expected file to contain "log message"`)
	}

	// Log something else
	sl.Log("ERROR", NewBase(`new file started`), 0)

	// wait for the writer
	<-writeChan

	// Shutdown and read log file to ensure log was written
	err = sl.Close()
	if err != nil {
		t.Error(err)
	}

	// Ensure new log written to the new file
	m, err = ioutil.ReadFile(originalName)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(m), "new file started") {
		t.Error(`expected file to contain "new file started"`)
	}
}

func TestServiceAgent_Log_Overflow(t *testing.T) {
	originalName := tmpFileName(t)
	defer func() { _ = os.Remove(originalName) }()
	f, err := go_reopen.OpenFile(originalName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	writeChan := make(chan bool, 1)
	snitcherFile := go_snitch.NewFile(f, nil, func() {
		writeChan <- true
	}, nil, nil)
	sl := NewServiceAgentDefaults(snitcherFile, "test-ServiceAgent", 0)

	// Log something else
	sl.Log("ERROR", NewBase(`new file started`), 0)

	// wait for the writer
	<-writeChan

	// Shutdown and read log file to ensure log was written
	err = sl.Close()
	if err != nil {
		t.Error(err)
	}

	// Ensure new log written to the new file
	m, err := ioutil.ReadFile(originalName)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(m), MsgBacklogFull) {
		t.Errorf(`expected file to contain "%s"`, MsgBacklogFull)
	}
}

// tmpFileName generates a unique file name
func tmpFileName(t *testing.T) string {
	tmpF, err := ioutil.TempFile("", "go-test-logger-*")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmpF.Close()
	return tmpF.Name()
}
