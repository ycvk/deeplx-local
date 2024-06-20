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
	engine.POST(d.routePath, d.Translate)
}
