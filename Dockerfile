# Setup build environment
FROM golang:alpine AS build-env
ADD . /go/src/github.com/maciekmm/uek-bruschetta
WORKDIR /go/src/github.com/maciekmm/uek-bruschetta
# Install all dependencies
RUN apk add --no-cache git
RUN go get -v ./...
# Build binary
RUN go build -o bruschetta

# Create minimal image
FROM alpine
# Copy sources + binary
# TODO: only copy assets and templates (?)
COPY --from=build-env /go/src/github.com/maciekmm/uek-bruschetta/ /app/
WORKDIR /app/
# Expose port
EXPOSE 3000
# Run the application
ENTRYPOINT ./bruschetta
