package web

import (
	"deeplx-local/domain"
	"deeplx-local/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DeepLXHandler struct {
	service   service.TranslateService
	routePath string
}

func NewDeepLXHandler(service service.TranslateService, customRoute string) *DeepLXHandler {
	if customRoute == "" {
		customRoute = "/translate"
	}
	if customRoute[0] != '/' {
		customRoute = "/" + customRoute
	}
	return &DeepLXHandler{service: service, routePath: customRoute}
}

func (d *DeepLXHandler) Translate(c *gin.Context) {
	var request domain.TranslateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	translatedText := d.service.GetTranslateData(request)
	c.JSON(http.StatusOK, translatedText)
}

func (d *DeepLXHandler) RegisterRoutes(engine *gin.Engine) {
	engine.Use(Cors())
	engine.POST(d.routePath, d.Translate)
}

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			c.Header("Access-Control-Max-Age", "86400") // 24 hours
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
