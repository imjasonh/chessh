package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/keygen"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

func runSSHServer(port int) {
	// Generate host key if it doesn't exist
	if err := os.MkdirAll(".ssh", 0700); err != nil {
		log.Fatalln("Failed to create .ssh directory:", err)
	}
	hostKeyPath := ".ssh/chess_host_key"
	if _, err := os.Stat(hostKeyPath); os.IsNotExist(err) {
		log.Println("Generating SSH host key...")
		if _, err := keygen.New(hostKeyPath, keygen.WithKeyType(keygen.Ed25519)); err != nil {
			log.Fatalln("Failed to generate host key:", err)
		}
	}

	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf(":%d", port)),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	log.Printf("Starting SSH chess server on :%d", port)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-ctx.Done()
	log.Println("Stopping SSH server")

	tctx, tcancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer tcancel()
	if err := s.Shutdown(tctx); err != nil {
		log.Fatalln(err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return initialModel(), []tea.ProgramOption{tea.WithAltScreen()}
}
