package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"parserTest/internal/common/awsS3"
	"parserTest/internal/lamoda/config"
	"parserTest/internal/lamoda/parser"
	"parserTest/internal/lamoda/server"

	"github.com/tebeka/selenium"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	cfg := config.LoadConfig()
	awsClient, err := awsS3.New(cfg.Endpoint, cfg.CredS3ID, cfg.CredS3Secret)
	if err != nil {
		log.Fatalln("[SERVICE][AWS] ERROR: ", err)
	}

	// driver, seleniumService, err := webDriver.NewChromeDriver()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer seleniumService.Stop()

	spider := parser.New(cfg,
		[]*awsS3.S3{
			awsClient,
		},
		[]selenium.WebDriver{
			// driver,
		})

	srvr := server.New(spider, cfg)
	ctx, cancel := context.WithCancel(context.Background())

	mainWg := &sync.WaitGroup{}
	mainWg.Add(1)
	go srvr.ParseLamoda(ctx, mainWg)

	<-stop
	cancel()
	mainWg.Wait()
	
	// ctxClose, cancel := context.WithTimeout(context.Background(), time.Second*5)
	// defer cancel()
	// err = srvr.Stop(ctxClose)
	// if err != nil {
	// 	log.Fatalln("[SERVICE][SERVER] STOP ERROR: ", err)
	// }
}
