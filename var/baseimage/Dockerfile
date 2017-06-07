FROM mhart/alpine-node:7.10.0

RUN apk add --update go \
                     musl-dev \
                     make \
                     git \
                     ca-certificates \
                     xmlsec \
        && rm -rf /var/cache/apk/* \
        && rm -rf /usr/lib/go/bin/windows_* \
        && rm -rf /usr/lib/go/bin/darwin_* \
        && rm -rf /usr/lib/go/bin/freebsd_*

# fix xmlsec-openssl
RUN ln -s libxmlsec1-openssl.so.1 /usr/lib/libxmlsec1-openssl.so

# set up go
RUN mkdir -p /go
ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

ENV GO15VENDOREXPERIMENT 1
RUN go get github.com/golang/lint/golint
RUN go get golang.org/x/tools/cmd/cover
