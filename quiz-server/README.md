# Realtime Quiz

## [Design document](https://docs.google.com/document/d/1oGQicC4gW5SavF1b3wWo4N7JPqhH1lnZDIbWj-uPvGs/edit?usp=sharing)

## How to run
1. Start kafka broker (port 9092)
2. Start redis (port 6379)
3. Build binary `go build ./`
4. Start server binary `PORT=8080 ./quiz`
5. Start quiz client
```
npm i
node index.js
```
6. Connect the client with the server: specify port & user ID
7. Start the quiz `curl localhost:1918/start/[quiz ID]`
8. Enter the quiz ID in the client