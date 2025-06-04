FROM golang:1.23-alpine

WORKDIR .

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o quiz .

EXPOSE 8081

CMD ["./quiz"]
