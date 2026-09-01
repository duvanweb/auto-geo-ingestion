package main

import (
	"context"
	"fmt"

	"go.uber.org/fx"
)

// @title           Auto Geo Ingestion API
// @version         1.0
// @description     GPS ingestion microservice with Hexagonal Architecture and Uber FX.
// @host            localhost:8080
// @BasePath        /api
// @schemes         http https
func main() {
	ctx := context.Background()

	app := fx.New(
		Module(),
	)

	if err := app.Start(ctx); err != nil {
		panic(fmt.Errorf("failed to start application: %w", err))
	}

	sig := <-app.Wait()
	fmt.Printf("application stopped with code: %v\n", sig.ExitCode)

	if err := app.Stop(ctx); err != nil {
		fmt.Printf("error stopping application: %v\n", err)
	}
}
