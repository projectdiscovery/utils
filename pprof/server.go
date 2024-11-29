package pprof

import (
	"context"
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"

	"github.com/felixge/fgprof"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/utils/env"
)

const (
	PPROFServerAddressENV = "PPROF_SERVER_ADDRESS"
)

type PprofServer struct {
	server *http.Server
}

func NewPprofServer() *PprofServer {
	address := env.GetEnvOrDefault(PPROFServerAddressENV, "127.0.0.1:8086")

	mux := http.NewServeMux()
	// Default pprof handlers
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Also add fgprof for more detailed profiling
	mux.Handle("/debug/fgprof", fgprof.Handler())

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}
	// Enable block and mutex profiling as well
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)

	return &PprofServer{server: server}
}

func (p *PprofServer) Start() {
	gologger.Info().Msgf("Listening pprof debug server on: %s", p.server.Addr)

	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			gologger.Error().Msgf("pprof server failed to start: %s", err)
		}
	}()
}

func (p *PprofServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = p.server.Shutdown(ctx)
}
