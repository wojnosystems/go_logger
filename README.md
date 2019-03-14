# Overview

go_logger is a service log that spawns an agent go-routine to write and manage the log file. This is intended to be used in web or other services that cannot block waiting for logging I/O. Logs can be a file or can be network/remote. The object simply needs to conform to io.Writer and my ReOpener interfaces.

# Example

```go
package main

import (	
     "github.com/wojnosystems/go_logger"
     "github.com/wojnosystems/go_reopen"
     "os"
 )

func main() {
	f, err := go_reopen.OpenFile("/var/log/my.log", os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0600)
    if err != nil {
        panic(err)	 
	}
	log := go_logger.NewServiceAgentDefaults(f, "my-service", 10)
	
	log.Log("ERROR", go_logger.NewBase(`this is my log message`), 0)
	
	// perform a log rotation
	_ = log.ReOpen()
	// log file rotation is asynchronous. It's possible that a 
	// log write that occurs after this line will appear in the 
	// old log and not the new log
	
	log.Log("INFO", go_logger.NewBase(`this is a message with extra data`).StreamData("extra_data", "my data. Accepts interface{}, so you can pass whatever you want into this"), 0)
	
	// Shutdown the logger. This will wait for the logs to finish writing
	err = log.Close()
	if err != nil {
		// some error writing to the logs or closing
		panic(err)
	}
}
```

# Operation

When a NewServiceAgent is created, a go-routine is spawned that will perform all of the log writes and file re-opens. Calls to "Log" are guaranteed to never block and thus, writing logs will be fast. Writing to the underlying go_reopen.WriterCloser is always done asynchronously to the caller of Log.

Log messages are sent via a channel to the writer agent. This is called the Backlog (more on this later).

## Custom Messages

There is an interface-type that allows you to specify your own log messages. All log messages are converted to JSON before being written to the log stream.

## Custom Serializer

You can override the serialization strategy (json by default) by specifying a serializationFactory. This method should return an encoder that writes to the provided writer when passed the underlying Serializer object. The serializer type allows implementers to add in custom state to the serializer, if desired.

# The Backlog

Because I'm using channels to communicate with an arbitrary and unknowable number of go-routines with the primary log agent routine, it's possible for this channel to over flow. Because I intend this library to be used in services, rather than blocking, which would be hazardous to a large system, the logger instead lets 1 more message through on a separate channel. The message is a combined message that includes the actual log message as well as a shorter message indicating that the log has overflowed. This is intended to allow ops teams to know when resources are starting to impact visibility.

# Copyright

Copyright © 2019 Chris Wojno. All rights reserved.

No Warranties. Use this software at your own risk.

# License

Attribution 4.0 International https://creativecommons.org/licenses/by/4.0/