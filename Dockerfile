FROM golang:1.13 as build

ARG OPTS

WORKDIR /buildspace
COPY go.mod .
COPY go.sum .

ENV GO111MODULE=on
#ENV GOPROXY="https://proxy.golang.org"
RUN go mod download

COPY . .

RUN VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///') && \
  GIT_COMMIT=$(git rev-list -1 HEAD) && \
  env ${OPTS} CGO_ENABLED=0 GOOS=linux \
  go build -o jeopardy-nodeselector -v \
  -ldflags "-s -w \
    -X github.com/drnic/jeopardy-nodeselector/pkg/version.Release=${VERSION} \
    -X github.com/drnic/jeopardy-nodeselector/pkg/version.SHA=${GIT_COMMIT}" \
  ./cmd

FROM alpine AS app
COPY --from=build /buildspace/jeopardy-nodeselector /usr/bin/jeopardy-nodeselector
EXPOSE 8443
VOLUME [ "/certs" ]

CMD ["/usr/bin/jeopardy-nodeselector", "-cert-path", "/certs/cert.crt", "-key-path", "/certs/key.key", "-port", "8443"]
