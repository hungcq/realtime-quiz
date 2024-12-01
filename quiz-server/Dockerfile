FROM golang:1.23

WORKDIR .

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o quiz .

EXPOSE 8081

CMD ["./quiz"]
