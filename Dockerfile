FROM ubuntu:23.04

WORKDIR /app

RUN apt-get update && apt-get -y install \
    golang \
    rubberband-cli


COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./ ./

#ENV GOOS=linux
#ENV GOARCH=arm

RUN go build -buildvcs=false -o /scala-bot

CMD [ "/scala-bot" ]
