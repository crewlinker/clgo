// Package clent includes re-usable components for working with the Ent ORM.
package clent

import (
	"fmt"
	"strconv"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"github.com/crewlinker/clgo/clid"
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
			}).SchemaType(map[string]string{
			dialect.Postgres: fmt.Sprintf("varchar(%d)", clid.StorageSize),
		})
	},
}
