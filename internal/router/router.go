package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ydesetiawan/mini-netflix/internal/handler"
	"github.com/ydesetiawan/mini-netflix/internal/middleware"
	"github.com/ydesetiawan/mini-netflix/internal/service"
)

func Setup(
	userSvc *service.UserService,
	contentHandler *handler.ContentHandler,
	searchHandler *handler.SearchHandler,
	recHandler *handler.RecommendationHandler,
	userHandler *handler.UserHandler,
) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	auth := middleware.Auth(userSvc)
	optAuth := middleware.OptionalAuth(userSvc)

	// Auth routes
	v1 := r.Group("/api/v1")
	{
		auth_routes := v1.Group("/auth")
		{
			auth_routes.POST("/register", userHandler.Register)
			auth_routes.POST("/login", userHandler.Login)
		}

		// Content routes
		content := v1.Group("/content")
		{
			content.POST("", auth, contentHandler.Create)
			content.GET("/:id", contentHandler.GetByID)
			content.POST("/:id/watch", optAuth, contentHandler.RecordWatch)
			content.POST("/:id/rate", auth, contentHandler.Rate)
			content.POST("/:id/watchlist", auth, userHandler.AddToWatchlist)
		}

		// Search routes
		search := v1.Group("/search")
		{
			search.GET("", searchHandler.Search)
			search.GET("/autocomplete", searchHandler.Autocomplete)
		}

		// Recommendation routes
		rec := v1.Group("/recommendations")
		{
			rec.GET("/popular", recHandler.MostPopular)
			rec.GET("/top-rated", recHandler.MostRecommended)
		}

		// User routes
		user := v1.Group("/user")
		user.Use(auth)
		{
			user.GET("/watchlist", userHandler.GetWatchlist)
		}
	}

	return r
}
