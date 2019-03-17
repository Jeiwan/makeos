package makeos

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/eoscanada/eos-go/system"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/sirupsen/logrus"
)

// Contract ...
type Contract struct {
	Path    string
	Account Account
}

// NewContract ...
func NewContract(path string, account Account) (*Contract, error) {
	return &Contract{
		Path:    path,
		Account: account,
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
func (c Contract) Deploy() error {
	Node.Client.SetSigner(eosgo.NewWalletSigner(
		Keos.Client,
		Keos.Wallet,
	))
	if err := Keos.Client.WalletUnlock(Keos.Wallet, Keos.WalletPassword); err != nil {
		if !strings.Contains(err.Error(), "Already unlocked") {
			logrus.Fatalln(err)
		}
	}

	pathParts := strings.Split(strings.TrimRight(c.Path, "/"), "/")
	wasmName := fmt.Sprintf("%s.wasm", pathParts[len(pathParts)-1])
	abiName := fmt.Sprintf("%s.abi", pathParts[len(pathParts)-1])

	setCode, err := system.NewSetCode(
		eosgo.AccountName(c.Account),
		fmt.Sprintf("%s/%s", c.Path, wasmName),
	)
	if err != nil {
		return err
	}

	setAbi, err := system.NewSetABI(
		eosgo.AccountName(c.Account),
		fmt.Sprintf("%s/%s", c.Path, abiName),
	)
	if err != nil {
		return err
	}

	_, err = Node.Client.SignPushActions(
		setCode,
		setAbi,
	)

	return err
}

// ReadTable ...
func (c Contract) ReadTable(table, scope string) ([]map[string]interface{}, error) {
	resp, err := Node.Client.GetTableRows(
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
