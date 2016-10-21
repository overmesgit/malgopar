set +e
export GOPATH=$GOPATH:/root/src/github.com/overmesgit/malgopar/
go test -v `ls /root/src/github.com/overmesgit/malgopar/src/ -I main`
go build -v /root/src/github.com/overmesgit/malgopar/src/main/worker.go
go build -v /root/src/github.com/overmesgit/malgopar/src/main/group.go
set -e