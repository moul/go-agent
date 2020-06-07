
generate: config/data_collection_rules_names.go filters/set_names.go

config/data_collection_rules_names.go: config/data_collection_rules.go
	go generate ./...

filters/set_names.go: filters/set.go
	go generate ./...

test_quick: filters/set_names.go
	go test ./...

test_racy: filters/set_names.go
	go test -race ./...
