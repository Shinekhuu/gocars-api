package handler

import (
	"net/http"
	"strconv"

	"gocars-api/internal/articles/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OemHandler struct {
	svc *service.OemService
}

func NewOemHandler(svc *service.OemService) *OemHandler {
	return &OemHandler{svc: svc}
}

func (h *OemHandler) GetOEMs(c *gin.Context) {
	parseArray := func(values []string) []*int {
		var result []*int
		for _, s := range values {
			if s == "" {
				result = append(result, nil)
				continue
			}
			if v, err := strconv.Atoi(s); err == nil {
				val := v
				result = append(result, &val)
			} else {
				result = append(result, nil)
			}
		}
		return result
	}

	ids := parseArray(c.QueryArray("id[]"))
	articleIDs := parseArray(c.QueryArray("article_id[]"))

	res, err := h.svc.GetOEMs(ids, articleIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *OemHandler) GetAIPart(c *gin.Context) {
	oem := c.DefaultQuery("oem", "")
	name := c.DefaultQuery("name", "")
	brand := c.DefaultQuery("brand", "")
	specs := c.DefaultQuery("specifications", "")

	car, err := service.GetResponseOpenAi(oem, name, brand, specs)
	if err != nil {
		zap.L().Error("GetResponseOpenAi error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"responses": car})
}

func (h *OemHandler) GetAIMapper(c *gin.Context) {
	name := c.DefaultQuery("name", "")
	oem := c.DefaultQuery("oem", "")
	brand := c.DefaultQuery("brand", "")
	modelStr := c.DefaultQuery("input", "")

	result, err := service.MapWithAI(name, oem, modelStr, brand)
	if err != nil {
		zap.L().Error("MapWithAI error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get mapping"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mapping": result})
}
