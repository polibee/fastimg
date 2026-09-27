package main

import (
	"fmt"
	"os"

	productionservices "goravel/app/services/production"
	"goravel/bootstrap"
)

func main() {
	app := bootstrap.Boot()
	if err := productionservices.Validate(productionservices.Environment()); err != nil {
		fmt.Fprintln(os.Stderr, "FastImg production configuration invalid:", err)
		os.Exit(1)
	}

	app.Start()
}
