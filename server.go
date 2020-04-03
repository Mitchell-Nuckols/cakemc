package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Server struct {
	stdIn chan string
	cmd   *exec.Cmd
	opts  ServerOptions
}

type ServerOptions struct {
	MaxRam    string
	MinRam    string
	Dir       string
	Jar       string
	AutoStart bool
}

func (s *Server) Start(opts ServerOptions) error {
	s.stdIn = make(chan string)
	s.cmd = exec.Command("java", "-Xmx"+opts.MaxRam, "-Xms"+opts.MinRam, "-jar", opts.Jar, "nogui")
	s.cmd.Dir = opts.Dir
	s.cmd.Stderr = os.Stderr
	s.opts = opts

	s.handleIn()
	s.handleOut()

	if err := s.cmd.Start(); err != nil {
		return err
	}

	return nil
}

func (s *Server) Wait() error {
	for {
		if err := s.cmd.Wait(); err != nil {
			return err
		}

		if exc := s.cmd.ProcessState.ExitCode(); s.opts.AutoStart && exc != 0 {
			log.Println("It looks like the server crashed :( Restarting...")
			s.Start(s.opts)
			continue
		}

		return nil
	}
}

func (s *Server) Stop() error {
	return s.cmd.Process.Signal(syscall.SIGTERM)
}

func (s *Server) Write(input string) {
	s.stdIn <- input
}

func (s *Server) StdoutPipe() (io.ReadCloser, error) {
	return s.cmd.StdoutPipe()
}

func (s *Server) StderrPipe() (io.ReadCloser, error) {
	return s.cmd.StderrPipe()
}

func (s *Server) handleIn() error {
	stdIn, err := s.cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdIn.Close()

		for {
			in := <-s.stdIn
			in = strings.TrimSuffix(in, "\n")
			io.WriteString(stdIn, in+"\r\n")
		}
	}()

	return nil
}

func (s *Server) handleOut() error {

	stdOut, err := s.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdOut.Close()

		reader := bufio.NewReader(stdOut)
		for {
			out, _ := reader.ReadString('\n')
			fmt.Print(out)
		}
	}()

	return nil
}
