import React, {useEffect, useState} from 'react';
import './App.css';
import {Button, Card, Col, Form, FormProps, Input, InputNumber, Layout, message, Progress, Rate, Row} from "antd";
const { Content, Footer, Header } = Layout;
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
    username?: string;
    quizId?: string;
};

type AnswerFieldType = {
    answer?: string
}

function App() {
    let [quizData, setQuizData] = useState<any>()
    let [currentQuestionIndex, setCurrentQuestionIndex] = useState(-1)
    let [leaderboard, setLeaderboard] = useState<any[]>()
    let [timeLeft, setTimeLeft] = useState(defaultQuestionTime)

    useEffect(() => {
        function onConnect() {
            console.log(`Connected to server with socket ID: ${socket.id}`);
        }

        function onDisconnect() {
            console.log('Disconnected from server');
        }

        const connectionError = (error: any) => {
            console.error('Connection error:', error);
        }

        const onQuizStarted = (quizData: any) => {
            setQuizData(quizData)
        }

        const onQuizError = (err: any) => {
            message.error(err)
        }

        const onQuestionStarted = (currentQuestionIndex: any, leaderboard: any) => {
            if (currentQuestionIndex > 0) {
                message.warning(`The time for question ${currentQuestionIndex} is up!`)
            }
            setCurrentQuestionIndex(+currentQuestionIndex)
            setLeaderboard(leaderboard)
            setTimeLeft(defaultQuestionTime)
            const id = setInterval(() => {
                setTimeLeft(prevState => {
                    return prevState - 1
                })
            }, 1000)
            setTimeout(() => {
                clearInterval(id);
            }, defaultQuestionTime * 1000);
        }

        const onScoreUpdated = (answeredUsername: any, leaderboard: any) => {
            message.success(`User ${answeredUsername} answered correctly!`)
            setLeaderboard(leaderboard)
        }

        const onAnswerChecked = (correctAnswerIndex: any, newScore: any) => {
            message.info(`Correct answer is: ${correctAnswerIndex + 1}. Your current score is: ${newScore}`,);
        }

        const onQuizEnded = (leaderboard: any) => {
            message.info('The quiz has ended.')
            setTimeLeft(0)
            setTimeout(() => {
                setLeaderboard([])
                setQuizData(null)
                setCurrentQuestionIndex(-1)
            }, 3000)
        }

        socket.on('connect', onConnect);
        socket.on('disconnect', onDisconnect);
        socket.on('connect_error', connectionError);
        socket.on(QuizData, onQuizStarted);
        socket.on(Error, onQuizError);
        socket.on(QuestionStarted, onQuestionStarted)
        socket.on(ScoreUpdated, onScoreUpdated)
        socket.on(AnswerChecked, onAnswerChecked)
        socket.on(QuizEnded, onQuizEnded)

        // Cleanup on component unmount
        return () => {
            socket.off('connect', onConnect);
            socket.off('disconnect', onDisconnect);
            socket.off('connect_error', connectionError);
            socket.off(QuizData, onQuizStarted);
            socket.off(Error, onQuizError);
            socket.off(QuestionStarted, onQuestionStarted)
            socket.off(ScoreUpdated, onScoreUpdated)
            socket.off(AnswerChecked, onAnswerChecked)
            socket.off(QuizEnded, onQuizEnded)
        };
    }, []);

    const onFinish: FormProps<FieldType>['onFinish'] = (values) => {
        socket.emit(JoinQuiz, values.username, values.quizId);
    }

    const onSubmitAnswer: FormProps<AnswerFieldType>['onFinish'] = (values) => {
        socket.emit(AnswerQuestion, JSON.stringify({
            quiz_id: quizData.id,
            question_index: Number(currentQuestionIndex),
            answer_index: Number(values.answer) - 1,
        }))
    }

    return (
        <Layout style={{
            height: '100vh'
        }}>
            <Header style={{
                textAlign: 'center',
                color: '#fff',
            }}>
                <h1>HungCQ & ThuyDB's Real-Time Quiz</h1>
            </Header>
            <Content style={{flex: 1, overflow: "auto"}}>
                <Row style={{marginTop: '3%'}}>
                    <Col span={4}></Col>
                    {quizData ?
                        <>
                            <Col span={8}>
                                <Card
                                    title={currentQuestionIndex >= 0 ? `Question ${currentQuestionIndex + 1}` : 'Loading question...'}
                                    style={{width: '100%', whiteSpace: "pre-wrap"}}>
                                    {currentQuestionIndex >= 0 &&
                                        <>
                                            <Progress status='normal' percent={timeLeft * 10}
                                                      format={(percent) => `${(percent || 0) / 10}s`}/>
                                            <p>{quizData.questions[currentQuestionIndex].content}</p>
                                            <Form
                                                name="basic"
                                                initialValues={{remember: true}}
                                                onFinish={onSubmitAnswer}
                                                autoComplete="off"
                                                layout={"vertical"}
                                            >

                                                <Form.Item<AnswerFieldType>
                                                    label="Your answer"
                                                    name="answer"
                                                    rules={[{required: true, message: 'Please input answer!'}]}
                                                >
                                                    <InputNumber max={4} min={1}/>
                                                </Form.Item>

                                                <Form.Item>
                                                    <Button type="primary" htmlType="submit">
                                                        Submit
                                                    </Button>
                                                </Form.Item>
                                            </Form>
                                        </>
                                    }
                                </Card>
                            </Col>
                            <Col span={8}>
                                <Card title="Leaderboard" style={{width: '100%'}}>
                                    {leaderboard && leaderboard.map((item: any) =>
                                        <h3>{`${item.username} `}<Rate count={item.score} value={item.score}/></h3>
                                    )
                                    }
                                </Card>
                            </Col>
                        </>
                        :
                        <Col span={16}>
                            <Form
                                name="basic"
                                labelCol={{span: 6}}
                                wrapperCol={{span: 10}}
                                style={{maxWidth: '100%', marginTop: '3%'}}
                                initialValues={{remember: true}}
                                onFinish={onFinish}
                                autoComplete="off"
                            >
                                <Form.Item<FieldType>
                                    label="Username"
                                    name="username"
                                    rules={[{required: true, message: 'Please input your user ID!'}]}
                                >
                                    <Input/>
                                </Form.Item>

                                <Form.Item<FieldType>
                                    label="Quiz ID"
                                    name="quizId"
                                    rules={[{required: true, message: 'Please input quiz ID!'}]}
                                >
                                    <InputNumber style={{ width: '100%' }}/>
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
            </Content>
            <Footer style={{
                textAlign: 'center',
                color: 'grey'
            }}>
                <h3>Copyright © 2024 HungCQ & ThuyDB</h3>
            </Footer>
        </Layout>
    );
}

export default App;
