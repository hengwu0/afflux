all:
	GOPATH=`pwd` GOARCH=amd64 go build -o afflux Afflux
test:
	GOPATH=`pwd` GOARCH=amd64 go test -count=1 -v Callsys
