FROM golang:1.11

# environment variables here
ENV PATH /go/bin:$PATH
ENV AB_PROFILE production

# Copy local package files to the container's workspace.
WORKDIR /airframe
COPY . .

RUN make build

EXPOSE 8080
EXPOSE 9090
ENTRYPOINT [ "/airframe/build/airframe" ]
