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
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates gosu && rm -rf /var/lib/apt/lists/*
# patched version for http2:
RUN pip install --no-cache-dir git+https://github.com/r1tch/instaloader.git@using_httpx_for_http2
COPY --from=builder /out/gmb-bot /usr/local/bin/gmb-bot
COPY scripts/instagram /app/scripts/instagram
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x scripts/instagram/fetch_posts.py
RUN chmod +x /usr/local/bin/entrypoint.sh
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["/usr/local/bin/gmb-bot"]
