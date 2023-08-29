package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"time"

	"github.com/chrisseto/scwl/pkg"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const (
	oracleURL = "postgresql://root@localhost:26257/defaultdb?sslmode=disable"
	sutURL    = ""
)

type NopWriter struct{}

func (NopWriter) Write(p []byte) (int, error) { return len(p), nil }

func MustT[T any](r T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
	return r
}

func NewSUT(ctx context.Context) pkg.System {
	logger := log.New(NopWriter{}, "", 0)

	sutTS := MustT(testserver.NewTestServer())
	go func() {
		<-ctx.Done()
		sutTS.Stop()
	}()

	sutDB := MustT(sqlx.Open("pgx", sutTS.PGURL().String()))
	return pkg.NewSUT(sutDB, logger)
}

func NewOracle(ctx context.Context) pkg.System {
	logger := log.New(NopWriter{}, "", 0)

	oracleTS := MustT(testserver.NewTestServer())
	go func() {
		<-ctx.Done()
		oracleTS.Stop()
	}()

	oracleDB := MustT(sqlx.Open("pgx", oracleTS.PGURL().String()))
	return MustT(pkg.NewOracle(oracleDB, logger))
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	sut := NewSUT(ctx)
	oracle := NewOracle(ctx)
	logger := log.Default()

	seed := time.Now().UnixNano()
	rand.Seed(seed)

	iterations := 500

	log.Printf("Iterations: %d, Seed: %d", iterations, seed)

	state := MustT(oracle.State(ctx))

	defer func() {
		state = MustT(oracle.State(ctx))
		sutState := MustT(sut.State(ctx))
		logger.Printf("\tSUT State: %s", MustT(json.MarshalIndent(sutState, "\t\t", "\t")))
		logger.Printf("\tOracle State: %s", MustT(json.MarshalIndent(state, "\t\t", "\t")))
	}()

	for i := 0; i < iterations; i++ {
		cmd := pkg.GenerateCommand(state.(*pkg.Root))

		logger.Printf("Step %d: %#v", i, cmd)

		if err := oracle.Execute(ctx, cmd); err != nil {
			panic(err)
		}
		if err := sut.Execute(ctx, cmd); err != nil {
			panic(err)
		}

		state = MustT(oracle.State(ctx))
		sutState := MustT(sut.State(ctx))

		if !reflect.DeepEqual(state, sutState) {
			panic("State mismatch!")
		}
	}
}
