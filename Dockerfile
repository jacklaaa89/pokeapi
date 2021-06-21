ARG GOLANG_VERSION=1.16.5
ARG ALPINE_VERSION=3.13

# make a reference to the cached-layer so we can use it in subsequent
# build targets.
FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION}

# install dependencies
RUN apk --update add git bash gcc musl-dev openssh-client && \
  go get -u golang.org/x/tools/go/loader && \
  go get -u golang.org/x/tools/imports

# set the workdir, this has to be outside to GOPATH
# so modules work correctly.
WORKDIR /project

# Copy source code to the container.
COPY . .

RUN chmod +x ./start.sh

# run the tests as part of the build.
RUN CGO_ENABLED=0 go test ./... -cover

# build a binary & move to /var
RUN CGO_ENABLED=0 go build -o api main.go

# expose initial ports
EXPOSE 5555

# run the binary.
CMD ./start.sh
