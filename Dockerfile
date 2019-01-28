# Modified from https://medium.com/@pierreprinetti/the-go-1-11-dockerfile-a3218319d191
FROM golang:1.11-alpine AS builder

# Install the Certificate-Authority certificates for the app to be able to make
# calls to HTTPS endpoints.
# Git is required for fetching the dependencies.
RUN apk add --no-cache ca-certificates git

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /src

# Fetch dependencies first; they are less susceptible to change on every build
# and will therefore be cached for speeding up the next build
COPY ./go.mod ./go.sum ./
RUN go mod download

# Import the code from the context.
COPY ./ ./

# Build the executable to `/app`. Mark the build as statically linked.
RUN CGO_ENABLED=0 go build \
    -installsuffix 'static' \
    -o /app .

# Final stage: the running container.
FROM alpine AS final

RUN apk add --no-cache tzdata ca-certificates

# Import the compiled executable from the first stage.
COPY --from=builder /app /app

# Run the compiled binary.
ENTRYPOINT ["/app"]