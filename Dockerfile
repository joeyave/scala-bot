# syntax=docker/dockerfile:1.7

FROM node:20-bookworm-slim AS frontend
WORKDIR /app/webapp-react

COPY webapp-react/package*.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

COPY webapp-react/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS backend
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/scala-bot .

FROM ubuntu:24.04 AS rubberband
ARG DEBIAN_FRONTEND=noninteractive
ARG RUBBERBAND_VERSION=v4.0.0

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    git \
    libsamplerate0-dev \
    libsndfile1-dev \
    meson \
    ninja-build \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

RUN git clone --branch ${RUBBERBAND_VERSION} --depth 1 https://github.com/breakfastquay/rubberband.git

WORKDIR /src/rubberband

RUN meson setup build \
    --buildtype=release \
    -Dauto_features=disabled \
    -Dcmdline=enabled \
    -Dfft=builtin \
    -Dresampler=libsamplerate

RUN ninja -C build

FROM ubuntu:24.04
ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    ffmpeg \
    libsamplerate0 \
    libsndfile1 \
    libstdc++6 \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=backend /out/scala-bot /usr/local/bin/scala-bot
COPY --from=rubberband /src/rubberband/build/rubberband /usr/local/bin/rubberband
COPY --from=frontend /app/webapp-react/dist ./webapp-react/dist

EXPOSE 8080

CMD ["/usr/local/bin/scala-bot"]
