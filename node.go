package makeos

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	eosgo "github.com/eoscanada/eos-go"
)

var nodeos *node
var keos *node

type node struct {
	Client         *eosgo.API
	URL            string
	Wallet         string
	WalletPassword string
}

func newNodeos(URL string) *node {
	return &node{
		Client: eosgo.New(URL),
		URL:    URL,
	}
}

func newKeos(URL, wallet, walletPassword string) *node {
	return &node{
		Client:         eosgo.New(URL),
		URL:            URL,
		Wallet:         wallet,
		WalletPassword: walletPassword,
	}
}

// Cleanup ...
func (n node) Cleanup() {
	pidFile := "./.eosnode.pid"
	tmpPath := fmt.Sprintf("%s/eosnode", os.TempDir())

	pidFileContent, err := ioutil.ReadFile(pidFile)
	if err != nil {
		logrus.Fatalln("read pidfile:", err)
	}

	pid, err := strconv.Atoi(string(pidFileContent))
	if err != nil {
		logrus.Fatalln("parse pid:", err)
	}

	logrus.Infoln("stopping nodeos")
	nodeos, err := os.FindProcess(pid)
	if err != nil {
		logrus.Fatalln("find nodeos:", err)
	}

	if err := nodeos.Signal(syscall.SIGTERM); err != nil {
		logrus.Fatalln("terminate nodeos:", err)
	}

	if err := os.Remove(pidFile); err != nil {
		logrus.Error("remove pidfile:", err)
	}

	if err := os.RemoveAll(tmpPath); err != nil {
		logrus.Fatal("remove tmp:", err)
	}
}

// Restart ...
func (n node) Restart() {
	eosnodeID := "eosnode"
	pidFilename := fmt.Sprintf("./.%s.pid", eosnodeID)
	if _, err := os.Stat(pidFilename); !os.IsNotExist(err) {
		logrus.Fatalln("nodeos is already running")
	}

	tmpPath := fmt.Sprintf("%s/%s", os.TempDir(), eosnodeID)
	if err := os.MkdirAll(tmpPath, os.ModePerm); err != nil {
		logrus.Fatalf("mkdir: %s", err.Error())
	}

	if err := os.MkdirAll(fmt.Sprintf("%s/config", tmpPath), os.ModePerm); err != nil {
		logrus.Fatalf("mkdir: %s", err.Error())
	}

	if err := os.MkdirAll(fmt.Sprintf("%s/data", tmpPath), os.ModePerm); err != nil {
		logrus.Fatalf("mkdir: %s", err.Error())
	}

	configContent, err := Asset("config.ini")
	if err != nil {
		logrus.Fatalf("load config asset: %s", err.Error())
	}
	configSrc := bytes.NewBuffer(configContent)

	configDst, err := os.Create(fmt.Sprintf("%s/config/config.ini", tmpPath))
	if err != nil {
		logrus.Fatalf("create config: %s", err.Error())
	}
	defer configDst.Close()

	_, err = io.Copy(configDst, configSrc)
	if err != nil {
		logrus.Fatalf("copy config: %s", err.Error())
	}

	cmd := exec.Command(
		"nodeos",
		"-e",
		"-p", "eosio",
		"-d", "./data",
		"--config-dir", "./config",
		"--contracts-console",
		"--verbose-http-errors",
		"--delete-all-blocks",
	)
	cmd.Dir = tmpPath
	if err := cmd.Start(); err != nil {
		logrus.Fatalf("nodeos start: %s", err.Error())
	}

	pidFile, err := os.Create(pidFilename)
	if err != nil {
		logrus.Fatalf("create pidfile: %s", err.Error())
	}
	defer pidFile.Close()

	if _, err := io.WriteString(pidFile, strconv.Itoa(cmd.Process.Pid)); err != nil {
		logrus.Fatalf("save pidfile: %s", err.Error())
	}

	logrus.Infoln("nodeos starting")
	for i := 0; i < 5; i++ {
		if _, err := nodeos.Client.GetInfo(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
}
