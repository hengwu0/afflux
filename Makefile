all:
	GO111MODULE=off GOPATH=`pwd` go build -ldflags "-s -w" -o afflux Afflux
test:
	GO111MODULE=off GOPATH=`pwd` GOARCH=amd64 go test -count=1 -v Callsys
