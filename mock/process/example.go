package main

import (
	logger "github.com/ipfs/go-log/v2"
	"tools/process"

	"os"
	"os/signal"
)

var (
	log = logger.Logger("mock")
)

func main() {
	_ = logger.SetLogLevel("*", "debug")
	pidFile, err := process.PreparePIDFile("/Users/alpha/Tmps", "app.pid")
	if err != nil {
		log.Errorf("创建pid文件遇到了问题: %v", err)
		return
	}
	singleProcess := process.NewSingleProcess(pidFile)
	exist, err := singleProcess.IsRunning()
	if err != nil {
		log.Errorf("检查进程遇到了问题: %v", err)
		return
	}

	if exist {
		log.Warnf("当前进程已存在，请稍后再启动")
		if err = singleProcess.KillTerm(); err != nil {
			log.Errorf("杀掉进程失败: %v", err)
			return
		} else {
			log.Infof("进程已杀死")
		}
		os.Exit(0)
	}

	defer func() { _ = singleProcess.RemovePIDFile() }()
	if err = singleProcess.TouchPID(); err != nil {
		log.Errorf("创建进程文件失败: %v", err)
		return
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Kill, os.Interrupt)
	<-ch
}
