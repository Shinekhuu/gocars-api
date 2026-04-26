package controllers

import (
	"gocars-api/services"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 🔥 GET /oems?id=1,23,5&category_id=123,5356
func GetOEMs(c *gin.Context) {
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

	// 🔥 NEW format
	ids := parseArray(c.QueryArray("id[]"))
	articleIDs := parseArray(c.QueryArray("article_id[]"))

	res, err := services.GetOEMs(ids, articleIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

func GetResponse(c *gin.Context) {
	oem := c.DefaultQuery("oem", "")
	name := c.DefaultQuery("name", "")
	brand := c.DefaultQuery("brand", "")
	modelOrSpecifications := c.DefaultQuery("specifications", "")

	car, err := services.GetResponseOpenAi(oem, name, brand, modelOrSpecifications)
	if err != nil {
		log.Println(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"responses": car,
	})
}

func GetMapper(c *gin.Context) {
	name := c.DefaultQuery("name", "")
	oem := c.DefaultQuery("oem", "")
	brand := c.DefaultQuery("brand", "")
	modelStr := c.DefaultQuery("input", "")

	result, err := services.MapWithAI(name, oem, modelStr, brand)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get mapping"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mapping": result,
	})
}
