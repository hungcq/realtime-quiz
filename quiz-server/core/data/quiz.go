package data

import "quiz/core/models"

var QuizData = map[models.QuizId]*models.Quiz{
	programmingQuiz.Id:       programmingQuiz,
	elementaryEnglishQuiz.Id: elementaryEnglishQuiz,
}

var programmingQuiz = &models.Quiz{
	Id: 1,
	Questions: []models.Question{
		{
			Content: `What is "polymorphism" in object-oriented programming?
1) The process of breaking a program into smaller modules.
2) The ability of a single function or class to operate on different data types.
3) The method of hiding implementation details from the user.
4) A technique to establish relationships between two classes.`,
			CorrectAnswerIndex: 1,
		},
		{
			Content: `What does the term "idempotent" mean in programming?
1) An operation that can be applied multiple times without changing the result beyond the initial application.
2) A function that depends only on its arguments and produces the same output every time.
3) A program that completes a task only once regardless of the input.
4) An algorithm that guarantees no duplicate results in a dataset.`,
			CorrectAnswerIndex: 0,
		},
		{
			Content: `In databases, what does "ACID" stand for?
1) Automatic, Consistent, Immediate, Durable
2) Atomicity, Consistency, Isolation, Durability
3) Asynchronous, Concurrent, Immediate, Durable
4) Availability, Compliance, Integrity, Design`,
			CorrectAnswerIndex: 1,
		},
		{
			Content: `What is a "deadlock" in concurrent programming?
1) A situation where two processes wait indefinitely for each other to release resources.
2) A mechanism that ensures tasks are executed in order.
3) A state where one process halts execution due to memory shortage.
4) A technique to prioritize tasks based on urgency.`,
			CorrectAnswerIndex: 0,
		},
		{
			Content: `What is "inheritance" in object-oriented programming?
1) Encapsulation of data in classes.
2) The ability to reuse methods and properties from one class in another class.
3) A method for enforcing access control in classes.
4) A feature to create anonymous functions.`,
			CorrectAnswerIndex: 1,
		},
	},
}

var elementaryEnglishQuiz = &models.Quiz{
	Id: 2,
	Questions: []models.Question{
		{
			Content: `What is the plural form of the word "child"?
1) Childs
2) Childrens
3) Children
4) Childer`,
			CorrectAnswerIndex: 2,
		},
		{
			Content: `Which of these is a synonym for "happy"?
1) Sad
2) Joyful
3) Angry
4) Tired`,
			CorrectAnswerIndex: 1,
		},
		{
			Content: `What is the correct article to use before the word "apple"?
1) A
2) An
3) The
4) None`,
			CorrectAnswerIndex: 1,
		},
		{
			Content: `Which sentence is grammatically correct?
1) She don’t like apples.
2) She doesn’t likes apples.
3) She doesn’t like apples.
4) She don’t likes apples.`,
			CorrectAnswerIndex: 2,
		},
		{
			Content: `What is the past tense of the verb "run"?
1) Runs
2) Running
3) Ran
4) Runned`,
			CorrectAnswerIndex: 2,
		},
	},
}
