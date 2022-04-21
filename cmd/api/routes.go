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

	// Banners
	router.HandlerFunc(http.MethodGet, "/api/banners", app.getBannersHandler)

	// Blogs
	router.HandlerFunc(http.MethodGet, "/api/blogs", app.getBlogsHandler)
	router.HandlerFunc(http.MethodGet, "/api/blog/:slug", app.getBlogBySlugHandler)
	router.HandlerFunc(http.MethodGet, "/api/blogs/recommendation", app.getBlogsRecommendationHandler)

	// Blog Categories
	router.HandlerFunc(http.MethodGet, "/api/blog-categories", app.getBlogCategoriesHandler)
	router.HandlerFunc(http.MethodGet, "/api/blog-categories/:slug", app.getBlogCategoriesBySlugHandler)

	// Brands
	router.HandlerFunc(http.MethodGet, "/api/brands", app.getBrandsHandler)

	// Product Categories
	router.HandlerFunc(http.MethodGet, "/api/product-categories", app.getProductCategoriesHandler)
	router.HandlerFunc(http.MethodGet, "/api/product-categories/:slug", app.getProductCategoriesBySlugHandler)

	// ====================================================================================
	// CMS - Backoffice Routes
	// ====================================================================================

	// Banners
	router.HandlerFunc(http.MethodGet, "/sql/banners", app.listBannersHandler)
	router.HandlerFunc(http.MethodGet, "/sql/banners/:id", app.showBannerHandler)
	router.HandlerFunc(http.MethodPost, "/sql/banners", app.createBannerHandler)
	router.HandlerFunc(http.MethodPut, "/sql/banners/:id", app.fullUpdateBannerHandler)
	router.HandlerFunc(http.MethodDelete, "/sql/banners/:id", app.deleteBannerHandler)

	router.HandlerFunc(http.MethodGet, "/cms/banners", app.gormListBannerHandler)
	router.HandlerFunc(http.MethodGet, "/cms/banners/:id", app.gormShowBannerHandler)
	router.HandlerFunc(http.MethodPost, "/cms/banners", app.gormCreateBannerHandler)
	router.HandlerFunc(http.MethodPut, "/cms/banners/:id", app.gormFullUpdateBannerHandler)
	router.HandlerFunc(http.MethodDelete, "/cms/banners/:id", app.gormDeleteBannerHandler)

	// Blogs
	router.HandlerFunc(http.MethodGet, "/cms/blogs", app.listBlogsHandler)
	router.HandlerFunc(http.MethodGet, "/cms/blogs/:id", app.showBlogHandler)
	router.HandlerFunc(http.MethodPost, "/cms/blogs", app.createBlogHandler)
	router.HandlerFunc(http.MethodPut, "/cms/blogs/:id", app.updateBlogHandler)
	router.HandlerFunc(http.MethodDelete, "/cms/blogs/:id", app.deleteBlogHandler)

	// Blog Categories
	router.HandlerFunc(http.MethodGet, "/cms/blog-categories", app.listBlogCategoriesHandler)
	router.HandlerFunc(http.MethodGet, "/cms/blog-categories/:id", app.showBlogCategoryHandler)
	router.HandlerFunc(http.MethodPost, "/cms/blog-categories", app.createBlogCategoryHandler)
	router.HandlerFunc(http.MethodPut, "/cms/blog-categories/:id", app.updateBlogCategoryHandler)
	router.HandlerFunc(http.MethodDelete, "/cms/blog-categories/:id", app.deleteBlogCategoryHandler)

	// Brands
	router.HandlerFunc(http.MethodGet, "/cms/brands", app.listBrandsHandler)
	router.HandlerFunc(http.MethodGet, "/cms/brands/:id", app.showBrandHandler)
	router.HandlerFunc(http.MethodPost, "/cms/brands", app.createBrandHandler)
	router.HandlerFunc(http.MethodPut, "/cms/brands/:id", app.updateBrandHandler)
	router.HandlerFunc(http.MethodDelete, "/cms/brands/:id", app.deleteBrandHandler)

	// Movies
	router.HandlerFunc(http.MethodGet, "/cms/movies", app.requireActivatedUser(app.listMoviesHandler))
	router.HandlerFunc(http.MethodPost, "/cms/movies", app.requirePermission("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/cms/movies/:id", app.requireActivatedUser(app.showMovieHandler))
	router.HandlerFunc(http.MethodPut, "/cms/movies/:id", app.requireActivatedUser(app.fullUpdateMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/cms/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/cms/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))

	// Product Categories
	router.HandlerFunc(http.MethodGet, "/cms/product-categories", app.listProductCategoriesHandler)
	router.HandlerFunc(http.MethodGet, "/cms/product-categories/:id", app.showProductCategoryHandler)
	router.HandlerFunc(http.MethodPost, "/cms/product-categories", app.createProductCategoryHandler)
	router.HandlerFunc(http.MethodPut, "/cms/product-categories/:id", app.updateProductCategoryHandler)
	router.HandlerFunc(http.MethodDelete, "/cms/product-categories/:id", app.deleteProductCategoryHandler)

	// ====================================================================================
	// Miscellaneous Routes
	// ====================================================================================

	router.HandlerFunc(http.MethodGet, "/", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	// Return the httprouter instance.
	// Wrap the router with the panic recovery middleware.
	// Use the authenticate() middleware on all requests.
	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
