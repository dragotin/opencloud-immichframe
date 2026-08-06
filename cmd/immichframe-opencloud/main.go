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

	"github.com/immichFrame/immichframe-opencloud/internal/config"
	"github.com/immichFrame/immichframe-opencloud/internal/frame"
	"github.com/immichFrame/immichframe-opencloud/internal/opencloud"
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
		"accept self-signed / invalid TLS certificates from the OpenCloud server (overrides OPENCLOUD_INSECURE_TLS)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	// A set -insecure flag forces TLS verification off; otherwise keep the
	// value derived from OPENCLOUD_INSECURE_TLS.
	if isFlagSet("insecure") {
		cfg.InsecureTLS = *insecure
	}
	if cfg.InsecureTLS {
		log.Print("WARNING: TLS certificate verification is disabled")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client, err := opencloud.New(initCtx, opencloud.Options{
		BaseURL:     cfg.OpenCloudBaseURL,
		SpaceID:     cfg.SpaceID,
		SpaceName:   cfg.SpaceName,
		Username:    cfg.Username,
		AppPassword: cfg.AppPassword,
		BearerToken: cfg.BearerToken,
		InsecureTLS: cfg.InsecureTLS,
	})
	if err != nil {
		return err
	}
	log.Printf("connected to OpenCloud %s, serving space %s", cfg.OpenCloudBaseURL, client.SpaceID())

	catalog := frame.NewCatalog(initCtx, client, cfg.CatalogRefresh)
	go catalog.Run(ctx)

	srv := frame.NewServer(cfg.Client, cfg.AuthSecret, version, cfg.WebRoot, catalog, client)

	httpServer := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	if cfg.WebRoot != "" {
		log.Printf("serving web UI from %s", cfg.WebRoot)
	}
	log.Printf("listening on %s (auth: %v)", cfg.ListenAddr, cfg.AuthSecret != "")
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
