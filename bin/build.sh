#!/bin/bash
WORKDIR=/home
GOOS=linux go build -o $WORKDIR/bin/studio-coap-linux $WORKDIR/cmd/coap/coap_exec.go
GOOS=windows go build -o $WORKDIR/bin/studio-coap-windows.exe $WORKDIR/cmd/coap/coap_exec.go
GOOS=linux go build -o $WORKDIR/bin/studio-mqtt-linux $WORKDIR/cmd/mqtt/*.go
GOOS=windows go build -o $WORKDIR/bin/studio-mqtt-windows.exe $WORKDIR/cmd/mqtt/*.go