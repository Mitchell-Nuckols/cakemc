package main

import (
	"bufio"
	"log"
	"os"
	"os/signal"
	"time"
)

var config Config

func init() {
	if len(os.Args) < 2 {
		log.Fatalln("Must specify a config file: mc-wrapper config.json")
	}

	var err error
	config, err = LoadConfig(os.Args[1])
	if err != nil {
		log.Fatalln("Error loading config file:", os.Args[1])
	}
}

func main() {

	opts := ServerOptions{
		MaxRam:    config.Xmx,
		MinRam:    config.Xms,
		Dir:       config.ServerDir,
		Jar:       config.Jarfile,
		AutoStart: config.AutoRecover,
	}

	bOpts := BackupOptions{
		RootDir:   config.ServerDir,
		BackupDir: config.BackupDir,
		WorldDir:  config.WorldName,
		Interval:  time.Duration(config.BackupInterval) * time.Minute,
		PruneTime: time.Duration(config.PruneAge) * time.Minute,
	}

	s := new(Server)
	s.Start(opts)

	go func() {
		reader := bufio.NewReader(os.Stdin)

		for {
			in, _ := reader.ReadString('\n')
			s.Write(in)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	go func() {
		<-sig
		s.Stop()
	}()

	go backup(s, bOpts)

	s.Wait()
}
