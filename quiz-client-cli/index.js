const io = require('socket.io-client');
const prompt = require('prompt-sync')();

function printLeaderboard(leaderboard) {
    if (leaderboard && leaderboard.length > 0) {
        console.log('===== LEADERBOARD =====')
        for (let i = 0; i < leaderboard.length; i++) {
            console.log(`${i + 1}.`, leaderboard[i])
        }
        console.log()
    }
}

const JoinQuiz = "join_quiz"
const AnswerQuestion = "answer_question"

const AnswerChecked = "answer_checked"
const QuestionStarted = "question_started"
const ScoreUpdated = "score_updated"
const QuizEnded = "quiz_ended"
const QuizData = "quiz_data"
const Error = "quiz_error"

const JoinQuizErrorType = "JoinQuizError"

const serverPort = 8081
const username = prompt("Enter username: ")

const socket = io(`http://localhost:${serverPort}`);

socket.on('connect', () => {
    console.log(`Connected to server with socket ID: ${socket.id}`);
    const quizId = prompt("Enter quiz ID: ")
    socket.emit(JoinQuiz, username, Number(quizId))
});

// Handle disconnection
socket.on('disconnect', () => {
    console.log('Disconnected from server');
});

// Handle connection errors
socket.on('connect_error', (error) => {
    console.error('Connection error:', error);
});

let quizData
socket.on(QuizData, (message) => {
    console.log('The quiz is starting...');
    quizData = message
});

socket.on(Error, (message) => {
    console.error(message)
    if (message.startsWith(JoinQuizErrorType)) {
        const quizId = prompt("Enter quiz ID: ")
        if (quizId === 'quit') {
            process.exit(0)
        }
        socket.emit(JoinQuiz, username, Number(quizId))
    }
});

socket.on(QuestionStarted, (currentQuestionIndex, leaderboard) => {
    if (currentQuestionIndex > 0) {
        console.log(`The time for question ${currentQuestionIndex} is up!`)
        console.log('')
    }
    if (!quizData) {
        console.error("quiz data is empty")
        return
    }
    const qc = quizData.questions[currentQuestionIndex].content
    console.log(qc)
    const answerStr = prompt("Your answer: ")
    const answer = Number(answerStr)
    socket.emit(AnswerQuestion, JSON.stringify({
        quiz_id: quizData.id,
        question_index: Number(currentQuestionIndex),
        answer_index: answer - 1
    }))
})

socket.on(ScoreUpdated, (answeredUsername, leaderboard) => {
    console.log(`User ${answeredUsername} answered correctly!`)
    printLeaderboard(leaderboard)
})

socket.on(AnswerChecked, (correctAnswerIndex, newScore) => {
    console.log('Correct answer is:', correctAnswerIndex + 1);
    console.log('Your current score is:', newScore)
    console.log()
})

socket.on(QuizEnded, (leaderboard) => {
    console.log('The quiz has ended.')
    printLeaderboard(leaderboard)
    const quizId = prompt("Enter quiz ID: ")
    socket.emit(JoinQuiz, username, Number(quizId))
})
