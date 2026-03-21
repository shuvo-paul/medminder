FROM golang:1.25

WORKDIR /app

ENV GO111MODULE=on

RUN go install github.com/air-verse/air@latest \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends make postgresql-client curl \
	&& curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - \
	&& apt-get install -y --no-install-recommends nodejs \
	&& rm -rf /var/lib/apt/lists/* \
	&& corepack enable

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080 5173

CMD ["make", "start"]
