FROM ubuntu:23.04

WORKDIR /app

RUN apt-get update && apt-get -y install \
    ca-certificates \
    golang \
    rubberband-cli \
    ffmpeg


COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./ ./

#ENV GOOS=linux
#ENV GOARCH=arm
#ENV GODEBUG=tls13=0

RUN go build -buildvcs=false -o /scala-bot

CMD [ "/scala-bot" ]
