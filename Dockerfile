# Use the offical Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.12 as builder

ENV GO112MODULE=on
# Copy local code to the container image.
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

# Build the command inside the container.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o main

# Use a Docker multi-stage build to create a lean production image.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine
RUN apk add --no-cache ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/main /main

# Run the web service on container startup.
CMD ["/main"]
