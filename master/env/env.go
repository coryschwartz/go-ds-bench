package env

import (
	"io"
	"os"
)

type Env interface {
	CopyFile(local, filename string, perm os.FileMode) error
	WriteFile(filename string, data []byte, perm os.FileMode) error

	Cmd(cmd string, args []string, sout io.Writer, serr io.Writer) func() error

	Close()
}

var Handlers = map[string]func(map[string]interface{}) (Env, error){
	"local": initLocal,
	"ssh":   initSsh,
}
