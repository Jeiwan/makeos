package makeos

import (
	"encoding/json"
	"strings"

	"github.com/eoscanada/eos-go/system"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/sirupsen/logrus"
)

const localNodeURL = "http://127.0.0.1:8888"
const localKeosdURL = "http://127.0.0.1:8899"

type node struct {
	Client         *eosgo.API
	Wallet         string
	WalletPassword string
}

// Node ...
var Node = node{
	Client: eosgo.New(localNodeURL),
	Wallet: "default",
}

// Keos ...
var Keos = node{
	Client: eosgo.New(localKeosdURL),
	Wallet: "default",
}

// EOSIO ...
var EOSIO = Account("eosio")

// Account ...
type Account string

// Name ...
func (a Account) Name() string {
	return string(a)
}

// PushAction ...
func (a Account) PushAction(contract *Contract, action string, args map[string]interface{}) error {
	Node.Client.SetSigner(eosgo.NewWalletSigner(
		Keos.Client,
		Keos.Wallet,
	))
	if err := Keos.Client.WalletUnlock(Keos.Wallet, Keos.WalletPassword); err != nil {
		if !strings.Contains(err.Error(), "Already unlocked") {
			return err
		}
	}

	actionData, err := json.Marshal(args)
	if err != nil {
		return err
	}

	abiResp, err := Node.Client.GetABI(eosgo.AccountName(contract.Account))
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

	_, err = Node.Client.SignPushActions(
		&eosgo.Action{
			Account: eosgo.AccountName(contract.Account),
			Name:    eosgo.ActionName(action),
			Authorization: []eosgo.PermissionLevel{
				eosgo.PermissionLevel{
					Actor:      eosgo.AccountName(a),
					Permission: eosgo.PermissionName("active"),
				},
			},
			ActionData: eosgo.NewActionDataFromHexData(actionDataHex),
		},
	)

	return err
}

// CreateAccount ...
func CreateAccount(accName string, owner Account) Account {
	Node.Client.SetSigner(eosgo.NewWalletSigner(
		Keos.Client,
		Keos.Wallet,
	))
	if err := Keos.Client.WalletUnlock(Keos.Wallet, Keos.WalletPassword); err != nil {
		if !strings.Contains(err.Error(), "Already unlocked") {
			logrus.Fatalln(err)
		}
	}

	pubKeys, err := Keos.Client.GetPublicKeys()
	if err != nil {
		logrus.Fatalln(err)
	}

	newAccountData := system.NewAccount{
		Creator: eosgo.AccountName(owner.Name()),
		Name:    eosgo.AccountName(accName),
		Owner: eosgo.Authority{
			Threshold: 1,
			Keys: []eosgo.KeyWeight{
				eosgo.KeyWeight{
					PublicKey: *pubKeys[0],
					Weight:    1,
				},
			},
		},
		Active: eosgo.Authority{
			Threshold: 1,
			Keys: []eosgo.KeyWeight{
				eosgo.KeyWeight{
					PublicKey: *pubKeys[0],
					Weight:    1,
				},
			},
		},
	}

	_, err = Node.Client.SignPushActions(
		&eosgo.Action{
			Account: "eosio",
			Name:    "newaccount",
			Authorization: []eosgo.PermissionLevel{
				eosgo.PermissionLevel{
					Actor:      eosgo.AccountName(owner),
					Permission: eosgo.PermissionName("active"),
				},
			},
			ActionData: eosgo.NewActionData(newAccountData),
		},
	)
	if err != nil {
		logrus.Fatalln(err)
	}

	return Account(accName)
}
