module github.com/byuoitav/av-cli/slack

go 1.14

require (
	github.com/byuoitav/av-cli v0.0.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/slack-go/slack v0.6.3
	github.com/spf13/pflag v1.0.5
	github.com/valyala/fasttemplate v1.1.0 // indirect
	go.uber.org/zap v1.14.1
	google.golang.org/grpc v1.28.1
)

replace github.com/byuoitav/av-cli => ../
