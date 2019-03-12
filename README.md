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

# Copyright

Copyright © 2019 Chris Wojno. All rights reserved.

No Warranties. Use this software at your own risk.

# License

Attribution 4.0 International https://creativecommons.org/licenses/by/4.0/