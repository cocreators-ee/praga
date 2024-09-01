package backend

import (
	"context"
	"errors"
	"fmt"
	"github.com/cocreators-ee/praga"
	"github.com/go-chi/chi/v5/middleware"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/unrolled/secure"
)

const DEBUG = false

type Server struct {
	Config        Config
	MailjetSender *MailjetSender
}

func (s *Server) getRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)

	secureMiddleware := secure.New(secure.Options{
		STSSeconds:            31536000,
		STSIncludeSubdomains:  true,
		STSPreload:            true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      false,
		ReferrerPolicy:        "no-referrer",
		ContentSecurityPolicy: "script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self'",
	})

	r.Use(secureMiddleware.Handler)

	// API Routes
	registerRoutes(s, r)

	// Embedded frontend build files
	buildFs, err := fs.Sub(praga.EmbeddedFrontendBuild, "frontend/build")
	if err != nil {
		panic(err)
	}

	if DEBUG {
		log.Print("Embedded filesystem to be served as static files:")
		fs.WalkDir(buildFs, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				fmt.Printf("%s/\n", path)
			} else {
				println(path)
			}
			return nil
		})
	}

	filesDir := http.FS(buildFs)
	FileServer(r, "/", filesDir)

	return r
}

func (s *Server) Start() {
	var listener net.Listener
	var err error

	if s.Config.Server.ListenType == "http" {
		addr := fmt.Sprintf("%s:%d", s.Config.Server.Host, s.Config.Server.Port)
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			log.Panicf("Error trying to listen to %s: %s", addr, err)
		}

		log.Printf("Listening to http://%s", addr)
	} else if s.Config.Server.ListenType == "unix" {
		// TODO: Test
		listener, err = net.Listen("unix", s.Config.Server.Socket)
		if err != nil {
			log.Panicf("Error trying to listen to socket %s: %s", s.Config.Server.Socket, err)
		}

		log.Printf("Listening to unix://%s", s.Config.Server.Socket)

		defer func() {
			err := os.Remove(s.Config.Server.Socket)
			if err != nil && os.IsNotExist(err) {
				log.Printf("Released socket unix://%s\n", s.Config.Server.Socket)
			} else if err != nil {
				log.Printf("Error releasing %s: %s\n", s.Config.Server.Socket, err)
			}
		}()
	} else {
		log.Panicf("Invalid listen_type %s", s.Config.Server.ListenType)
	}

	server := &http.Server{
		Handler: s.getRouter(),
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		quitSignal := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(quitSignal, os.Interrupt)
		// sigterm signal sent from kubernetes
		signal.Notify(quitSignal, syscall.SIGTERM)

		<-quitSignal

		// We received an interrupt signal, shut down.
		if err := server.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	if err := server.Serve(listener); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Error from server: %s\n", err)
		}
	}

	<-idleConnsClosed
}

func NewServer(config Config) *Server {
	s := &Server{
		Config: config,
	}

	// If mailjet is configured setup the client
	if s.Config.Mailjet.APIKeyPublic != "" {
		s.MailjetSender = getMailjetSender(s)
	}

	return s
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fileServer := http.StripPrefix(pathPrefix, http.FileServer(root))
		fileServer.ServeHTTP(w, r)
	})
}
