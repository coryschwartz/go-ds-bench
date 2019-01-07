package env

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh/knownhosts"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

type sshEnv struct {
	client  *ssh.Client
	workDir string
}

func initSsh(conf map[string]interface{}) (Env, error) {
	key, err := ioutil.ReadFile(conf["KeyFile"].(string))
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
	}

	kh, err := homedir.Expand("~/.ssh/known_hosts")
	if err != nil {
		return nil, err
	}

	hostKeyCb, err := knownhosts.New(kh)
	if err != nil {
		return nil, err
	}

	sshConf := &ssh.ClientConfig{
		User: conf["User"].(string),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCb,
	}

	client, err := ssh.Dial("tcp", conf["Addr"].(string), sshConf)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %v", err)
	}

	s, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	dir, err := s.Output("mktemp -d")
	if err != nil {
		return nil, err
	}
	sdir := strings.TrimSpace(string(dir))

	log.Printf("SSH temp %s", sdir)
	s.Close()

	return &sshEnv{
		client:  client,
		workDir: sdir,
	}, nil
}

func (e *sshEnv) WriteFile(filename string, data []byte, perm os.FileMode) error {
	ftp, err := sftp.NewClient(e.client)
	if err != nil {
		return err
	}
	defer ftp.Close()

	log.Printf("WriteFile %s", filepath.Join(e.workDir, filename))
	f, err := ftp.OpenFile(filepath.Join(e.workDir, filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}

	if err := f.Chmod(perm); err != nil {
		return nil
	}

	return nil
}

func (e *sshEnv) CopyFile(local, filename string, perm os.FileMode) error {
	b, err := ioutil.ReadFile(local)
	if err != nil {
		return err
	}

	return e.WriteFile(filename, b, perm)
}

func (e *sshEnv) Close() {
	defer e.client.Close()

	s, err := e.client.NewSession()
	if err != nil {
		return
	}

	if err := s.Run("rm -rf " + e.workDir); err != nil {
		return
	}

	s.Close()
}

func (e *sshEnv) Cmd(cmd string, args []string, sout io.Writer, serr io.Writer) func() error {
	s, err := e.client.NewSession()
	if err != nil {
		return func() error { return err }
	}
	s.Stdout = sout
	s.Stderr = serr
	if err != nil {
		return func() error { return err }
	}

	return func() error {
		escargs := make([]string, len(args))
		for n := range args {
			escargs[n] = strings.Replace(args[n], " ", "\\ ", -1)
			escargs[n] = strings.Replace(escargs[n], "'", "\\'", -1)
			escargs[n] = strings.Replace(escargs[n], "\"", "\\\"", -1)
		}

		log.Printf("SSH RUN: %s", cmd+" "+strings.Join(escargs, " "))
		defer s.Close()
		return s.Run(fmt.Sprintf("cd '%s' && %s %s", e.workDir, cmd, " " + strings.Join(escargs, " ")))
	}
}
