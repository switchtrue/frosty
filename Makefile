generate-bin-debug:
	go-bindata -debug -o tmpl/bindata.go -pkg tmpl tmpl

build:
	go-bindata -o tmpl/bindata.go -pkg tmpl tmpl
	go build frosty.go