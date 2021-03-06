package makeos

import (
	eosgo "github.com/eoscanada/eos-go"
)

// Environment ...
type Environment struct {
	Autofail       bool
	Autostart      bool
	NodeosURL      string
	KeosURL        string
	Wallet         string
	WalletPassword string
}

// DevEnvironment ...
var DevEnvironment = &Environment{
	Autofail:       false,
	Autostart:      true,
	NodeosURL:      "http://127.0.0.1:8888",
	KeosURL:        "http://127.0.0.1:8899",
	Wallet:         "dev",
	WalletPassword: "password",
}

// WithEnvironment ...
func WithEnvironment(environment *Environment, body func(*Node)) {
	nodeos = newNodeos(environment.NodeosURL)
	if environment.Autofail {
		nodeos.FailOnError = true
	}
	keos = newKeos(
		environment.KeosURL,
		environment.Wallet,
		environment.WalletPassword,
	)
	nodeos.Client.SetSigner(eosgo.NewWalletSigner(
		keos.Client,
		keos.Wallet,
	))

	defer func() {
		if environment.Autostart {
			nodeos.Cleanup()
		}
	}()

	if environment.Autostart {
		nodeos.Start()
	}

	body(nodeos)
}
