package clcdk

import "strconv"

type conventions struct {
	qualifier  string
	mainRegion string
	account    string
}

// NewConventions inits a convention instance.
func NewConventions(qual, mainRegion, account string) Conventions {
	return conventions{qualifier: qual, mainRegion: mainRegion, account: account}
}

func (c conventions) InstancedStackName(instance int) string {
	return c.Qualifier() +
		strconv.Itoa(instance)
}

func (c conventions) SingletonStackName() string {
	return c.Qualifier()
}

func (c conventions) Qualifier() string {
	return c.qualifier
}

func (c conventions) Account() string {
	return c.account
}

func (c conventions) MainRegion() string {
	return c.mainRegion
}

func (c conventions) EnvSecretName() string {
	return c.Qualifier() + "EnvSecret"
}

// Conventions describes the interface for retrieving info that needs to be consistent between
// the stack and the other programs, i.e: magefiles. Conventions are shared between all stacks,
// instances, accounts and regions.
type Conventions interface {
	InstancedStackName(instance int) string
	SingletonStackName() string
	Qualifier() string
	MainRegion() string
	Account() string
	EnvSecretName() string
}
