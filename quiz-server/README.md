# Realtime Quiz

## [Design document](https://docs.google.com/document/d/1oGQicC4gW5SavF1b3wWo4N7JPqhH1lnZDIbWj-uPvGs/edit?usp=sharing)

## How to run
1. Start kafka broker (port 9092)
2. Start redis (port 6379)
3. Start temporal (port 7233)
4. Build server binary `cd quiz-server && go build -o ../ . && cd ..`
5. Build temporal worker `cd quiz-server/workflow/worker && go build -o ../../../ . && cd ../../../`
6. Start server binary `PORT=8081 ./quiz`
7. Start temporal worker binary `./worker`
8. Start the quiz client `cd quiz-client && npm i && npm start`
9. Start the quiz `curl localhost:8081/start/[quiz ID]`
10. Enter username and quiz ID in the client to join the quiz