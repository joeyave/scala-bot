FROM golang:1.20

WORKDIR /app

RUN echo "deb http://deb.debian.org/debian sid main" | sudo tee -a /etc/apt/sources.list
RUN apt-get update && apt-get install -y rubberband-cli=3.1.2+dfsg0

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./ ./

#ENV GOOS=linux
#ENV GOARCH=arm

RUN go build -buildvcs=false -o /scala-bot

CMD [ "/scala-bot" ]
