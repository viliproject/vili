FROM quay.io/airware/vilibase:20170606-173659

ENV GO15VENDOREXPERIMENT=1

COPY . /go/src/github.com/airware/vili
WORKDIR /go/src/github.com/airware/vili

RUN npm install
RUN npm run build

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' -o main
