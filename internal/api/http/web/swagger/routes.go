// Package swagger contains Swagger documentation configuration and handlers.
package swagger

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/xelarion/go-layout/internal/api/http/web/swagger/docs" // Import swagger docs
)

// RegisterRoutes registers the Swagger UI routes
func RegisterRoutes(router *gin.Engine) {
	// Serve Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DeepLinking(true),
		ginSwagger.DocExpansion("list"),
	))

	// Redirect /swagger to /swagger/index.html
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
}
