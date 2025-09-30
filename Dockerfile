# Use Go base image
FROM golang:1.23-bullseye AS builder

# Install go-face dependencies + ccache
RUN apt-get update && apt-get -y install \
    libdlib-dev \
    libblas-dev \
    libatlas-base-dev \
    liblapack-dev \
    libjpeg62-turbo-dev \
    ccache \
 && rm -rf /var/lib/apt/lists/*

# Ensure ccache is useds
ENV PATH="/usr/lib/ccache:$PATH"

# Set working directory
WORKDIR /app

# Copy only go.mod & go.sum first (better caching)
COPY go.mod go.sum ./

# Download modules with cache
RUN --mount=type=cache,target=/go/pkg \
    go mod download

# Copy project files
COPY . .

# Build app with caching enabled
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=1 go build -o main .

# Final runtime stage (slim image)
FROM debian:bullseye-slim AS runtime

WORKDIR /app

# Install runtime deps (needed by go-face binary)
RUN apt-get update && apt-get install -y \
    libdlib19 \
    libblas3 \
    libatlas3-base \
    liblapack3 \
    libjpeg62-turbo \
 && rm -rf /var/lib/apt/lists/*

# Copy only built binary + models
COPY --from=builder /app/main .
COPY --from=builder /app/api/db/migrations ./api/db/migrations
COPY --from=builder /app/models ./models
COPY --from=builder /app/images ./images

# Run the binary
CMD ["./main"]
