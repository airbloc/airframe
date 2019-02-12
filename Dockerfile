FROM golang:1.11

# environment variables here
ENV PATH /go/bin:$PATH
ENV AB_PROFILE production

# Copy local package files to the container's workspace.
WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build

EXPOSE 8080
ENTRYPOINT [ "/app/build/app" ]
