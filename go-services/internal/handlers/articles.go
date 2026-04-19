package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vnstock-hybrid/internal/models"
	"vnstock-hybrid/internal/services"
)

func ListArticles(svc *services.ArticleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.DefaultQuery("status", "published")
		symbol := c.Query("symbol")
		search := c.Query("q")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		if limit > 100 {
			limit = 100
		}
		result, err := svc.List(status, symbol, search, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func RelatedArticles(svc *services.ArticleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		symbol := c.Query("symbol")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "3"))
		if limit > 10 {
			limit = 10
		}
		articles, err := svc.GetRelated(slug, symbol, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"articles": articles})
	}
}

func TrendingArticles(svc *services.ArticleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
		if days > 30 {
			days = 30
		}
		if limit > 20 {
			limit = 20
		}
		items, err := svc.Trending(days, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"trending": items})
	}
}

func ScreenerHandler(svc *services.ArticleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rec := c.Query("recommendation")
		minConf, _ := strconv.ParseFloat(c.DefaultQuery("min_confidence", "0"), 64)
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		if limit > 200 {
			limit = 200
		}
		items, err := svc.Screener(rec, minConf, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"results": items, "total": len(items)})
	}
}

func GetArticleBySlug(svc *services.ArticleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		article, err := svc.GetBySlug(slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		c.JSON(http.StatusOK, article)
	}
}

func CreateArticle(svc *services.ArticleService, internalKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if internalKey != "" && c.GetHeader("X-Internal-Key") != internalKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		var input services.CreateArticleInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		article, err := svc.Create(input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, article)
	}
}

func UpdateArticleStatus(svc *services.ArticleService, internalKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if internalKey != "" && c.GetHeader("X-Internal-Key") != internalKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		idStr := c.Param("id")
		id64, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var input services.UpdateStatusInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		article, err := svc.UpdateStatus(uint(id64), models.ArticleStatus(input.Status))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, article)
	}
}
