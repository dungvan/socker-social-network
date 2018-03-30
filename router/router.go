package router

import (
	"net/http"

	"github.com/dungvan2512/socker-social-network/infrastructure"
	"github.com/dungvan2512/socker-social-network/post"
	module "github.com/dungvan2512/socker-social-network/sample-module"
	"github.com/dungvan2512/socker-social-network/shared/base"
	mMiddleware "github.com/dungvan2512/socker-social-network/shared/middleware"
	"github.com/dungvan2512/socker-social-network/user"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Router is application struct hold Mux and db connection
type Router struct {
	Mux                *chi.Mux
	SQLHandler         *infrastructure.SQL
	S3Handler          *infrastructure.S3
	CacheHandler       *infrastructure.Cache
	LoggerHandler      *infrastructure.Logger
	TranslationHandler *infrastructure.Translation
}

// InitializeRouter initializes Mux and middleware
func (r *Router) InitializeRouter() {
	r.Mux.Use(middleware.RequestID)
	r.Mux.Use(middleware.RealIP)
	// Custom middleware(Translation)
	r.Mux.Use(r.TranslationHandler.Middleware.Middleware)
	// Custom middleware(Logger)
	r.Mux.Use(mMiddleware.Logger(r.LoggerHandler))
}

// SetupHandler set database and redis and usecase.
func (r *Router) SetupHandler() {
	// error handler set.
	eh := base.NewHTTPErrorHandler(r.LoggerHandler.Log)
	r.Mux.NotFound(eh.StatusNotFound)
	r.Mux.MethodNotAllowed(eh.StatusMethodNotAllowed)

	r.Mux.Method(http.MethodGet, "/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// base set.
	bh := base.NewHTTPHandler(r.LoggerHandler.Log)
	// base set.
	br := base.NewRepository(r.LoggerHandler.Log)
	// base set.
	bu := base.NewUsecase(r.LoggerHandler.Log)
	// sample set.
	mh := module.NewHTTPHandler(bh, bu, br, r.SQLHandler, r.CacheHandler)
	// user set
	uh := user.NewHTTPHandler(bh, bu, br, r.SQLHandler, r.CacheHandler)
	// post set
	ph := post.NewHTTPHandler(bh, bu, br, r.SQLHandler, r.CacheHandler)
	// authentication middleware.
	// authMiddleware := mMiddleware.JwtAuth(r.LoggerHandler, r.SQLHandler.DB)
	r.Mux.Route("/", func(cr chi.Router) {
		cr.Get("/sample", mh.SampleHandler)
	})

	r.Mux.Route("/users", func(cr chi.Router) {
		cr.Post("/register", uh.Register)
		cr.Post("/login", uh.Login)
	})

	r.Mux.Route("/posts", func(cr chi.Router) {
		cr.With(mMiddleware.JwtAuth(r.LoggerHandler, r.SQLHandler.DB)).Post("/", ph.Create)
	})
}
