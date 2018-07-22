package handler

import (
	"net/http"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func ExecShell(path, shellName string) {
	shellPath := path
	if path[len(path)-1] != '/' {
		shellPath += "/"
	}
	shellPath += shellName
	if exist := fileExist(shellPath); !exist {
		log.Error("shell file not exist, ", shellPath)
		return
	}
	cmd := exec.Command("/bin/bash", shellPath)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("exec comment %s err: %s, %s", shellPath, string(output), err)
		return
	}
	log.Infof("exec command %s:%s\n", shellPath, string(output))
}

func fileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func ResponseErr(w http.ResponseWriter) {
	w.WriteHeader(501)
	w.Write([]byte{})
}

// PushInfo push 过来的数据
type PushInfoStruct struct {
	Secret      string
	Branch      string
	ProjectAddr string
}
