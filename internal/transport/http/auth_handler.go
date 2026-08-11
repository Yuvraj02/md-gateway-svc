package httpserver

import (
	"net/http"

	authv1 "github.com/Yuvraj02/md-protos/proto/auth/v1"
)

func (s *Server) handleAuthPing(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("message")
	resp, err := s.clients.Auth.Ping(r.Context(), &authv1.PingRequest{Message: msg})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": resp.GetMessage(), "service": resp.GetService()})
}

// GET /v1/users/me - return the seeded site owner (no OAuth yet).
func (s *Server) handleGetMe(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Auth.GetMe(r.Context(), &authv1.GetMeRequest{})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, userJSON(resp.User))
}

// GET /v1/usernames/{username}
func (s *Server) handleGetUserByUsername(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Auth.GetUserByUsername(r.Context(), &authv1.GetUserByUsernameRequest{
		Username: r.PathValue("username"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, userJSON(resp.User))
}

// GET /v1/users/{userID} - public profile by id.
func (s *Server) handleGetPublicProfile(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Auth.GetPublicProfile(r.Context(), &authv1.GetPublicProfileRequest{
		UserId: r.PathValue("userID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, userJSON(resp.User))
}

// GET /v1/users/{userID}/analytics
func (s *Server) handleGetUserAnalytics(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Auth.GetUserAnalytics(r.Context(), &authv1.GetUserAnalyticsRequest{
		UserId: r.PathValue("userID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, analyticsJSON(resp.Analytics))
}
