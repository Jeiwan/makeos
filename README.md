## makeos (Proof of Concept)

### Usage example
```go
package main

import (
  "github.com/sirupsen/logrus"
  . "github.com/Jeiwan/makeos"
)

func main() {
	WithEnvironment(DevEnvironment, func() {
		tokenAcc := CreateAccount("eosio.token", EOSIO)
		alice := CreateAccount("alice", EOSIO)
		bob := CreateAccount("bob", EOSIO)

		token, err := NewContract("/Users/User/.tmp/eosio.contracts/eosio.token", tokenAcc)
		if err != nil {
			logrus.Fatalln(err)
		}

		if err := token.Build(); err != nil {
			logrus.Fatalln(err)
		}
		if err := token.Deploy(); err != nil {
			logrus.Fatalln(err)
		}

		if err := tokenAcc.PushAction(
			token,
			"create",
			map[string]interface{}{
				"issuer":         tokenAcc.Name(),
				"maximum_supply": "1000000.0000 BTC",
			},
		); err != nil {
			logrus.Fatalln(err)
		}

		if err := tokenAcc.PushAction(
			token,
			"issue",
			map[string]interface{}{
				"to":       alice.Name(),
				"quantity": "100.0000 BTC",
				"memo":     "a gift",
			},
		); err != nil {
			logrus.Fatalln(err)
		}

		if err := alice.PushAction(
			token,
			"transfer",
			map[string]interface{}{
				"from":     alice.Name(),
				"to":       bob.Name(),
				"quantity": "1.0000 BTC",
				"memo":     "a gift",
			},
		); err != nil {
			logrus.Fatalln(err)
		}

		rows, err := token.ReadTable(
			"accounts",
			alice.Name(),
		)
		if err != nil {
			logrus.Fatalln(err)
		}

		logrus.Infof("Alice's balance: %s\n", rows[0]["balance"].(string))

		rows, err = token.ReadTable(
			"accounts",
			bob.Name(),
		)
		if err != nil {
			logrus.Fatalln(err)
		}
		logrus.Infof("Bob's balance: %s\n", rows[0]["balance"].(string))
	})
}
```

### Design ideas
Every time I work with EOS smart contracts I end up with a bash script that initializes my local blockchain and runs several scenarios that usitlize the contract.
I use such script to automate blockchain initialization and to test contracts. It usullay looks like so:
```sh
EOS_BUILD_PATH=~/.tmp/eosio.cdt
KEY=[my public key]
KEOSD_ADDR=http://127.0.0.1:8899

cleos --wallet-url $KEOSD_ADDR set contract ...
cleos --wallet-url $KEOSD_ADDR create account ...
cleos --wallet-url $KEOSD_ADDR push action ...
```

What do I have here? First, there are global options I want to use for every command. In some, more complex, cases I'd also want more flexibility, e.g. I'd want to have more public key (I usulaly use only one for development, to save time).

Then, there are command that do the work. And they're long. It's difficuly to understand what's going on without reading every line entirely. Also, there're a lot of repetitions.

Another drawback of such scripts is that they don't allow to have multiple scenarios and they don't do assertions. Well, of course it's possible to code these things in Bash, but they will look very messy. It's better to have a framework that make such scripts neat and nice.

### Solution?
A tool that allow to code blockchain interaction scenarios and keep them in `git` or elsewhere. This should be very similar to `docker-compose.yml`, `Jenkinsfile`, Kubernetes templates and other such coded configurations. It's also possible that `makeos` would be use for integration testing: scenarios would be able to fire up a node, initialize contracts, do things and assert on data stored in tables and any other data (like account's permissions).

### To YAML or not to YAML?
I really don't know. I hate its syntax quite often, but it's clear and easily readable. YAML scripts for `makeos` could look like so:
```yaml
settings:
  keosd_url: http://127.0.0.1:8899
  public_keys:
    - DEV:[pubkey]
    - USER:[pubkey]
  cdt_build_path: ~/.tmp/eosio.cdt

stages:
- name: initialization
  steps:
  - command: set_contract
    permission: eosio@active
    args:
      account: eosio
      code_path: /eosio.bios # relative to cdt_build_path

  - command: create_account
    permission: eosio@active
    args:
      creator: eosio
      account: eosio.msig
      owner_key: DEV
      active_key: DEV

  - command: set_contract
    permission: eosio.msig@active
    args:
      account: eosio.msig
      code_path: /eosio.msig # relative to cdt_build_path
```