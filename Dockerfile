FROM golang:1.13 as build

LABEL repo="https://github.com/starkandwayne/jeopardy-nodeselector"
ARG GIT_COMMIT=""
LABEL commit=$GIT_COMMIT
ENV GIT_COMMIT=$GIT_COMMIT

WORKDIR /buildspace
COPY go.mod .
COPY go.sum .

ENV GO111MODULE=on
#ENV GOPROXY="https://proxy.golang.org"
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o jeopardy-nodeselector -v ./cmd

FROM alpine AS app
COPY --from=build /buildspace/jeopardy-nodeselector /usr/bin/jeopardy-nodeselector
EXPOSE 8443

CMD ["/usr/bin/jeopardy-nodeselector", "-cert-path", "/certs/cert.crt", "-key-path", "/certs/key.key", "-port", "8443"]
