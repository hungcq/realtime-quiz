import React, {useEffect, useMemo, useState} from 'react';
import './App.css';
import {Alert, Button, Card, Col, Form, FormProps, Input, Layout, Row} from "antd";
import {Footer, Header} from "antd/es/layout/layout";
import {socket} from "./socket";

const defaultQuestionTime = 10 // seconds
const JoinQuiz = "join_quiz"
const AnswerQuestion = "answer_question"

const AnswerChecked = "answer_checked"
const QuestionStarted = "question_started"
const ScoreUpdated = "score_updated"
const QuizEnded = "quiz_ended"
const QuizData = "quiz_data"
const Error = "quiz_error"

type FieldType = {
    userId?: string;
    quizId?: string;
};


type Quiz = {}

function App() {
    const [isConnected, setIsConnected] = useState(socket.connected);
    let [quizData, setQuizData] = useState<any>()
    let [userId, setUserId] = useState(0)
    let [quizId, setQuizId] = useState(0)
    let [error, setError] = useState("")
    let leaderboard = useState()

    useEffect(() => {
        function onConnect() {
            console.log(`Connected to server with socket ID: ${socket.id}`);
            setIsConnected(true);
            socket.emit('join_quiz', 'abc')
        }

        function onDisconnect() {
            console.log('Disconnected from server');
            setIsConnected(false);
        }

        // Listen for the 'welcome' event
        socket.on('connect', onConnect);

        // Handle disconnection
        socket.on('disconnect', onDisconnect);

        // Handle connection errors
        socket.on('connect_error', (error) => {
            console.error('Connection error:', error);
        });

        socket.on(QuizData, (message) => {
            console.log('The quiz is starting...');
            quizData = message
        });

        socket.on(Error, (message) => {
            console.log(message)
            setError(message)
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

        socket.on(ScoreUpdated, (answeredUserId, leaderboard) => {
            console.log(`User ${answeredUserId} answered correctly!`)
            // printLeaderboard(leaderboard)
        })

        socket.on(AnswerChecked, (correctAnswerIndex, newScore) => {
            console.log('Correct answer is:', correctAnswerIndex + 1);
            console.log('Your current score is:', newScore)
            console.log()
        })

        socket.on(QuizEnded, (leaderboard) => {
            console.log('The quiz has ended.')
            // printLeaderboard(leaderboard)
            const quizId = prompt("Enter quiz ID: ")
            socket.emit(JoinQuiz, userId, quizId)
        })
        // Cleanup on component unmount
        return () => {
            socket.off('connect', onConnect);
            socket.off('disconnect', onDisconnect);
        };
    }, []);

    const onFinish: FormProps<FieldType>['onFinish'] = (values) => {
        setUserId(+values.userId!)
        setQuizId(+values.quizId!)
        socket.emit('join_quiz', values.userId, values.quizId);
    }

    return (
        <Layout>
            <Header/>
            {error && <Alert message={error} type="error" />}
            <Row>
                <Col span={4}></Col>
                {quizData ?
                    <>
                        <Col span={8}>
                            <Card title="Question" extra={`Time left: ${1}s`} style={{width: '80%'}}>
                                <p>Card content</p>
                                <p>Card content</p>
                                <p>Card content</p>
                            </Card>
                        </Col>
                        <Col span={8}>
                            <Card title="Leaderboard" style={{width: '80%'}}>
                                <p>Card content</p>
                                <p>Card content</p>
                                <p>Card content</p>
                            </Card>
                        </Col>
                    </>
                    :
                    <Col span={16}>
                        <Form
                            name="basic"
                            labelCol={{span: 6}}
                            wrapperCol={{span: 10}}
                            style={{maxWidth: '100%'}}
                            initialValues={{remember: true}}
                            onFinish={onFinish}
                            autoComplete="off"
                        >
                            <Form.Item<FieldType>
                                label="User ID"
                                name="userId"
                                rules={[{required: true, message: 'Please input your user ID!'}]}
                            >
                                <Input/>
                            </Form.Item>

                            <Form.Item<FieldType>
                                label="Quiz ID"
                                name="quizId"
                                rules={[{required: true, message: 'Please input quiz ID!'}]}
                            >
                                <Input/>
                            </Form.Item>

                            <Form.Item label={null}>
                                <Button type="primary" htmlType="submit">
                                    Submit
                                </Button>
                            </Form.Item>
                        </Form>
                    </Col>
                }
                <Col span={4}></Col>
            </Row>
            <Footer/>
        </Layout>
    );
}

export default App;
