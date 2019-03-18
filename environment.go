package makeos

// DevEnvironment ...
var DevEnvironment = &Environment{
	NodeosURL:      "http://127.0.0.1:8888",
	KeosURL:        "http://127.0.0.1:8899",
	Wallet:         "dev",
	WalletPassword: "password",
}

// Environment ...
type Environment struct {
	NodeosURL      string
	KeosURL        string
	Wallet         string
	WalletPassword string
}

// WithEnvironment ...
func WithEnvironment(environment *Environment, body func()) {
	nodeos = newNodeos(environment.NodeosURL)
	keos = newKeos(
		environment.KeosURL,
		environment.Wallet,
		environment.WalletPassword,
	)

	defer nodeos.Cleanup()

	nodeos.Restart()

	body()
}

// Scenario ...
func Scenario(title string, actions func()) {

}
