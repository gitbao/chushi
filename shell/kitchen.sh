


go install github.com/gitbao/gitbao/cmd/kitchen
cd $GOPATH"src/github.com/gitbao/gitbao/cmd/kitchen"
touch server.log
$GOPATH"bin/kitchen" > server.log 2>&1 &
# exit 0