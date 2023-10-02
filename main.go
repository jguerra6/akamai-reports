package main

import (
	"log"
	"os"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/edgegrid"
	"github.com/jguerra6/akamai-reports/master"
	"github.com/jguerra6/akamai-reports/report"
)

type Runner interface {
	Init([]string) error
	Run() error
	Name() string
}

func initAkamai() *edgegrid.Config {
	edgerc := edgegrid.Must(edgegrid.New(
		edgegrid.WithFile("./config/.edgerc"),
		edgegrid.WithSection("default"),
	))

	return edgerc
}

func main() {

	edgerc := initAkamai()

	masterCommand := master.New(edgerc)
	reportCommand := report.New(edgerc)

	commands := []Runner{
		masterCommand,
		reportCommand,
	}

	subcommand := os.Args[1]

	for _, command := range commands {
		if command.Name() == subcommand {
			err := command.Init(os.Args[2:])
			if err != nil {
				log.Fatalf("failed initing command: %s", err)
			}
			err = command.Run()
			if err != nil {
				log.Fatalf("Error: %s", err)
			}
		}
	}

}
