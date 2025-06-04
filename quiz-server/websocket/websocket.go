package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"quiz/configs"
	"quiz/core/data"
	"quiz/core/managers"
	"quiz/core/models"
	"quiz/workflow"

	socketio "github.com/karagenc/socket.io-go"
)

type webSocketHandler struct {
	quizSessionManager *managers.QuizSession
	server             *socketio.Server
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		next.ServeHTTP(w, r)
	})
}

func ListenAndHandleEvent(manager *managers.QuizSession, server *socketio.Server) {
	portStr := os.Getenv("PORT")
	_, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalln("invalid port", portStr)
	}

	handler := &webSocketHandler{
		quizSessionManager: manager,
		server:             server,
	}
	server.Of("/").OnConnection(func(socket socketio.ServerSocket) {
		fmt.Println("on connect:", socket.ID())
		socket.OnEvent(string(configs.AnswerQuestion), handler.onQuestionAnswered(socket))
		socket.OnEvent(string(configs.JoinQuiz), handler.onJoinQuiz(socket))

		socket.OnDisconnect(func(reason socketio.Reason) {
			fmt.Println("on disconnect:", reason)
		})
	})
	if err := server.Run(); err != nil {
		log.Fatalln(err)
	}

	fs := http.FileServer(http.Dir("public"))
	router := http.NewServeMux()
	router.Handle("/socket.io/", corsMiddleware(server))
	router.Handle("/", fs)
	// Define a simple GET route
	router.HandleFunc("/start/", startQuiz)

	httpServer := &http.Server{
		Addr:    "0.0.0.0:" + portStr,
		Handler: router,

		// It is always a good practice to set timeouts.
		ReadTimeout: 120 * time.Second,
		IdleTimeout: 120 * time.Second,

		// HTTPWriteTimeout returns io.PollTimeout + 10 seconds (extra 10 seconds to write the response).
		// You should either set this timeout to 0 (infinite) or some value greater than the io.PollTimeout.
		// Otherwise poll requests may fail.
		WriteTimeout: server.HTTPWriteTimeout(),
	}

	fmt.Println("Listening on:", portStr)
	if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln(err)
	}
}

func startQuiz(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract `quiz_id` from the URL path
	quizIdStr := r.URL.Path[len("/start/"):]
	quizId, err := strconv.Atoi(quizIdStr)
	if err != nil {
		http.Error(w, jsonError(err.Error()), http.StatusBadRequest)
		return
	}

	// Fetch quiz data
	quiz := data.QuizData[models.QuizId(quizId)]
	if quiz == nil {
		http.Error(w, jsonError("quiz not found"), http.StatusNotFound)
		return
	}

	// Start the quiz workflow
	if err = workflow.StartQuizWorkflow(r.Context(), quiz); err != nil {
		http.Error(w, jsonError(err.Error()), http.StatusInternalServerError)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "start quiz successfully",
	})
}

// jsonError creates a JSON error response string
func jsonError(message string) string {
	response := map[string]string{"error": message}
	jsonBytes, _ := json.Marshal(response) // Ignoring errors here for simplicity
	return string(jsonBytes)
}

var JoinQuizError = errors.New("JoinQuizError")

func (h *webSocketHandler) onJoinQuiz(s socketio.ServerSocket) func(username string, quizId int) {
	return func(username string, quizId int) {
		if username == "" {
			s.Emit(string(configs.Error), fmt.Sprintf("%s: user id is empty", JoinQuizError))
			return
		}

		quiz, err := h.quizSessionManager.JoinQuiz(context.Background(), models.QuizId(quizId), models.Username(username), s)
		if err != nil {
			s.Emit(string(configs.Error), fmt.Sprintf("%s: %s", JoinQuizError, err))
			fmt.Println(fmt.Sprintf("join quiz err: %s", err))
			return
		}
		s.Emit(string(configs.QuizData), quiz)
		fmt.Println("join quiz successfully. username:", username, "quizid:", quizId)
		h.server.SocketsJoin(socketio.Room(models.QuizId(quizId).String()))

		return
	}
}

func (h *webSocketHandler) onQuestionAnswered(s socketio.ServerSocket) func(msg string) {
	return func(msg string) {
		fmt.Println("on question answered", msg)
		answer := &models.QuestionAnsweredPayload{}
		err := json.Unmarshal([]byte(msg), answer)
		if err != nil {
			s.Emit(string(configs.Error), "invalid data")
			return
		}

		res, err := h.quizSessionManager.AnswerQuestion(s, answer.QuizId, answer.QuestionIndex, answer.AnswerIndex)
		if err != nil {
			s.Emit(string(configs.Error), err.Error())
			fmt.Println("handle question answered websocket event error:", err)
			return
		}
		s.Emit(string(configs.AnswerChecked), res.CorrectAnswerIndex, res.NewScore)
		return
	}
}
