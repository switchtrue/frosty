generate-bindata-debug:
	go-bindata -debug -o tmpl/bindata.go -pkg tmpl tmpl

build:
	go-bindata -o tmpl/bindata.go -pkg tmpl tmpl
	go build -ldflags "-X github.com/mleonard87/frosty/cli.frostyVersion=0.0.1" frosty.go