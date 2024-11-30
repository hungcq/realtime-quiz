package httphandlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"quiz/core/data"
	"quiz/core/models"
	"quiz/workflow"
)

func StartHttpServer() {
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

		quiz := data.QuizData[models.QuizId(quizId)]
		if quiz == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "quiz not found",
			})
		}

		if err = workflow.StartQuizWorkflow(c.Request.Context(), quiz); err != nil {
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
