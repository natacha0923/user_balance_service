FROM golang:1.13.1
ENV SRC /go_projects/user_balance_service
ADD . ${SRC}
WORKDIR ${SRC}
RUN go build -v
CMD ./user_balance_service
