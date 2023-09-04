package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/chrisseto/scwl/pkg"
	"github.com/chrisseto/scwl/pkg/dag"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
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
	// logger := log.Default()

	// Use 23.1.0 to target https://github.com/cockroachdb/cockroach/pull/107633
	// Seed: 1693416869569725000 will produce a reproduction at eed69fee47857c2a3d50b47878180b4a1f198bd6
	sutTS := MustT(testserver.NewTestServer(testserver.CustomVersionOpt("v23.1.0")))
	go func() {
		<-ctx.Done()
		sutTS.Stop()
	}()

	url := sutTS.PGURL()
	url.Path = "system"
	sutDB := MustT(sqlx.Open("pgx", url.String()))
	return pkg.NewSUT(sutDB, logger)
}

func NewOracle(ctx context.Context) pkg.System {
	logger := log.New(NopWriter{}, "", 0)
	// logger := log.Default()

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
		ctx := context.Background()

		state = MustT(oracle.State(ctx))
		sutState := MustT(sut.State(ctx))
		logger.Printf("\tSUT State: %s", sutState.String())
		logger.Printf("\tOracle State: %s", state.String())
	}()

	for i := 0; i < iterations; i++ {
		cmd := pkg.GenerateCommand(state)

		logger.Printf("Step %d: %s", i, pkg.CommandToString(cmd))

		if err := oracle.Execute(ctx, cmd); err != nil {
			panic(err)
		}
		if err := sut.Execute(ctx, cmd); err != nil {
			panic(err)
		}

		state = MustT(oracle.State(ctx))
		sutState := MustT(sut.State(ctx))

		opts := []cmp.Option{
			// cmp.AllowUnexported(dag.Graph{}),
			cmpopts.IgnoreTypes(dag.Node{}),
			// cmpopts.IgnoreUnexported(dag.Node{}),
			cmp.Transformer("Comparable", func(g *dag.Graph) []dag.CNode {
				return g.Comparable()
			}),
		}

		// x := cmp.Transformer("graph", func(g *dag.Graph) []dag.CNode {
		// 	return g.Comparable()
		// 	// return dag.All[dag.INode](g)
		// })

		if diff := cmp.Diff(
			state, sutState, opts...,
		); diff != "" {
			// log.Printf("oracle: %s", MustT(json.MarshalIndent(state.Comparable(), "", "\t")))
			// log.Printf("sut: %s", MustT(json.MarshalIndent(sutState.Comparable(), "", "\t")))
			logger.Printf("\tSUT State: %s", sutState.String())
			logger.Printf("\tOracle State: %s", state.String())
			log.Fatalf("State Mismatch!\n%s", diff)
		}
	}
}
