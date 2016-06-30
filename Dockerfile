FROM quay.io/airware/vilibase:20160630-160722

ENV GO15VENDOREXPERIMENT=1

COPY . /go/src/github.com/airware/vili
WORKDIR /go/src/github.com/airware/vili

RUN cd public && npm install
RUN cd public && node node_modules/gulp/bin/gulp.js webpack

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' -o main
