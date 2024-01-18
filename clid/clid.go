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

const (
	// Separator for string encoding, we like '-' for showing them in urls.
	Separator = "-"
	// must be this length, to prevent above through large strings and to make storage size predictable.
	PrefixSize = 4
	// ZeroPrefix is shown when a zero value is encoded, for recognizing that case easily.
	ZeroPrefix = "zzzz"
	// StorageSize determines the size of our identifier when stored, useful for sql schemas.
	StorageSize = 26 + PrefixSize + 1
)

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
	if len(prefix) != PrefixSize {
		panic(fmt.Sprintf("clid: prefix size must be: %d", PrefixSize))
	}

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
		return strings.Join([]string{ZeroPrefix, id.d.String()}, Separator)
	}

	return strings.Join([]string{id.p, id.d.String()}, Separator)
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

		id.p, after, found = strings.Cut(v, Separator)
		if !found {
			return ScanError{v: v, m: "missing separator '" + Separator + "'"}
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
