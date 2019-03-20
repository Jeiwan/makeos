package makeos

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/eoscanada/eos-go/system"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/sirupsen/logrus"
)

// Contract holds a connection between a blockchain account and local contract code.
// Implements operations to interact with a contract deployed on the blockchain.
type Contract struct {
	Path    string
	Account *Account
}

// NewContract returns a Contract
func NewContract(contractPath string) *Contract {
	var err error

	if !path.IsAbs(contractPath) {
		contractPath, err = filepath.Abs(contractPath)
		if err != nil {
			nodeos.PushError(err)
			return nil
		}
	}

	if _, err := os.Stat(contractPath); err != nil {
		nodeos.PushError(err)
		return nil
	}

	return &Contract{
		Path: contractPath,
	}
}

// Build compiles a contract. It relies on 'cmake' to provide compilation scripts and expects that Makefile exists in contracts directory.
func (c Contract) Build() {
	command := exec.Command("make")
	command.Dir = c.Path

	out, err := command.CombinedOutput()
	if err != nil {
		nodeos.PushError(err)
	}
	logrus.Info(string(out))
}

// Deploy deploys a contracts on the blockchain and assigns account to the contract.
func (c *Contract) Deploy(account *Account) {
	if err := keos.Client.WalletUnlock(keos.Wallet, keos.WalletPassword); err != nil {
		if !strings.Contains(err.Error(), "Already unlocked") {
			logrus.Fatalln(err)
		}
	}

	pathParts := strings.Split(strings.TrimRight(c.Path, "/"), "/")
	wasmName := fmt.Sprintf("%s.wasm", pathParts[len(pathParts)-1])
	abiName := fmt.Sprintf("%s.abi", pathParts[len(pathParts)-1])

	c.Account = account

	setCode, err := system.NewSetCode(
		eosgo.AccountName(c.Account.Name),
		fmt.Sprintf("%s/%s", c.Path, wasmName),
	)
	if err != nil {
		nodeos.PushError(err)
		return
	}

	setAbi, err := system.NewSetABI(
		eosgo.AccountName(c.Account.Name),
		fmt.Sprintf("%s/%s", c.Path, abiName),
	)
	if err != nil {
		nodeos.PushError(err)
		return
	}

	if _, err = nodeos.Client.SignPushActions(
		setCode,
		setAbi,
	); err != nil {
		nodeos.PushError(err)
	}

	return
}

// Name returns contract account's name
func (c Contract) Name() string {
	return c.Account.Name
}

// PushAction pushes an action to the blockchain
func (c Contract) PushAction(action string, args map[string]interface{}, permission *Permission) {
	if err := keos.Client.WalletUnlock(keos.Wallet, keos.WalletPassword); err != nil {
		if !strings.Contains(err.Error(), "Already unlocked") {
			nodeos.PushError(err)
			return
		}
	}

	actionData, err := json.Marshal(args)
	if err != nil {
		nodeos.PushError(err)
		return
	}

	abiResp, err := nodeos.Client.GetABI(eosgo.AccountName(c.Name()))
	if err != nil {
		nodeos.PushError(err)
		return
	}

	actionDataHex, err := abiResp.ABI.EncodeAction(
		eosgo.ActionName(action),
		actionData,
	)
	if err != nil {
		nodeos.PushError(err)
		return
	}

	if _, err = nodeos.Client.SignPushActions(
		&eosgo.Action{
			Account: eosgo.AccountName(c.Name()),
			Name:    eosgo.ActionName(action),
			Authorization: []eosgo.PermissionLevel{
				eosgo.PermissionLevel{
					Actor:      eosgo.AccountName(permission.Actor),
					Permission: eosgo.PermissionName(permission.Level),
				},
			},
			ActionData: eosgo.NewActionDataFromHexData(actionDataHex),
		},
	); err != nil {
		nodeos.PushError(err)
		return
	}
}

// ReadTable returns rows from a contract's table.
func (c Contract) ReadTable(table, scope string) []map[string]interface{} {
	resp, err := nodeos.Client.GetTableRows(
		eosgo.GetTableRowsRequest{
			Code:  c.Account.Name,
			Table: table,
			Scope: scope,
			JSON:  true,
		},
	)
	if err != nil {
		nodeos.PushError(err)
		return nil
	}

	data := []map[string]interface{}{}
	if err := resp.JSONToStructs(&data); err != nil {
		nodeos.PushError(err)
		return nil
	}

	return data
}
