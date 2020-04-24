module github.com/byuoitav/av-cli/api

replace github.com/byuoitav/av-cli => ../

go 1.14

require (
	github.com/byuoitav/av-cli v0.0.0-00010101000000-000000000000
	github.com/byuoitav/av-cli/cli v0.0.0-20200423221954-131a0fa73198 // indirect
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.14.1
	google.golang.org/grpc v1.28.1
)
