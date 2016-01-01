# ESP-01-TCPServer #

This project contains source code of the TCP server used to test ESP8266 TCP client implemented by `TCPClient` example in `ESP01` project.

To cross-compile a version for a different environment than the one you compile on, define `GOOS` and `GOARCH` environment variables to the [right values](https://golang.org/doc/install/source#environment), go into source code directory, and build it:

```
$ go build -v
``` 

Resulting executable file is created in the current directory.