package shell

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"

	"github.com/gitbao/gitbao/model"
	"golang.org/x/crypto/ssh"
)

type shellScript struct {
	body string
}

func (s *shellScript) addDependencies() {
	s.body += `
export DEBIAN_FRONTEND=noninteractive
sudo apt-get update
sudo apt-get -y install golang
sudo apt-get -y install git
sudo apt-get -y install nginx
`
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
}

func (s *shellScript) initServerByKind(kind string) {
	s.body += `
cd /home/ubuntu/
source .profile
`
	if kind == "kitchen" {
		s.body += `
go install github.com/gitbao/gitbao/cmd/kitchen
cd /home/ubuntu/golang/src/github.com/gitbao/gitbao/cmd/kitchen/
touch server.log
/home/ubuntu/golang/bin/kitchen > server.log 2>&1 &
lsof -wni
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

	script.addDependencies()
	script.setupGitbao()
	script.initServerByKind(kind)

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

func Update(kind string, server *model.Server) error {
	var err error
	var script shellScript

	script.body = "killall " + kind + "\n"
	script.initServerByKind(kind)

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
