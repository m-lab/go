module github.com/m-lab/go

go 1.16

// These v1 versions were published incorrectly. Retracting to prevent go mod
// from automatically selecting them.
retract [v1.0.0, v1.4.1]

require (
	cloud.google.com/go/bigquery v1.6.0
	cloud.google.com/go/datastore v1.1.0
	cloud.google.com/go/storage v1.6.0
	github.com/araddon/dateparse v0.0.0-20200409225146-d820a6159ab1
	github.com/go-test/deep v1.0.6
	github.com/googleapis/google-cloud-go-testing v0.0.0-20191008195207-8e1d251e947d
	github.com/kabukky/httpscerts v0.0.0-20150320125433-617593d7dcb3
	github.com/kr/pretty v0.2.0
	github.com/kr/text v0.2.0 // indirect
	github.com/m-lab/uuid-annotator v0.4.1
	github.com/prometheus/client_golang v1.7.1
	golang.org/x/net v0.0.0-20200421231249-e086a090c8fd
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/tools v0.0.0-20200422205258-72e4a01eba43 // indirect
	google.golang.org/api v0.22.0
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/genproto v0.0.0-20200420144010-e5e8543f8aeb // indirect
	google.golang.org/grpc v1.29.0 // indirect
	gopkg.in/yaml.v2 v2.2.8
)
