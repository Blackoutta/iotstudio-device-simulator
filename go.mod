module code.iot.chinamobile.com/huyangyi/studio-coap-client

go 1.14

require (
	github.com/dustin/go-coap v0.0.0-20190908170653-752e0f79981e
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9 // indirect
	golang.org/x/sys v0.0.0-20200610111108-226ff32320da // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
)

replace github.com/dustin/go-coap v0.0.0-20190908170653-752e0f79981e => ./go-coap
