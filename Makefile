generate-bindata-debug:
	go-bindata -debug -o tmpl/bindata.go -pkg tmpl tmpl

build:
	go-bindata -o tmpl/bindata.go -pkg tmpl tmpl
	env GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/mleonard87/frosty/cli.frostyVersion=0.0.1" -o "build/darwin_amd64" frosty.go
	env GOOS=linux GOARCH=386 go build -ldflags "-X github.com/mleonard87/frosty/cli.frostyVersion=0.0.1" -o "build/linux_386" frosty.go
	env GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/mleonard87/frosty/cli.frostyVersion=0.0.1" -o "build/linux_amd64" frosty.go
	env GOOS=linux GOARCH=arm go build -ldflags "-X github.com/mleonard87/frosty/cli.frostyVersion=0.0.1" -o "build/linux_arm" frosty.go
	env GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/mleonard87/frosty/cli.frostyVersion=0.0.1" -o "build/windows_amd64.exe" frosty.go
	env GOOS=windows GOARCH=386 go build -ldflags "-X github.com/mleonard87/frosty/cli.frostyVersion=0.0.1" -o "build/windows_386.exe" frosty.go