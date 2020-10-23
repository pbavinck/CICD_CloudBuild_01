# Step 1: Compile the application
FROM golang:1.13 as builder

WORKDIR /app

# Retrieve application dependencies.
COPY go.* ./
RUN go mod download

# Copy local code
COPY . ./

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o server

# Step 2: Create container with compiled application
# FROM debian:stable-slim
# RUN apt-get update -qq &&\
#     apt-get -qq install -qqy ca-certificates

FROM alpine:3
RUN apk add --no-cache ca-certificates

# Copy application from Step 1
COPY --from=builder /app/server /server

# Copy static files from source
WORKDIR /app/static
COPY ./static .

# Run the web service on container startup
CMD ["/server"]