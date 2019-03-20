package makeos

import (
	"strings"

	"github.com/eoscanada/eos-go/system"

	eosgo "github.com/eoscanada/eos-go"
)

// EOSIO is "eosio" account
var EOSIO = NewAccount("eosio")

// Account represents an EOS account
type Account struct {
	Name string
}

// NewAccount returns an Account
func NewAccount(name string) *Account {
	return &Account{
		Name: name,
	}
}

// Permission returns an account permission of specified level
func (a Account) Permission(level string) *Permission {
	return &Permission{
		Actor: a.Name,
		Level: level,
	}
}

// CreateAccount creates an account in the blockchain.
// It uses first public key in the wallet as account's 'owner' and 'active' keys.
func CreateAccount(accName string, owner *Account) *Account {
	if err := keos.Client.WalletUnlock(keos.Wallet, keos.WalletPassword); err != nil {
		if !strings.Contains(err.Error(), "Already unlocked") {
			nodeos.PushError(err)
			return nil
		}
	}

	pubKeys, err := nodeos.Client.Signer.AvailableKeys()
	if err != nil {
		nodeos.PushError(err)
		return nil
	}

	newAccount := system.NewNewAccount(
		eosgo.AccountName(owner.Name),
		eosgo.AccountName(accName),
		pubKeys[0],
	)

	_, err = nodeos.Client.SignPushActions(newAccount)
	if err != nil {
		nodeos.PushError(err)
		return nil
	}

	return NewAccount(accName)
}
