package main

import (
	"context"
	"flag"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/config"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func TestHighspeed(t *testing.T) {
	flag.Parse()
	start, end := test(t)
	diff := end.UnixMilli() - start.UnixMilli()
	println(diff)
}

func cfg(config2 *config.Config) {
	config2.FetchSize = -1
}

func test(t *testing.T) (time.Time, time.Time) {
	driver, err := neo4j.NewDriverWithContext(
		"bolt://localhost",
		neo4j.BasicAuth("neo4j", "password", ""),
		cfg,
	)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	defer func(driver neo4j.DriverWithContext, ctx context.Context) {
		err := driver.Close(ctx)
		if err != nil {
			panic(err)
		}
	}(driver, ctx)

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		panic(err)
	}
	start := time.Now()

	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j", FetchSize: -1})
	defer func(session neo4j.SessionWithContext, ctx context.Context) {
		err := session.Close(ctx)
		if err != nil {
			panic(err)
		}
	}(session, ctx)

	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		panic(err)
	}
	defer func(tx neo4j.ExplicitTransaction, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			panic(err)
		}
	}(tx, ctx)

	InnerTest(ctx, tx, t)

	end := time.Now()
	return start, end
}

func InnerTest(ctx context.Context, tx neo4j.ExplicitTransaction, t *testing.T) {
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	log.Println("starting test")
	// ... rest of the program ...
	t.Run("running", func(t *testing.T) {
		//for i := 0; i < 100; i++ {
		//	cursor, err := tx.Run(ctx, "UNWIND(RANGE(1, 10000)) AS x RETURN collect(toString(x)) as y", nil)
		//	if err != nil {
		//		panic(err)
		//	}
		//	_, err = cursor.Collect(ctx)
		//	if err != nil {
		//		panic(err)
		//	}
		//}

		//for i := 0; i < 1; i++ {
		cursor, err := tx.Run(ctx, "UNWIND(RANGE(1, 10000)) AS x RETURN collect(x) ", nil)
		if err != nil {
			panic(err)
		}
		_, err = cursor.Collect(ctx)
		if err != nil {
			panic(err)
		}
		//}
	})

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
