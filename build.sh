echo $GOPATH
export GOPATH=$GOPATH:/root/src/github.com/overmesgit/malgopar/src/
go test -v `ls /root/src/github.com/overmesgit/malgopar/src/ -I main`