FROM golang:1.9.1-alpine3.6

RUN apk add -U --no-cache \
    musl-dev \
    make \
    git \
    ca-certificates \
    xmlsec \
    nodejs-dev \
    nodejs-npm

RUN go get \
    github.com/golang/lint/golint \
    golang.org/x/tools/cmd/cover

WORKDIR /go/src/github.com/airware/vili/

# run npm install first for dependencies
COPY package.json /go/src/github.com/airware/vili/
RUN npm install

# then copy the rest of the app and install
COPY . /go/src/github.com/airware/vili/

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' -o main

RUN npm run build

# second stage, just have the compiled binary
FROM alpine:3.6

RUN apk --no-cache add curl ca-certificates xmlsec && update-ca-certificates

# Install kubectl
RUN curl -L https://storage.googleapis.com/kubernetes-release/release/v1.8.1/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl

WORKDIR /app/

COPY --from=0 /go/src/github.com/airware/vili/main .
COPY --from=0 /go/src/github.com/airware/vili/public/build build

ENV HOME /app
ENV NODE_ENV production
ENV BUILD_DIR /app/build

EXPOSE 80
ENTRYPOINT ["/app/main"]
