package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// Initialize a new httprouter router instance.
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// ====================================================================================
	// API - Business Routes
	// ====================================================================================

	// Users
	router.HandlerFunc(http.MethodPost, "/api/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodGet, "/api/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPut, "/api/users/password", app.updateUserPasswordHandler)

	// Tokens
	router.HandlerFunc(http.MethodPost, "/api/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/api/tokens/activation", app.createActivationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/api/tokens/password-reset", app.createPasswordResetTokenHandler)

	// ====================================================================================
	// CMS - Backoffice Routes
	// ====================================================================================

	// Movies
	router.HandlerFunc(http.MethodGet, "/cms/movies", app.requireActivatedUser(app.listMoviesHandler))
	// router.HandlerFunc(http.MethodGet, "/movies", app.listMoviesHandler)
	router.HandlerFunc(http.MethodPost, "/cms/movies", app.requirePermission("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/cms/movies/:id", app.requireActivatedUser(app.showMovieHandler))
	router.HandlerFunc(http.MethodPut, "/cms/movies/:id", app.requireActivatedUser(app.fullUpdateMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/cms/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/cms/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))

	// ====================================================================================
	// Miscellaneous Routes
	// ====================================================================================

	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	// Return the httprouter instance.
	// Wrap the router with the panic recovery middleware.
	// Use the authenticate() middleware on all requests.
	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
