module github.com/cbsinteractive/bakery

replace github.com/zencoder/go-dash => github.com/cbsinteractive/go-dash v0.0.0-20200617014501-54010516d9b0

replace github.com/grafov/m3u8 => github.com/cbsinteractive/m3u8 v0.11.2-0.20200411022055-4abfe1f82646

replace github.com/cbsinteractive/propeller-go => github.com/cbsinteractive/propeller-go v0.0.0-20200828160349-e31e5d2b12de

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.7 // indirect
	github.com/cbsinteractive/pkg/tracing v0.0.0-20200409233703-f2037b1185c6
	github.com/cbsinteractive/pkg/xrayutil v0.0.0-20200409233703-f2037b1185c6
	github.com/cbsinteractive/propeller-go v0.0.0-20200503222720-e53e98ec6b80
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/grafov/m3u8 v0.11.1
	github.com/justinas/alice v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/rs/zerolog v1.18.0
	github.com/zencoder/go-dash v0.0.0-20200221191004-4c1e141085cb
)
