module github.com/cbsinteractive/bakery

replace github.com/zencoder/go-dash => github.com/cbsinteractive/go-dash v0.0.0-20200330181831-e1a70e66546e

replace github.com/grafov/m3u8 => github.com/cbsinteractive/m3u8 v0.11.2-0.20200411022055-4abfe1f82646

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.7 // indirect
	github.com/cbsinteractive/pkg/tracing v0.0.0-20200409233703-f2037b1185c6
	github.com/cbsinteractive/pkg/xrayutil v0.0.0-20200409233703-f2037b1185c6
	github.com/cbsinteractive/propeller-go v0.0.0-20200424170524-41b023ada10e
	github.com/google/go-cmp v0.4.0
	github.com/grafov/m3u8 v0.11.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/zencoder/go-dash v0.0.0-20200221191004-4c1e141085cb
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
)
