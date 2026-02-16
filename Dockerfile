FROM golang:1.25

WORKDIR /app

ENV GO111MODULE=on

RUN go install github.com/air-verse/air@latest \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends make postgresql-client \
	&& rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["make", "dev"]
