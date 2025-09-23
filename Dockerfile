FROM golang:1.22-bullseye

# Install go-face dependencies
RUN apt-get update && apt-get -y install \
    libdlib-dev \
    libblas-dev \
    libatlas-base-dev \
    liblapack-dev \
    libjpeg62-turbo-dev

# Set workdir
WORKDIR /app

# Copy go.mod and go.sum first (for caching)
COPY go.mod ./

RUN go mod download

# Copy source code
COPY . .

# Build app
RUN go build -o main .

# Run app
CMD ["./main"]