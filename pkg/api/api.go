package api

import (
	"fmt"
	"github.com/analog-substance/arsenic/pkg/api/controller/host"
	"github.com/analog-substance/arsenic/pkg/api/controller/lead"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func Serve(port int) error {

	router := gin.Default()

	// Default Config:
	// - No origin allowed by default
	// - GET, POST, PUT, HEAD methods
	// - Credentials share disabled
	// - Preflight requests cached for 12 hours
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AddAllowMethods("OPTIONS")
	router.Use(cors.New(config))

	api := router.Group("/api")
	api.GET("/ping", ping)

	api_host := api.Group("/host")
	host.AddRoutes(api_host)

	api_lead := api.Group("/lead")
	lead.AddRoutes(api_lead)

	address := fmt.Sprintf("localhost:%d", port)
	fmt.Printf("[+] Listening on %s\n", address)
	return router.Run(address)
}
