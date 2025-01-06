package bootstrap

type Bootstrapper interface {
	Run()
	checkBootstrapped() bool
	generateTemplate(string) string
}

