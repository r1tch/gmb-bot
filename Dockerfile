FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS builder
WORKDIR /src
ARG TARGETOS
ARG TARGETARCH
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/gmb-bot ./cmd/gmb-bot

FROM python:3.12-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
RUN pip install --no-cache-dir instaloader
COPY --from=builder /out/gmb-bot /usr/local/bin/gmb-bot
COPY scripts/instagram /app/scripts/instagram
RUN chmod +x scripts/instagram/fetch_posts.py
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/gmb-bot"]
