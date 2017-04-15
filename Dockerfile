FROM golang
ADD . /go/src/github.com/tobyjsullivan/ocs-driver-notify
RUN  go install github.com/tobyjsullivan/ocs-driver-notify
CMD /go/bin/ocs-driver-notify
