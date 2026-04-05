package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

type Server struct {
	httpServer     *http.Server
	redirectServer *http.Server // HTTP→HTTPS redirect + ACME challenges
}

type TLSConfig struct {
	Domain  string
	CertDir string
}

func New(addr string, handler http.Handler, tlsCfg *TLSConfig) *Server {
	srv := &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 120 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	if tlsCfg != nil && tlsCfg.Domain != "" {
		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(tlsCfg.Domain),
			Cache:      autocert.DirCache(tlsCfg.CertDir),
		}

		srv.httpServer.Addr = ":443"
		srv.httpServer.TLSConfig = &tls.Config{
			GetCertificate: m.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		}

		// HTTP server for ACME challenges and HTTPS redirect
		srv.redirectServer = &http.Server{
			Addr:         ":80",
			Handler:      m.HTTPHandler(nil), // nil = redirect non-ACME to HTTPS
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
	}

	return srv
}

func (s *Server) Start() error {
	if s.redirectServer != nil {
		go func() {
			slog.Info("starting HTTP redirect server", "addr", ":80")
			if err := s.redirectServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("redirect server error", "error", err)
			}
		}()

		slog.Info("starting HTTPS server", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("tls server error: %w", err)
		}
		return nil
	}

	slog.Info("starting HTTP server", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("shutting down server")
	if s.redirectServer != nil {
		if err := s.redirectServer.Shutdown(ctx); err != nil {
			slog.Error("redirect server shutdown error", "error", err)
		}
	}
	return s.httpServer.Shutdown(ctx)
}
