
generate: interception/log_level_names.go filters/set_names.go interception/shape_hash.pb.go

imports_graph: docs/imports.svg

docs/imports.svg: docs/imports.dot
	dot -Tsvg docs/imports.dot > docs/imports.svg

docs/imports.dot: config/* events/* filters/* interception/* proxy/*
	go mod vendor
	godepgraph -nostdlib -novendor github.com/bearer/go-agent | sed s/splines=ortho// > docs/imports.dot
	rm -fr vendor

interception/shape_hash.pb.go: interception/shape_hash.proto interception/shape_hash.go
	go generate ./...

interception/log_level_names.go: interception/log_level.go
	go generate ./...

filters/set_names.go: filters/set.go
	go generate ./...

lint:
	golint -min_confidence=0.3 ./... && golangci-lint run ./...

test_quick: filters/set_names.go
	go test ./...

test_racy: filters/set_names.go
	go test -v -race ./...
