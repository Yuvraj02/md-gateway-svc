module github.com/marketing-digest/gateway

go 1.25.0

require (
	github.com/Yuvraj02/md-protos/proto/auth v0.1.1
	github.com/Yuvraj02/md-protos/proto/blog v0.1.1
	github.com/google/uuid v1.6.0
	github.com/marketing-digest/pkg v0.0.0
	google.golang.org/grpc v1.83.0
)

require (
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

replace github.com/marketing-digest/pkg => ./pkg
