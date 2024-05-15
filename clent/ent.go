// Package clent includes re-usable components for working with the Ent ORM.
package clent

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"github.com/crewlinker/clgo/clid"
	"github.com/oklog/ulid/v2"
)

// Field is a namespace for re-usable fields.
var Field = struct {
	CLID func(string, string) ent.Field
}{
	// CLID builds an prefixed-ulid ID field with the given name.
	CLID: func(name, prefix string) ent.Field {
		if len(prefix) != clid.PrefixSize {
			panic("clent: clid prefix not of length: " + strconv.Itoa(clid.PrefixSize))
		}

		return field.String(name).
			GoType(clid.ID{}).
			DefaultFunc(func() clid.ID {
				return clid.New(prefix)
			}).
			Validate(Validator.CLID(prefix)).
			SchemaType(map[string]string{
				dialect.Postgres: fmt.Sprintf("varchar(%d)", clid.StorageSize),
				dialect.MySQL:    fmt.Sprintf("varchar(%d)", clid.StorageSize),
			})
	},
}

// Validator name-spaced re-usable (field validators).
var Validator = struct {
	CLID func(string) func(string) error
}{
	// CLIDValidator creates a validator for validating CLID fields.
	CLID: func(expPrefix string) func(s string) error {
		return func(s string) error {
			before, after, found := strings.Cut(s, clid.Separator)
			if !found {
				return errors.New("clent: invalid clid, no separator") //nolint:goerr113
			}

			if _, err := ulid.ParseStrict(after); err != nil {
				return fmt.Errorf("clent: invalid ulid: %w", err)
			}

			if before != expPrefix {
				//nolint:goerr113
				return fmt.Errorf("clent: provided clid prefix '%s' is invalid, expected: '%s'", before, expPrefix)
			}

			return nil
		}
	},
}
