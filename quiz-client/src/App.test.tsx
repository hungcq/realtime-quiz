/// <reference types="jest" />
const React = require('react');
const { render, screen, fireEvent, act } = require('@testing-library/react');
require('@testing-library/jest-dom');
const App = require('./App').default;
const { socket } = require('./socket');

// Mock matchMedia
const matchMediaMock = () => ({
  matches: false,
  media: '',
  onchange: null,
  addListener: jest.fn(),
  removeListener: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
  dispatchEvent: jest.fn(),
});

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: matchMediaMock,
});

// Ensure matchMedia is available globally
global.matchMedia = matchMediaMock;

type MockSocket = {
  on: jest.Mock;
  off: jest.Mock;
  emit: jest.Mock;
  id: string;
};

// Mock socket
jest.mock('./socket', () => ({
  socket: {
    on: jest.fn(),
    off: jest.fn(),
    emit: jest.fn(),
    id: 'test-socket-id'
  } as MockSocket
}));

describe('App Component', () => {
  beforeEach(() => {
    // Clear all mocks before each test
    jest.clearAllMocks();
  });

  test('renders join quiz form initially', () => {
    render(<App />);
    
    // Check for form elements
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/quiz id/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /submit/i })).toBeInTheDocument();
  });

  test('submits join quiz form with correct data', async () => {
    render(<App />);
    
    // Fill in the form
    await act(async () => {
      fireEvent.change(screen.getByLabelText(/username/i), { target: { value: 'testUser' } });
      fireEvent.change(screen.getByLabelText(/quiz id/i), { target: { value: '123' } });
      
      // Submit the form
      fireEvent.click(screen.getByRole('button', { name: /submit/i }));
    });
    
    // Wait for next tick to ensure all state updates are processed
    await act(async () => {
      await Promise.resolve();
    });
    
    // Verify socket emit was called with correct data
    expect(socket.emit).toHaveBeenCalledWith('join_quiz', 'testUser', 123);
  });

  test('displays quiz content when quiz data is received', () => {
    render(<App />);
    
    // Simulate receiving quiz data
    act(() => {
      const mockQuizData = {
        id: '123',
        questions: [
          { content: 'Test question 1' },
          { content: 'Test question 2' }
        ]
      };
      const quizDataCallback = (socket.on as jest.Mock).mock.calls.find((call: [string, Function]) => call[0] === 'quiz_data')?.[1];
      if (quizDataCallback) {
        quizDataCallback(mockQuizData);
      }
    });

    // Check if quiz content is displayed
    expect(screen.getByText(/loading question/i)).toBeInTheDocument();
  });

  test('displays leaderboard when data is received', async () => {
    render(<App />);
    
    // First simulate receiving quiz data
    act(() => {
      const mockQuizData = {
        id: '123',
        questions: [{ content: 'Test question 1' }]
      };
      const quizDataCallback = (socket.on as jest.Mock).mock.calls.find((call: [string, Function]) => call[0] === 'quiz_data')?.[1];
      if (quizDataCallback) {
        quizDataCallback(mockQuizData);
      }
    });
    
    // Simulate receiving leaderboard data
    act(() => {
      const mockLeaderboard = [
        { username: 'user1', score: 3 },
        { username: 'user2', score: 2 }
      ];
      const questionStartedCallback = (socket.on as jest.Mock).mock.calls.find((call: [string, Function]) => call[0] === 'question_started')?.[1];
      if (questionStartedCallback) {
        questionStartedCallback(0, mockLeaderboard);
      }
    });

    // Wait for next tick to ensure all state updates are processed
    await act(async () => {
      await Promise.resolve();
    });
    
    // Check if leaderboard is displayed
    await screen.findByText(/user1/);
    await screen.findByText(/user2/);
  });

  test('submits answer when question is displayed', async () => {
    render(<App />);
    
    // First simulate receiving quiz data
    act(() => {
      const mockQuizData = {
        id: '123',
        questions: [
          { content: 'Test question 1' }
        ]
      };
      const quizDataCallback = (socket.on as jest.Mock).mock.calls.find((call: [string, Function]) => call[0] === 'quiz_data')?.[1];
      if (quizDataCallback) {
        quizDataCallback(mockQuizData);
      }
    });

    // Then simulate question started
    act(() => {
      const questionStartedCallback = (socket.on as jest.Mock).mock.calls.find((call: [string, Function]) => call[0] === 'question_started')?.[1];
      if (questionStartedCallback) {
        questionStartedCallback(0, []);
      }
    });

    // Fill and submit answer
    await act(async () => {
      fireEvent.change(screen.getByLabelText(/your answer/i), { target: { value: '1' } });
      fireEvent.click(screen.getByRole('button', { name: /submit/i }));
    });

    // Wait for next tick to ensure all state updates are processed
    await act(async () => {
      await Promise.resolve();
    });
    
    // Verify answer was submitted
    expect(socket.emit).toHaveBeenCalledWith(
      'answer_question',
      JSON.stringify({
        quiz_id: '123',
        question_index: 0,
        answer_index: 0
      })
    );
  });

  test('displays error message when quiz error occurs', async () => {
    render(<App />);
    
    // Simulate quiz error
    act(() => {
      const errorCallback = (socket.on as jest.Mock).mock.calls.find((call: [string, Function]) => call[0] === 'quiz_error')?.[1];
      if (errorCallback) {
        errorCallback('Test error message');
      }
    });
  
    // Check if error message is displayed
    await screen.findByText(/test error message/i);
  });

  test('cleans up socket listeners on unmount', () => {
    const { unmount } = render(<App />);
    
    // Unmount component
    unmount();
    
    // Verify all socket listeners were removed
    expect(socket.off).toHaveBeenCalledTimes(9); // Number of socket.on calls in the component
  });
});
