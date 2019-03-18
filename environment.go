package makeos

// DevEnvironment ...
var DevEnvironment = &Environment{}

// Environment ...
type Environment struct {
}

// WithEnvironment ...
func WithEnvironment(environment *Environment, body func()) {
	defer Node.Cleanup()

	Node.Restart()

	body()
}

// Scenario ...
func Scenario(title string, actions func()) {

}
