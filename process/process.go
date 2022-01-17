package process

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	logger "github.com/ipfs/go-log/v2"
)

var (
	log = logger.Logger("tools")
)

// PreparePIDFile 如果workdir和fileName为空字符串，则会在用户的目录下创建名为app.pid的进程号文件
func PreparePIDFile(workdir, fileName string) (pidFilePath string, err error) {
	homeDir := workdir
	if homeDir == "" {
		homeDir, err = os.UserHomeDir()
		if err != nil {
			log.Errorf("读取用户目录失败: %v", err)
			return
		}
	}

	pidFile := fileName
	if pidFile == "" {
		pidFile = "app.pid"
	}

	if _, err = os.Stat(homeDir); os.IsNotExist(err) {
		if err = os.MkdirAll(homeDir, 0755); err != nil {
			return "", fmt.Errorf("创建pid工作目录失败: %v", err)
		}
	} else if err != nil {
		return "", fmt.Errorf("检视pid工作目录失败: %v", err)
	}

	pidFilePath = filepath.Join(homeDir, pidFile)
	return pidFilePath, nil
}

func NewSingleProcess(pidPath string) *SingleProcess {
	return &SingleProcess{pidPath: pidPath}
}

type SingleProcess struct {
	pidPath string
}

func (s *SingleProcess) IsRunning() (bool, error) {
	exist, _, err := s.searchProcess()
	return exist, err
}

func (s *SingleProcess) TouchPID() error {
	pid := fmt.Sprint(os.Getpid())
	pidFile, err := os.Create(s.pidPath)
	if err != nil {
		return fmt.Errorf("创建pid文件失败! %v", err)
	}
	if _, err = pidFile.WriteString(pid); err != nil {
		return fmt.Errorf("写入pid失败！%v", err)
	}

	return pidFile.Close()
}
func (s *SingleProcess) KillUSER1() error {
	return s.Kill(syscall.SIGUSR1)
}

func (s *SingleProcess) KillTerm() error {
	return s.Kill(os.Kill)
}

func (s *SingleProcess) KillInterrupt() error {
	return s.Kill(os.Interrupt)
}

func (s *SingleProcess) Kill(signal os.Signal) error {
	exist, pid, err := s.searchProcess()
	if err != nil {
		return err
	}

	if !exist {
		return nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("查询进程失败！PID: %d, Err: %v", pid, err)
	}
	if err = process.Signal(signal); err != nil {
		return fmt.Errorf("发送进程信号失败: %v, Err: %v", signal, err)
	}

	return s.RemovePIDFile()
}

func (s *SingleProcess) searchProcess() (bool, int, error) {
	if _, err := os.Stat(s.pidPath); os.IsNotExist(err) {
		return false, 0, nil
	}

	pidFile, err := os.Open(s.pidPath)
	if err != nil {
		return false, 0, fmt.Errorf("指定的pid文件不存在:%s, Err: %v", s.pidPath, err)
	}
	defer func() { _ = pidFile.Close() }()

	filePid, err := ioutil.ReadAll(pidFile)
	if err != nil {
		return false, 0, fmt.Errorf("读取pid文件失败: %s, Err: %v", s.pidPath, err)
	}

	pid, err := strconv.Atoi(strings.Trim(fmt.Sprintf("%s", filePid), "\n"))
	if err != nil {
		return false, 0, fmt.Errorf("pid文件中不是有效的进程号: %s, Err: %v", filePid, err)
	}
	// FindProcess 通过其 pid 查找正在运行的进程。
	// 它返回的 Process 可用于获取有关底层操作系统进程的信息。
	// 在 Unix 系统上，无论进程是否存在，FindProcess 总是成功并返回给定 pid 的进程。
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, 0, fmt.Errorf("查询进程失败！PID: %d, Err: %v", pid, err)
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		log.Warnf("指定的进程%d不存在", pid)
		return false, 0, nil
	}

	return true, pid, nil
}

func (s *SingleProcess) RemovePIDFile() error {
	return os.Remove(s.pidPath)
}
