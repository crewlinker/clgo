package clcdk

import "strconv"

type conventions struct {
	qualifier  string
	mainRegion string
}

// NewConventions inits a convention instance.
func NewConventions(qual, mainRegion string) Conventions {
	return conventions{qualifier: qual, mainRegion: mainRegion}
}

func (c conventions) InstancedStackName(instance int) string {
	return c.Qualifier() +
		strconv.Itoa(instance)
}

func (c conventions) Qualifier() string {
	return c.qualifier
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
	Qualifier() string
	MainRegion() string
	EnvSecretName() string
}
