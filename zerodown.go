// Package zerodown provides easy to use graceful restart
// functionality for HTTP server.
package zerodown

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const (
	graceful    = "-graceful"
	waitTimeout = 20 * time.Second // default timeout is 20s
)

type grace struct {
	srv      *http.Server
	listener net.Listener
	err      error
}

func (g *grace) reload() *grace {
	f, err := g.listener.(*net.TCPListener).File()
	if err != nil {
		g.err = err
		return g
	}

	args := os.Args
	if !contains(args, graceful) {
		args = append(args, graceful)
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{f}

	g.err = cmd.Start()
	return g
}

func (g *grace) stop(ctx context.Context) *grace {
	if g.err != nil {
		return g
	}
	if err := g.srv.Shutdown(ctx); err != nil {
		g.err = err
	}
	return g
}

func contains(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}

func (g *grace) run() (err error) {
	if contains(os.Args, graceful) {
		f := os.NewFile(3, "")
		if g.listener, err = net.FileListener(f); err != nil {
			return
		}
	} else {
		if g.listener, err = net.Listen("tcp", g.srv.Addr); err != nil {
			return
		}
	}

	terminate := make(chan error)
	go func() {
		if err := g.srv.Serve(g.listener); err != nil {
			terminate <- err
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit)

	for {
		select {
		case s := <-quit:
			timeout := time.Duration(int64(g.srv.ReadTimeout) + int64(g.srv.WriteTimeout))
			if timeout == 0 {
				timeout = waitTimeout
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			switch s {
			case syscall.SIGINT, syscall.SIGTERM:
				signal.Stop(quit)
				return g.stop(ctx).err
			case syscall.SIGUSR2:
				return g.reload().stop(ctx).err
			}
		case err = <-terminate:
			return
		}
	}
}

// Run accepts a custom http Server and provice signal magic.
func Run(srv *http.Server) error {
	return (&grace{srv: srv}).run()
}

// ListenAndServe wraps http.ListenAndServe and provides signal magic.
func ListenAndServe(addr string, handler http.Handler) error {
	server := &http.Server{Addr: addr, Handler: handler}
	return (&grace{srv: server}).run()
}
