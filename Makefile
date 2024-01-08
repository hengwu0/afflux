all:
	GO111MODULE=off GOPATH=`pwd` GOARCH=amd64 go build -ldflags "-s -w" -o afflux Afflux
debug:
	GO111MODULE=off GOPATH=`pwd` GOARCH=amd64 go build -gcflags=all="-N -l" -o afflux Afflux
test:
	GO111MODULE=off GOPATH=`pwd` GOARCH=amd64 go test -count=1 -v Callsys
