FROM arm64v8/golang:1.18

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./ ./

ENV GOOS=linux
ENV GOARCH=arm

RUN go build -o /scala-bot

CMD [ "/scala-bot" ]
