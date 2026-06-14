package rest

import "net/http"

func RegisterRoutes(
	mux *http.ServeMux,
	authHandler *AuthHandler,
	linkHandler *LinkHandler,
	authMiddleware func(http.Handler) http.Handler,
) {
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.Handle("POST /api/v1/links", authMiddleware(http.HandlerFunc(linkHandler.Create)))
	mux.Handle("GET /api/v1/links", authMiddleware(http.HandlerFunc(linkHandler.List)))
	mux.Handle("DELETE /api/v1/links/{id}", authMiddleware(http.HandlerFunc(linkHandler.Delete)))
	mux.Handle("GET /api/v1/links/{id}/stats", authMiddleware(http.HandlerFunc(linkHandler.Stats)))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /{code}", linkHandler.Redirect)
}
