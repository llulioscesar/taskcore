package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/storage"
)

type Config struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func DefaultConfig() Config {
	return Config{
		Addr:            ":8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

type Server struct {
	cfg     Config
	db      *sqlx.DB
	storage storage.Storage
	log     *slog.Logger
	http    *http.Server
	mux     *http.ServeMux
}

func New(cfg Config, db *sqlx.DB, store storage.Storage, log *slog.Logger) *Server {
	s := &Server{
		cfg:     cfg,
		db:      db,
		storage: store,
		log:     log,
		mux:     http.NewServeMux(),
	}

	s.http = &http.Server{
		Addr:         cfg.Addr,
		Handler:      s.mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return s
}

func (s *Server) DB() *sqlx.DB       { return s.db }
func (s *Server) Storage() storage.Storage { return s.storage }
func (s *Server) Log() *slog.Logger  { return s.log }
func (s *Server) Mux() *http.ServeMux { return s.mux }

func (s *Server) Start() error {
	s.log.Info("server starting", "addr", s.cfg.Addr)
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("server shutting down")
	return s.http.Shutdown(ctx)
}
