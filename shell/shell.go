package shell

import (
	"os"
	"os/exec"
	"strings"

	"github.com/gitbao/chushi/model"
)

func init() {

}

func Initialize(kind string, server *model.Server) {
	goPath := os.Getenv("GOPATH")
	chushiRoot := goPath + "src/github.com/gitbao/chushi/"
	homePath := os.Getenv("HOME")

	storeSshFingerprint := exec.Command("ssh-keyscan",
		"-t", "ssh-rsa", server.Ip)
	ipAndFingerprint, err := storeSshFingerprint.Output()

	f, err := os.OpenFile(homePath+"/.ssh/known_hosts", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	if _, err = f.WriteString(string(ipAndFingerprint)); err != nil {
		panic(err)
	}
	f.Close()

	cat := exec.Command(
		"cat",
		chushiRoot+"shell/"+kind+".sh",
	)
	catOut, _ := cat.CombinedOutput()

	cmd := exec.Command(
		"ssh",
		"-i", homePath+"/dev.pem",
		"ubuntu@"+server.Ip,
	)
	cmd.Stdin = strings.NewReader(string(catOut))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// fmt.Printf("%#v", cmd.Args)

	cmd.Run()

	return
}
