package server

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Middleware for http.Handler
type Middleware func(http.Handler) http.Handler

// App is a http.Handler with Startup and Shutdown lifecycle
type App interface {
	http.Handler

	// Startup controls app startup
	Startup(ctx context.Context) error

	// Shutdown controls app shutdown
	Shutdown(ctx context.Context) error
}

// Server wraps the App with additional http and app lifecycle handling
type Server struct {
	http.Server `json:"-"`
	App         App `json:"-"`
	Address     string
	Port        int
	CertFile    string
	KeyFile     string
	PathPrefix  string
	Logger      *zap.Logger `json:"-"`
	Debug       bool
}

// New create new Server
func New(app App, options ...Option) *Server {
	s := &Server{}
	s.App = app
	s.Port = 9000
	s.MaxHeaderBytes = 1 << 20
	s.Logger = zap.NewNop()

	s.Handler = pathHandler(http.MethodGet, map[string]http.HandlerFunc{
		"/favicon.ico": handleFavicon,
		"/health":      handleHealth,
	})(s.App)

	for _, option := range options {
		option(s)
	}
	if s.PathPrefix != "" {
		s.Handler = http.StripPrefix(s.PathPrefix, s.Handler)
	}
	s.Handler = s.panicHandler(s.Handler)
	s.Addr = s.Address + ":" + strconv.Itoa(s.Port)

	if s.Debug {
		s.Logger.Debug("config", zap.Any("server", s))
	}
	return s
}

func (s *Server) Run() {
	if err := s.App.Startup(context.Background()); err != nil {
		s.Logger.Fatal("app startup", zap.Error(err))
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.listenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Logger.Fatal("listen", zap.Error(err))
		}
	}()

	s.Logger.Info("server start", zap.String("addr", s.Addr))
	<-done

	// graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		s.Logger.Error("server shutdown", zap.Error(err))
	}
	if err := s.App.Shutdown(ctx); err != nil {
		s.Logger.Error("app shutdown", zap.Error(err))
	}
	s.Logger.Info("exit")
	return
}

func (s *Server) listenAndServe() error {
	if s.CertFile != "" && s.KeyFile != "" {
		return s.ListenAndServeTLS(s.CertFile, s.KeyFile)
	}
	return s.ListenAndServe()
}