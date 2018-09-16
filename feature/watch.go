package feature

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

/*
	监听文件变化重新执行程序
*/

var (
	processing = false
)

type Watcher struct {
	cmd       *exec.Cmd // exec.Cmd 结构体
	cmdCancel context.CancelFunc
	dir       string // 工作目录
	cmdStr    string // cmd 执行的 shell 命令
}

// StartWater 开启一个新的监控
// workDir 工作目录，go main 包所在的目录
// args 额外参数，这里是指定执行的 shell 命令
func StartWatcher(workDir string, args []string) {
	watcher := Watcher{}
	watcher.init(workDir, args)
}
func (w *Watcher) init(pwdDir string, args []string) {
	dir, err := filepath.Abs(filepath.Dir(pwdDir))
	if err != nil {
		log.Fatal("Get dir err: ", err)
	}
	log.Info("Watching directory: ", dir)
	w.dir = dir
	if len(args) > 0 {
		w.cmdStr = strings.Join(args, " ")
	} else {
		w.cmdStr = "go run *.go"
	}
	w.initCmd()
	w.watch()
}

// initCmd 初始化命令
func (w *Watcher) initCmd() *exec.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", w.cmdStr)
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()
	go printCmdResult(stdout)
	go printCmdResult(stderr)
	if w.cmdCancel != nil {
		w.cmdCancel()
	}
	w.cmdCancel = cancel
	return cmd
}

func (w *Watcher) watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Get watcher failed: ", err)
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}
				go w.execCMD()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Info("watch err: ", err)
			}
		}
	}()
	err = watcher.Add(w.dir)
	if err != nil {
		log.Fatal("watch err: ", err)
	}
	<-done
}

// exec 当文件修改时执行命令
func (w *Watcher) execCMD() {
	// 一个执行后 3s 才能进行新的执行
	if processing {
		return
	}
	processing = true
	go func() {
		time.Sleep(3 * time.Second)
		processing = false
	}()

	// 清屏
	clear()
	log.Infof("Rerun: [%s]", w.cmdStr)
	cmd := w.initCmd()
	cmd.Start()
	cmd.Wait()
}

func printCmdResult(r io.Reader) {
	reader := bufio.NewReader(r)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Debug("printCmdResult err: ", err)
			return
		}
		fmt.Println(string(line))
	}
}

// 清屏
func clear() {
	var clearCMD *exec.Cmd
	if runtime.GOOS == "windows" {
		clearCMD = exec.Command("cmd", "/c", "cls")
	} else {
		clearCMD = exec.Command("clear")
	}
	clearCMD.Stdout = os.Stdout
	clearCMD.Run()
}