FROM golang:1.24.1 AS build

ENV CGO_ENABLED=0

RUN curl -fsSL https://deb.nodesource.com/setup_23.x | bash - && \
	apt-get update && \
	apt-get install -y nodejs && \
	npm install --global pnpm

RUN mkdir -p /app \
	&& adduser --system --group user \
	&& chown -R user:user /app

WORKDIR /app

COPY go.* .

RUN go mod download

COPY . .

RUN go generate ./... && \
	CGO_ENABLED=0 go build -ldflags "-s -w" -o /app/build/hubble ./cmd/hubble

FROM gcr.io/distroless/static-debian11

COPY --from=build /app/build/hubble hubble

EXPOSE 3288

CMD ["/hubble"]
