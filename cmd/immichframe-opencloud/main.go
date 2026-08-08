// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Command immichframe-opencloud serves the ImmichFrame HTTP API from an
// OpenCloud space.
//
// This project is developed by Klaas Freitag with the assistance of Claude
// (Anthropic).
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/immichFrame/immichframe-opencloud/internal/frame"
	"github.com/immichFrame/immichframe-opencloud/internal/opencloud"
	"github.com/immichFrame/immichframe-opencloud/pkg/config/defaults"
	"github.com/immichFrame/immichframe-opencloud/pkg/config/parser"
)

// version is overridable at build time: -ldflags "-X main.version=1.2.3".
var version = "dev"

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func run() error {
	insecure := flag.Bool("insecure", false,
		"accept self-signed / invalid TLS certificates from the OpenCloud server (overrides OC_INSECURE)")
	flag.Parse()

	cfg := defaults.FullDefaultConfig()
	if err := parser.ParseConfig(cfg); err != nil {
		return err
	}
	// A set -insecure flag forces TLS verification off; otherwise keep the
	// value derived from OC_INSECURE.
	if isFlagSet("insecure") {
		cfg.OpenCloud.Insecure = *insecure
	}
	if cfg.OpenCloud.Insecure {
		log.Print("WARNING: TLS certificate verification is disabled")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client, err := opencloud.New(initCtx, opencloud.Options{
		BaseURL:     cfg.OpenCloud.URL,
		SpaceID:     cfg.OpenCloud.SpaceID,
		SpaceName:   cfg.OpenCloud.SpaceName,
		Username:    cfg.OpenCloud.Username,
		AppPassword: cfg.OpenCloud.AppPassword,
		BearerToken: cfg.OpenCloud.BearerToken,
		InsecureTLS: cfg.OpenCloud.Insecure,
	})
	if err != nil {
		return err
	}
	log.Printf("connected to OpenCloud %s, serving space %s", cfg.OpenCloud.URL, client.SpaceID())

	catalog := frame.NewCatalog(initCtx, client, cfg.ImmichFrame.CatalogRefresh)
	go catalog.Run(ctx)

	srv := frame.NewServer(cfg.Client, cfg.ImmichFrame.AuthSecret, version, cfg.ImmichFrame.WebRoot, catalog, client)

	httpServer := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	if cfg.ImmichFrame.WebRoot != "" {
		log.Printf("serving web UI from %s", cfg.ImmichFrame.WebRoot)
	}
	log.Printf("listening on %s (auth: %v)", cfg.HTTP.Addr, cfg.ImmichFrame.AuthSecret != "")
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// isFlagSet reports whether the named flag was explicitly provided on the
// command line, so it can override the environment default.
func isFlagSet(name string) bool {
	set := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})
	return set
}
