## makeos – EOS automation and contract testing tool (early development)
[![GoDoc](https://godoc.org/github.com/Jeiwan/makeos?status.svg)](https://godoc.org/github.com/Jeiwan/makeos)

### WARNING!
It's still in earliest development stage, APIs are not final, many useful and necessary things are missing, but it works! The name is also not final.

### Usage example
```go
package token_test

import (
	"testing"

	. "github.com/Jeiwan/makeos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// `keosd` must already be running, `makeos` doesn't start it automatically.
func TestIssue(t *testing.T) {
	DevEnvironment.Wallet = "default"
	DevEnvironment.WalletPassword = "password"

	token := NewContract("/path/to/eosio.token")
	token.Build()

	t.Run("ok", func(tt *testing.T) {
		WithEnvironment(DevEnvironment, func(node *Node) {
			tokenAcc := CreateAccount("eosio.token", EOSIO)
			user := CreateAccount("user", EOSIO)

			token.Deploy(tokenAcc)
			require.Nil(tt, node.LastError())

			token.PushAction(
				"create",
				map[string]interface{}{
					"issuer":         tokenAcc.Name,
					"maximum_supply": "1000000.0000 EOS",
				},
				token.Account.Permission("active"),
			)
			require.Nil(tt, node.LastError())

			token.PushAction(
				"issue",
				map[string]interface{}{
					"to":       user.Name,
					"quantity": "10.0000 EOS",
					"memo":     "testing",
				},
				token.Account.Permission("active"),
			)
			require.Nil(tt, node.LastError())

			rows := token.ReadTable("accounts", user.Name)
			require.Nil(tt, node.LastError())
			require.Len(tt, rows, 1)
			assert.Equal(tt, "10.0000 EOS", rows[0]["balance"].(string))

			rows = token.ReadTable("stat", "EOS")
			require.Nil(tt, node.LastError())
			require.Len(tt, rows, 1)
			assert.Equal(tt, "10.0000 EOS", rows[0]["supply"].(string))
		})
	})
}
```