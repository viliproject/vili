FROM quay.io/airware/vilibase:20160630-160722

ENV GO15VENDOREXPERIMENT=1

COPY . /go/src/github.com/airware/vili
WORKDIR /go/src/github.com/airware/vili

RUN npm install
RUN npn run build

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' -o main
