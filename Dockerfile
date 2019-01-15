FROM alpine:3.2
ADD raft /go/bin/raft
ENTRYPOINT /go/bin/raft
