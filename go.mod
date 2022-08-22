module github.com/m-lab/go

go 1.18

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
	github.com/m-lab/uuid-annotator v0.4.1
	github.com/prometheus/client_golang v1.7.1
	golang.org/x/net v0.0.0-20200421231249-e086a090c8fd
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.22.0
	gopkg.in/yaml.v2 v2.2.8
)

require (
	cloud.google.com/go v0.56.0 // indirect
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.10.0 // indirect
	github.com/prometheus/procfs v0.1.3 // indirect
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/mod v0.2.0 // indirect
	golang.org/x/sys v0.0.0-20200615200032-f1bc736245b1 // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/tools v0.0.0-20200422205258-72e4a01eba43 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/genproto v0.0.0-20200420144010-e5e8543f8aeb // indirect
	google.golang.org/grpc v1.29.0 // indirect
	google.golang.org/protobuf v1.23.0 // indirect
	honnef.co/go/tools v0.0.1-2020.1.3 // indirect
)
