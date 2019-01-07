package env

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type localEnv struct {
	workDir string
}

func initLocal(_ map[string]interface{}) (Env, error) {
	wd, err := ioutil.TempDir("/tmp", "dsbench-")
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stderr, wd)

	return &localEnv{
		workDir: wd,
	}, nil
}

func (e *localEnv) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filepath.Join(e.workDir, filename), data, perm)
}

func (e *localEnv) CopyFile(local, filename string, perm os.FileMode) error {
	b, err := ioutil.ReadFile(local)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(e.workDir, filename), b, perm)
}

func (e *localEnv) Close() {
	os.Remove(e.workDir)
}

func (e *localEnv) Cmd(cmd string, args []string, sout io.Writer, serr io.Writer) func() error {
	c := exec.Command(cmd, args...)
	c.Stdout = sout
	c.Stderr = serr
	c.Dir = e.workDir
	return c.Run
}
