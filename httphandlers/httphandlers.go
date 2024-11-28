package httphandlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"quiz/core/managers"
	"quiz/core/models"
)

func StartHttpServer(quizSessionManager *managers.QuizSession) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	// Define a simple GET route
	router.GET("/start/:quiz_id", func(c *gin.Context) {
		quizIdStr := c.Param("quiz_id")
		quizId, err := strconv.Atoi(quizIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if err = quizSessionManager.StartQuiz(c.Request.Context(), models.QuizId(quizId)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "start quiz successfully",
		})
	})

	go router.Run(":1918")
}
