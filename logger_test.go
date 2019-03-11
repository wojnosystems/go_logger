package go_logger

import (
	reopen "github.com/wojnosystems/go_reopen"
	snitcher "github.com/wojnosystems/go_snitch"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestService_Log(t *testing.T) {
	tmpF, err := ioutil.TempFile("", "go-test-logger-*")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmpF.Close()
	originalName := tmpF.Name()
	defer func() { _ = os.Remove(originalName) }()
	f, err := reopen.OpenFile(tmpF.Name(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	writeChan := make(chan bool, 1)
	reOpenChan := make(chan bool, 1)
	snitcherFile := snitcher.NewFile(f, nil, func() {
		writeChan <- true
	}, nil, func() {
		reOpenChan <- true
	})
	sl := newService(snitcherFile, "test-service", time.RFC3339Nano, "\n", 10, defaultNowFactory)
	sl.loggerRoutine = defaultLoggerAgent(sl)

	sl.Log("ERROR", NewBase(`log message`), 0)

	// wait for the write
	<-writeChan

	// rotate
	tmpF, err = ioutil.TempFile("", "go-test-logger-*")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmpF.Close()
	defer func() { _ = os.Remove(tmpF.Name()) }()
	err = os.Rename(originalName, tmpF.Name())
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
	oldF, err := os.Open(tmpF.Name())
	if err != nil {
		t.Fatal(err)
	}
	m, err := ioutil.ReadAll(oldF)
	if err != nil {
		t.Fatal(err)
	}
	_ = oldF.Close()

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
	newF, err := os.Open(originalName)
	if err != nil {
		t.Fatal(err)
	}
	m, err = ioutil.ReadAll(newF)
	if err != nil {
		t.Fatal(err)
	}
	_ = newF.Close()

	if !strings.Contains(string(m), "new file started") {
		t.Error(`expected file to contain "new file started"`)
	}
}
