package makeos

import (
	"fmt"

	eosgo "github.com/eoscanada/eos-go"
)

// TokenContract ...
type TokenContract struct {
	Contract
}

// Balance ...
func (c TokenContract) Balance(account, symbol string) *eosgo.Asset {
	balances := c.ReadTable("accounts", account)
	balance, ok := balances[0]["balance"].(string)
	if !ok {
		nodeos.PushError(fmt.Errorf("failed to read balance of %s", account))
		return nil
	}

	asset, err := eosgo.NewAsset(balance)
	if err != nil {
		nodeos.PushError(err)
		return nil
	}

	return &asset
}

// Create calls 'create' action of a token contract.
func (c TokenContract) Create(issuer string, maxSupply string) {
	c.PushAction(
		"create",
		map[string]interface{}{
			"issuer":         issuer,
			"maximum_supply": maxSupply,
		},
		&Permission{
			Actor: c.Account.Name,
			Level: "active",
		},
	)
}

// Issue calls 'issue' action of a token contract.
func (c TokenContract) Issue(to, quantity, memo string) {
	c.PushAction(
		"issue",
		map[string]interface{}{
			"to":       to,
			"quantity": quantity,
			"memo":     memo,
		},
		&Permission{
			Actor: c.Account.Name,
			Level: "active",
		},
	)
}

// Transfer calls 'transfer' action of a token contract.
func (c TokenContract) Transfer(from, to, quantity, memo string) {
	c.PushAction(
		"transfer",
		map[string]interface{}{
			"from":     from,
			"to":       to,
			"quantity": quantity,
			"memo":     memo,
		},
		&Permission{
			Actor: from,
			Level: "active",
		},
	)
}
