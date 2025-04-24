package bootstrap

import "context"

type Bootstrapper interface {
	Run(context.Context, string) error
	CheckIfBootstrapped() (bool, error)
	Show(context.Context) error
}

