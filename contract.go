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

// Contract ...
type Contract struct {
	Path    string
	Account *Account
}

// NewContract ...
func NewContract(contractPath string) (*Contract, error) {
	var err error

	if !path.IsAbs(contractPath) {
		contractPath, err = filepath.Abs(contractPath)
		if err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(contractPath); err != nil {
		return nil, err
	}

	return &Contract{
		Path: contractPath,
	}, nil
}

// Build ...
func (c Contract) Build() error {
	command := exec.Command("make")
	command.Dir = c.Path

	out, err := command.CombinedOutput()
	logrus.Println(string(out))

	return err
}

// Deploy ...
func (c *Contract) Deploy(account *Account) error {
	nodeos.Client.SetSigner(eosgo.NewWalletSigner(
		keos.Client,
		keos.Wallet,
	))
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
		eosgo.AccountName(*c.Account),
		fmt.Sprintf("%s/%s", c.Path, wasmName),
	)
	if err != nil {
		return err
	}

	setAbi, err := system.NewSetABI(
		eosgo.AccountName(*c.Account),
		fmt.Sprintf("%s/%s", c.Path, abiName),
	)
	if err != nil {
		return err
	}

	_, err = nodeos.Client.SignPushActions(
		setCode,
		setAbi,
	)

	return err
}

// Name returns contract account's name
func (c Contract) Name() string {
	return c.Account.Name()
}

// PushAction pushes an action to the blockchain
func (c Contract) PushAction(action string, args map[string]interface{}, permission *Permission) error {
	if err := keos.Client.WalletUnlock(keos.Wallet, keos.WalletPassword); err != nil {
		if !strings.Contains(err.Error(), "Already unlocked") {
			return err
		}
	}

	actionData, err := json.Marshal(args)
	if err != nil {
		return err
	}

	abiResp, err := nodeos.Client.GetABI(eosgo.AccountName(c.Name()))
	if err != nil {
		return err
	}

	actionDataHex, err := abiResp.ABI.EncodeAction(
		eosgo.ActionName(action),
		actionData,
	)
	if err != nil {
		return err
	}

	_, err = nodeos.Client.SignPushActions(
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
	)

	return err
}

// ReadTable ...
func (c Contract) ReadTable(table, scope string) ([]map[string]interface{}, error) {
	resp, err := nodeos.Client.GetTableRows(
		eosgo.GetTableRowsRequest{
			Code:  c.Account.Name(),
			Table: table,
			Scope: scope,
			JSON:  true,
		},
	)
	if err != nil {
		return nil, err
	}

	data := []map[string]interface{}{}
	if err := resp.JSONToStructs(&data); err != nil {
		return nil, err
	}

	return data, nil
}
