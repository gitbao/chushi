export DEBIAN_FRONTEND=noninteractive
sudo apt-get -y install update
sudo apt-get -y install golang
sudo apt-get -y install git
sudo apt-get -y install nginx
echo "export GOPATH=$HOME/golang/" >> .profile
source .profile
go get github.com/gitbao/gitbao
go get ./...
go install github.com/gitbao/gitbao/cmd/kitchen
./golang/bin/kitchen &