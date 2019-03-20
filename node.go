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

	"github.com/Jeiwan/makeos/internal"
	"github.com/sirupsen/logrus"

	eosgo "github.com/eoscanada/eos-go"
)

var nodeos *Node
var keos *Node

// Node represents a nodeos or keos
type Node struct {
	Client         *eosgo.API
	Errors         []error
	URL            string
	Wallet         string
	WalletPassword string
}

func newNodeos(URL string) *Node {
	return &Node{
		Client: eosgo.New(URL),
		Errors: []error{},
		URL:    URL,
	}
}

func newKeos(URL, wallet, walletPassword string) *Node {
	return &Node{
		Client:         eosgo.New(URL),
		Errors:         []error{},
		URL:            URL,
		Wallet:         wallet,
		WalletPassword: walletPassword,
	}
}

// Cleanup ...
func (n Node) Cleanup() {
	pidFile := "./.eosnode.pid"
	tmpPath := fmt.Sprintf("%s/eosnode", os.TempDir())

	if _, err := os.Stat(pidFile); err == os.ErrNotExist {
		logrus.Info("pid file not found, skipping cleanup")
		return
	}

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

// LastError returns last error happened while interacting with the node
func (n Node) LastError() error {
	if len(n.Errors) < 1 {
		return nil
	}

	return n.Errors[0]
}

// PushError appends an error to the list of node's errors
func (n *Node) PushError(err error) {
	n.Errors = append(n.Errors, err)
}

// Start ...
func (n Node) Start() {
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

	configContent, err := internal.Asset("config.ini")
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
