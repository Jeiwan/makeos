package makeos

import (
	eosgo "github.com/eoscanada/eos-go"
)

// Environment ...
type Environment struct {
	Autostart      bool
	NodeosURL      string
	KeosURL        string
	Wallet         string
	WalletPassword string
}

// DevEnvironment ...
var DevEnvironment = &Environment{
	Autostart:      true,
	NodeosURL:      "http://127.0.0.1:8888",
	KeosURL:        "http://127.0.0.1:8899",
	Wallet:         "dev",
	WalletPassword: "password",
}

// WithEnvironment ...
func WithEnvironment(environment *Environment, body func()) {
	nodeos = newNodeos(environment.NodeosURL)
	keos = newKeos(
		environment.KeosURL,
		environment.Wallet,
		environment.WalletPassword,
	)
	nodeos.Client.SetSigner(eosgo.NewWalletSigner(
		keos.Client,
		keos.Wallet,
	))

	defer nodeos.Cleanup()

	if environment.Autostart {
		nodeos.Start()
	}

	body()
}

// Scenario ...
func Scenario(title string, actions func()) {

}
