default:

compile:
		go generate
		go build -ldflags "-X main.version=$(shell git describe --tags || git rev-parse --short HEAD)" .

bindata:
		go-bindata nameservers.yaml
