FROM golang:1.22-bookworm AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gmb-bot ./cmd/gmb-bot

FROM node:20-bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates chromium && rm -rf /var/lib/apt/lists/*
ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright
COPY package.json ./
RUN npm install --omit=dev && npx playwright install chromium
COPY --from=builder /out/gmb-bot /usr/local/bin/gmb-bot
COPY . .
RUN chmod +x scripts/tiktok/fetch_videos.mjs
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/gmb-bot"]
