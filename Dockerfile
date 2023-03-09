FROM ubuntu:23.04

RUN apt-get update -y && apt-get install -y \
    ca-certificates \
    git \
    build-essential \
    meson ninja-build \
    golang \
    ffmpeg
#   rubberband-cli

## Installing sleef.
#RUN apt-get update -y && apt-get install -y \
#    libmpfr-dev  \
#    libssl-dev  \
#    libfftw3-dev
#
#RUN git clone https://github.com/shibatch/sleef.git
#
#WORKDIR /sleef
#
#RUN mkdir build
#WORKDIR /sleef/build
#
#RUN cmake -DCMAKE_INSTALL_PREFIX=/usr -DBUILD_DFT=TRUE ..
#
#RUN make
#RUN make test
#RUN make install

# Installing rubberband CLI.
RUN apt-get update -y && apt-get install -y \
    libsndfile1-dev \
    libsamplerate0-dev

WORKDIR /

RUN git clone https://github.com/breakfastquay/rubberband.git

WORKDIR /rubberband

RUN meson setup build  \
    -Dauto_features=disabled  \
    -Dcmdline=enabled \
    -Dfft=builtin \
    -Dresampler=libsamplerate

RUN ninja -C build
RUN ninja -C build install

RUN rubberband --version

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./ ./

#ENV GOOS=linux
#ENV GOARCH=arm
#ENV GODEBUG=tls13=0

RUN go build -buildvcs=false -o /scala-bot

CMD [ "/scala-bot" ]
