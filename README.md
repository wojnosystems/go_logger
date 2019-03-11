# Overview

go_logger is a service log that spawns an agent go-routine to write and manage the log file. This is intended to be used in web or other services that cannot block waiting for logging I/O. Logs can be a file or can be network/remote. The object simply needs to conform to io.Writer and my ReOpener interfaces.

# Copyright

Copyright © 2019 Chris Wojno. All rights reserved.

No Warranties. Use this software at your own risk.

# License

Attribution 4.0 International https://creativecommons.org/licenses/by/4.0/