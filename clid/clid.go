// Package clid implements a prefixed ulid ID type.
package clid

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/oklog/ulid/v2"
)

const separator = "-"

// ID implements a prefixed ULID identifier.
type ID struct {
	p string
	d ulid.ULID
}

// New with default tme and entropy sources and panic when it fails.
func New(prefix string) (id ID) {
	id, err := NewFromParts(prefix, ulid.Now(), ulid.DefaultEntropy())
	if err != nil {
		panic("clid: " + err.Error())
	}

	return
}

// NewFromParts creates an id from its parts.
func NewFromParts(prefix string, ms uint64, entr io.Reader) (id ID, err error) {
	id.p = prefix

	id.d, err = ulid.New(ms, entr)
	if err != nil {
		return id, fmt.Errorf("unable to unit ulid: %w", err)
	}

	return id, nil
}

// String implements the fmt.Stringer interface.
func (id ID) String() string {
	if id.p == "" {
		return fmt.Sprintf("zzz%s%s", separator, id.d) // for zero value stringing
	}

	return fmt.Sprintf("%s%s%s", id.p, separator, id.d)
}

// Value implements the driver Valuer interface.
func (id ID) Value() (driver.Value, error) {
	return id.String(), nil
}

// ScanError describes a failure to scan an ID.
type ScanError struct {
	v any
	m string
}

func (e ScanError) Error() string {
	return fmt.Sprintf("clid: failed to scan %T: %s", e.v, e.m)
}

// Scan implements the sql.Scanner.
func (id *ID) Scan(v any) error {
	var (
		after string
		found bool
		err   error
	)

	switch v := v.(type) {
	case string:
		*id = ID{}

		id.p, after, found = strings.Cut(v, separator)
		if !found {
			return ScanError{v: v, m: "missing separator '" + separator + "'"}
		}

		id.d, err = ulid.ParseStrict(after)
		if err != nil {
			return fmt.Errorf("clid: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("%w", ScanError{v: v, m: "type not supported"})
	}
}

// MarshalJSON implements the json.Marshaler interface.
func (id ID) MarshalJSON() (data []byte, err error) {
	return json.Marshal(id.String()) //nolint:wrapcheck
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (id *ID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("clid: failed to unmarshal as string: %w", err)
	}

	return id.Scan(s)
}

// MarshalGQL marshals the identifier.
func (id ID) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(strconv.Quote(id.String())))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface.
func (id *ID) UnmarshalGQL(v interface{}) error {
	return id.Scan(v)
}
