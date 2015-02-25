package shell

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"

	"github.com/gitbao/gitbao/model"
	"golang.org/x/crypto/ssh"
)

type shellScript struct {
	body string
	kind string
}

func (s *shellScript) addDependencies() {
	s.body += `
touch .profile
export DEBIAN_FRONTEND=noninteractive
sudo apt-get update
sudo apt-get -y install golang
sudo apt-get -y install git`
	if s.kind == "xiaolong" {
		s.body += "\nsudo apt-get -y install docker.io\n"
	} else {
		s.body += "\nsudo apt-get -y install nginx\n"
	}
}

func (s *shellScript) addEnvVariables() {
	localKeysToAdd := []string{
		"GITBAO_DBNAME",
		"GITBAO_DBUSERNAME",
		"GITBAO_DBHOST",
		"GITBAO_DBPASSWORD",
		"GITHUB_GIST_ACCESS_KEY",
	}
	for _, value := range localKeysToAdd {
		s.body += getLocalEnvVarForScript(value)
	}
	s.body += genEnvVarString("GO_ENV", "production")
	s.body += genEnvVarString("GOPATH", "$HOME/golang/")
}

func (s *shellScript) setupGitbao() {
	s.addEnvVariables()
	s.body += `
source .profile
mkdir -p golang
go get github.com/gitbao/gitbao
cd /home/ubuntu/golang/src/
go get ./...
`
	if s.kind != "xiaolong" {
		s.body += "sudo rm /etc/nginx/sites-enabled/default\n"
	}
}

func (s *shellScript) prepareForUpdate() {
	s.body += `
source .profile
rm -rf /home/ubuntu/golang/src/github.com/gitbao
go get github.com/gitbao/gitbao
cd /home/ubuntu/golang/src/
`
}

func (s *shellScript) initServer() {
	s.body += `
cd /home/ubuntu/
source .profile
`
	switch s.kind {

	case "kitchen":
		s.body += `
go install github.com/gitbao/gitbao/cmd/kitchen
cd /home/ubuntu/golang/src/github.com/gitbao/gitbao/cmd/kitchen/
wget https://raw.githubusercontent.com/gitbao/chushi/master/nginx/kitchen
sudo mv kitchen /etc/nginx/sites-enabled/
touch server.log
/home/ubuntu/golang/bin/kitchen > server.log 2>&1 &
sudo service nginx restart
`
	case "router":
		s.body += `
go install github.com/gitbao/gitbao/cmd/router
cd /home/ubuntu/golang/src/github.com/gitbao/gitbao/cmd/router/
wget https://raw.githubusercontent.com/gitbao/chushi/master/nginx/router
sudo mv router /etc/nginx/sites-enabled/
touch server.log
/home/ubuntu/golang/bin/router > server.log 2>&1 &
sudo service nginx restart
`
	case "xiaolong":
		s.body += `
go install github.com/gitbao/gitbao/cmd/xiaolong
cd /home/ubuntu/golang/src/github.com/gitbao/gitbao/cmd/xiaolong/
touch server.log
/home/ubuntu/golang/bin/xiaolong > server.log 2>&1 &
`

	}
}

var homePath string
var chushiRoot string

func init() {
	goPath := os.Getenv("GOPATH")
	chushiRoot = goPath + "src/github.com/gitbao/chushi/"
	usr, _ := user.Current()
	homePath = usr.HomeDir
}

func getLocalEnvVarForScript(key string) string {
	value := os.Getenv(key)
	return genEnvVarString(key, value)
}

func genEnvVarString(key, value string) string {
	return "echo \"export " + key + "=" + value + "\" >> .profile\n"
}

func sshConnect(ip string) (session *ssh.Session, err error) {
	privateBytes, err := ioutil.ReadFile(homePath + "/dev.pem")
	if err != nil {
		panic("Failed to load private key")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("Failed to parse private key")
	}

	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{ssh.PublicKeys(private)},
	}

	client, err := ssh.Dial("tcp", ip+":22", config)
	if err != nil {
		panic(err)
	}

	return client.NewSession()
}

func Initialize(kind string, server *model.Server) error {
	var err error
	var script shellScript

	script.kind = kind

	if kind == "xiaolong" {
		stringId := strconv.Itoa(int(server.Id))
		if err != nil {
			panic(err)
		}
		script.body = genEnvVarString("SERVER_ID", stringId)
	}

	script.addDependencies()
	script.setupGitbao()
	script.initServer()

	fmt.Println(script.body)
	session, err := sshConnect(server.Ip)
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Run(script.body)
	if err != nil {
		log.Fatal(err)
	}
	session.Close()

	if err != nil {
		return err
	}

	return nil
}

func Update(server *model.Server, hard bool) error {
	var err error
	var script shellScript

	if hard == true {
		script.body += "\n sudo rm -rf /home/ubuntu/golang/src/github.com/gitbao/gitbao"
	}
	script.kind = server.Kind
	script.body += "killall " + server.Kind + "\n echo \"process killed\""

	script.prepareForUpdate()
	script.initServer()

	session, err := sshConnect(server.Ip)
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Run(script.body)
	if err != nil {
		log.Fatal(err)
	}
	session.Close()

	return nil
}

func Logs(server *model.Server) {
	var err error

	session, err := sshConnect(server.Ip)
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Run("cd /home/ubuntu/golang/" +
		"src/github.com/gitbao/gitbao/cmd/" + server.Kind + "/ && tail -f server.log")
	if err != nil {
		log.Fatal(err)
	}
	session.Close()

	return
}

func Ssh(server *model.Server) {
	home := os.Getenv("HOME")
	cmd := exec.Command("ssh", "-i", home+"/dev.pem", "ubuntu@"+server.Ip)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
