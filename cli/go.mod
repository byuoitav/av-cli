module github.com/byuoitav/av-cli/cli

go 1.14

replace github.com/byuoitav/av-cli => ../

require (
	github.com/byuoitav/av-cli v0.0.0-00010101000000-000000000000
	github.com/byuoitav/common v0.0.0-20191210190714-e9b411b3cc0d
	github.com/cheggaaa/pb v2.0.7+incompatible
	github.com/cheggaaa/pb/v3 v3.0.4
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fatih/color v1.9.0
	github.com/manifoldco/promptui v0.7.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.6.3
	golang.org/x/crypto v0.0.0-20200414173820-0848c9571904
	google.golang.org/grpc v1.28.1
)
