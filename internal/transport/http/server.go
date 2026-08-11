package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcclients "github.com/marketing-digest/gateway/internal/clients/grpc"
	"github.com/marketing-digest/gateway/internal/middleware"
)

// Server is the thin HTTP gateway. Handlers only translate HTTP ↔ gRPC.
type Server struct {
	http         *http.Server
	clients      *grpcclients.Clients
	log          *slog.Logger
	ownerSecret  string
}

func New(addr string, clients *grpcclients.Clients, log *slog.Logger, corsOrigins, ownerSecret string) *Server {
	mux := http.NewServeMux()
	s := &Server{clients: clients, log: log, ownerSecret: ownerSecret}

	mux.HandleFunc("GET /healthz", s.handleLiveness)
	mux.HandleFunc("GET /readyz", s.handleReadiness)
	mux.HandleFunc("GET /v1/auth/ping", s.handleAuthPing)
	mux.HandleFunc("GET /v1/blog/ping", s.handleBlogPing)

	// Users
	mux.HandleFunc("GET /v1/users/me", s.handleGetMe)
	mux.HandleFunc("GET /v1/usernames/{username}", s.handleGetUserByUsername)
	mux.HandleFunc("GET /v1/users/{userID}", s.handleGetPublicProfile)
	mux.HandleFunc("GET /v1/users/{userID}/analytics", s.handleGetUserAnalytics)

	// Blogs (public reads)
	mux.HandleFunc("GET /v1/blogs", s.handleListBlogs)
	mux.HandleFunc("GET /v1/blogs/{blogID}", s.handleGetBlog)
	mux.HandleFunc("POST /v1/blogs/{blogID}/upvotes", s.handleUpvoteBlog)

	// Blogs (studio writes — owner secret)
	mux.Handle("POST /v1/blogs", s.requireOwner(http.HandlerFunc(s.handleCreateBlog)))
	mux.Handle("PATCH /v1/blogs/{blogID}", s.requireOwner(http.HandlerFunc(s.handleUpdateBlog)))

	// Studio articles
	mux.HandleFunc("GET /v1/blogs/{blogID}/articles", s.handleListArticles)
	mux.HandleFunc("GET /v1/blogs/{blogID}/articles/{articleID}", s.handleGetArticle)
	mux.Handle("POST /v1/blogs/{blogID}/articles", s.requireOwner(http.HandlerFunc(s.handleCreateArticle)))
	mux.Handle("PATCH /v1/blogs/{blogID}/articles/{articleID}", s.requireOwner(http.HandlerFunc(s.handleUpdateArticle)))
	mux.Handle("DELETE /v1/blogs/{blogID}/articles/{articleID}", s.requireOwner(http.HandlerFunc(s.handleDeleteArticle)))
	mux.HandleFunc("POST /v1/blogs/{blogID}/articles/{articleID}/upvotes", s.handleUpvoteArticle)
	mux.Handle("POST /v1/blogs/{blogID}/articles/{articleID}/publish", s.requireOwner(http.HandlerFunc(s.handlePublishArticle)))
	mux.Handle("POST /v1/blogs/{blogID}/articles/{articleID}/draft", s.requireOwner(http.HandlerFunc(s.handleEnsureDraft)))
	mux.Handle("GET /v1/blogs/{blogID}/articles/{articleID}/draft", s.requireOwner(http.HandlerFunc(s.handleGetDraft)))
	mux.Handle("PATCH /v1/blogs/{blogID}/articles/{articleID}/draft", s.requireOwner(http.HandlerFunc(s.handleUpdateDraft)))
	mux.Handle("POST /v1/blogs/{blogID}/articles/{articleID}/draft/publish", s.requireOwner(http.HandlerFunc(s.handlePublishDraft)))

	// Public articles (register /trending before /{articleID})
	mux.HandleFunc("GET /v1/articles", s.handleListPublished)
	mux.HandleFunc("GET /v1/articles/trending", s.handleListTrending)
	mux.HandleFunc("GET /v1/articles/{articleID}", s.handleGetPublishedArticle)
	mux.HandleFunc("GET /v1/articles/{articleID}/related", s.handleListRelated)
	mux.HandleFunc("POST /v1/articles/{articleID}/views", s.handleRecordView)
	mux.HandleFunc("POST /v1/articles/{articleID}/reads", s.handleRecordRead)
	mux.HandleFunc("GET /v1/articles/{articleID}/comments", s.handleListComments)
	mux.HandleFunc("POST /v1/articles/{articleID}/comments", s.handleCreateComment)

	handler := middleware.Chain(
		mux,
		middleware.WithRecovery(log),
		middleware.WithRequestID,
		middleware.WithLogging(log),
		middleware.WithCORS(corsOrigins),
	)

	s.http = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

// requireOwner checks X-Owner-Secret against OWNER_STUDIO_SECRET.
func (s *Server) requireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.ownerSecret == "" || r.Header.Get("X-Owner-Secret") != s.ownerSecret {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "owner secret required"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) ListenAndServe() error {
	s.log.Info("http listening", "addr", s.http.Addr)
	err := s.http.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func (s *Server) Handler() http.Handler {
	return s.http.Handler
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	code := http.StatusInternalServerError
	msg := "internal error"
	switch st.Code() {
	case codes.InvalidArgument:
		code, msg = http.StatusBadRequest, "invalid argument"
	case codes.NotFound:
		code, msg = http.StatusNotFound, "not found"
	case codes.Unauthenticated:
		code, msg = http.StatusUnauthorized, "unauthenticated"
	case codes.PermissionDenied:
		code, msg = http.StatusForbidden, "permission denied"
	case codes.AlreadyExists:
		code, msg = http.StatusConflict, "conflict"
	case codes.Unavailable:
		code, msg = http.StatusServiceUnavailable, "service unavailable"
	case codes.DeadlineExceeded:
		code, msg = http.StatusGatewayTimeout, "deadline exceeded"
	}
	writeJSON(w, code, map[string]string{"error": msg})
}
