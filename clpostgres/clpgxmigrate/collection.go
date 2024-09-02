package clpgxmigrate

import (
	"maps"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

var DefaulCollection = NewCollection()

type Collection interface {
	Step(version int64) (Step, bool)
	Versions() []int64
	Register(filename string, step Step) error
}

func NewCollection() Collection {
	c := &mapCollection{registered: make(map[int64]Step)}

	return c
}

type mapCollection struct {
	registered map[int64]Step
}

func (c *mapCollection) Step(version int64) (Step, bool) {
	step, ok := c.registered[version]

	return step, ok
}

func (c *mapCollection) Versions() []int64 {
	versions := slices.Collect(maps.Keys(c.registered))
	slices.Sort(versions)

	return versions
}

func (c *mapCollection) Register(filename string, step Step) error {
	base := filepath.Base(filename)

	version, err := NumericFromFilename(base)
	if err != nil {
		return registerError(filename, "failed to get numeric element from filename: %w", err)
	}

	if _, exists := c.registered[version]; exists {
		return registerError(filename, "migration with the same version (%d) was already registered", version)
	}

	c.registered[version] = step

	return nil
}

func NumericFromFilename(filename string) (int64, error) {
	base, _, found := strings.Cut(filename, "_")
	if !found {
		return 0, registerError(filename, "migration filename has no '_' separator")
	}

	nr, err := strconv.ParseInt(base, 10, 64)
	if err != nil {
		return 0, registerError(filename, "failed to parse numeric element from migration filename: %s: %w", base, err)
	}
	if nr < 1 {
		return 0, registerError(filename, "numeric element in filename is not greater then zero")
	}

	return nr, nil
}

func Register(up Step) {
	_, filename, _, _ := runtime.Caller(1) //nolint:dogsled
	if err := DefaulCollection.Register(filename, up); err != nil {
		panic("clpgxmigrate: " + err.Error())
	}
}
