package main

import (
	"flag"
	"log"

	"github.com/kardianos/service"
)

func main() {
	svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()

	config, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	prg := &program{
		Config: config,
	}

	svcConfig := &service.Config{
		Name:        "svc-secure-pdf",
		DisplayName: "Go Service to Encrypt PDFs in a folder",
		Description: "This is service monitors a folder and encrypt all the pdfs that are written to it.",
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	errs := make(chan error, 5)
	logger, err = s.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
