package clpgxmigrate_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/crewlinker/clgo/clpostgres/clpgxmigrate"
	"github.com/jackc/pgx/v5"
	. "github.com/onsi/gomega"
)

func Setup(tb testing.TB) (context.Context, Gomega) {
	tb.Helper()
	g := NewWithT(tb)

	return context.Background(), g
}

func SetupConn(tb testing.TB) (context.Context, Gomega, *pgx.Conn) {
	tb.Helper()
	ctx, g := Setup(tb)

	bootConn, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost:5435/postgres")
	g.Expect(err).ToNot(HaveOccurred())

	var rngb [10]byte
	rand.Read(rngb[:])

	_, err = bootConn.Exec(ctx, fmt.Sprintf("CREATE DATABASE clpgxm_%x TEMPLATE template0", rngb))
	g.Expect(err).ToNot(HaveOccurred())

	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:postgres@localhost:5435/clpgxm_%x", rngb))
	g.Expect(err).ToNot(HaveOccurred())

	tb.Cleanup(func() {
		_, err = bootConn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS clpgxm_%x (FORCE)", rngb))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bootConn.Close(ctx)).To(Succeed())
		g.Expect(conn.Close(ctx)).To(Succeed())
	})

	return ctx, g, conn
}

func SetupProvider(tb testing.TB) (context.Context, Gomega, *clpgxmigrate.Provider) {
	tb.Helper()
	ctx, g, conn := SetupConn(tb)

	return ctx, g, clpgxmigrate.New(conn)
}
