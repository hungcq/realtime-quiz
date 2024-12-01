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

	socketio "github.com/karagenc/socket.io-go"
	"quiz/configs"
	"quiz/core/managers"
	"quiz/core/models"
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

var JoinQuizError = errors.New("JoinQuizError")

func (h *webSocketHandler) onJoinQuiz(s socketio.ServerSocket) func(userIdStr, quizIdStr string) {
	return func(userIdStr, quizIdStr string) {
		if userIdStr == "" {
			s.Emit(string(configs.Error), fmt.Sprintf("%s: user id is empty", JoinQuizError))
			return
		}
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			s.Emit(string(configs.Error), fmt.Sprintf("%s: invalid user id", JoinQuizError))
			return
		}

		if quizIdStr == "" {
			s.Emit(string(configs.Error), fmt.Sprintf("%s: quiz id is empty", JoinQuizError))
			return
		}
		quizId, err := strconv.Atoi(quizIdStr)
		if err != nil {
			s.Emit(string(configs.Error), fmt.Sprintf("%s: invalid quiz id", JoinQuizError))
			return
		}

		quiz, err := h.quizSessionManager.JoinQuiz(context.Background(), models.QuizId(quizId), models.UserId(userId), s)
		if err != nil {
			s.Emit(string(configs.Error), fmt.Sprintf("%s: %s", JoinQuizError, err))
			fmt.Println(fmt.Sprintf("join quiz err: %s", err))
			return
		}
		s.Emit(string(configs.QuizData), quiz)
		fmt.Println("join quiz successfully. userid:", userIdStr, "quizid:", quizIdStr)
		h.server.SocketsJoin(socketio.Room(quizIdStr))
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
