package main

import (
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/hhrutter/pdfcpu/pkg/pdfcpu"
	"github.com/kardianos/service"
)

var logger service.Logger

type program struct {
	exit chan struct{}

	*Config
	pdfConfig *pdfcpu.Configuration
}

func (p *program) Start(s service.Service) error {
	logger.Infof("I'm running %v.", service.Platform())
	if service.Interactive() {
		logger.Info("Running in terminal.")
	} else {
		logger.Info("Running under service manager.")
	}

	p.pdfConfig = pdfcpu.NewDefaultConfiguration()
	p.pdfConfig.UserPW = p.Config.Password
	p.pdfConfig.OwnerPW = p.Config.Password
	p.pdfConfig.Mode = pdfcpu.ENCRYPT

	p.exit = make(chan struct{})
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	// Any work in Stop should be quick, usually a few seconds at most.
	logger.Info("Stopping!")
	close(p.exit)
	return nil
}

func (p *program) run() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error(err)
	}

	logger.Info("Watching: ", p.Config.Path)
	logger.Info("Pattern: ", p.Config.Pattern)
	err = watcher.Add(p.Config.Path)
	if err != nil {
		logger.Error(err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				_, name := filepath.Split(event.Name)
				match, _ := filepath.Match(p.Config.Pattern, name)
				if match {
					err = p.encryptPdf(event.Name)
					if err != nil {
						logger.Errorf("%v: %v", name, err)
					}
				}
			}
		case err := <-watcher.Errors:
			logger.Error(err)
		case <-p.exit:
			watcher.Close()
			return nil
		}
	}
}

func (p *program) encryptPdf(path string) error {
	ctx, err := pdfcpu.ReadPDFFile(path, p.pdfConfig)
	if err != nil {
		if err.Error() == "encrypt: This file is already encrypted" {
			return nil // don't care if already encrypted
		}
		return err
	}

	err = pdfcpu.ValidateXRefTable(ctx.XRefTable)
	if err != nil {
		return err
	}

	err = pdfcpu.OptimizeXRefTable(ctx)
	if err != nil {
		return err
	}

	ctx.Write.DirName, ctx.Write.FileName = filepath.Split(path)

	err = pdfcpu.WritePDFFile(ctx)
	if err != nil {
		return err
	}

	if ctx.StatsFileName != "" {
		err = pdfcpu.AppendStatsFile(ctx)
		if err != nil {
			return err
		}
	}

	logger.Infof("Encrypted: %v", ctx.Write.FileName)
	return nil
}
