
filters/set_names.go: filters/set.go
	go generate ./...

test_quick: filters/set_names.go
	go test ./...

test_racy: filters/set_names.go
	go test -race ./...
