# Multi-stage build for Runbyte Server
# Stage 1: Build WASM sandbox
FROM node:20-alpine AS wasm-builder

WORKDIR /build/pkg/wasm

# Install extism-js compiler and binaryen
RUN apk add --no-cache curl bash sudo gcompat

RUN curl -O https://raw.githubusercontent.com/extism/js-pdk/main/install.sh

RUN bash install.sh

RUN rm install.sh

# Add extism-js to PATH
ENV PATH="/root/.extism/bin:${PATH}"

# Copy package files
COPY pkg/wasm/package*.json ./

# Install dependencies
RUN npm install

# Copy WASM source
COPY pkg/wasm/ ./

# Build WASM
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS go-builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy built WASM from previous stage
COPY --from=wasm-builder /build/pkg/wasm/dist/sandbox.wasm ./pkg/wasm/dist/sandbox.wasm

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o runbyte ./cmd/runbyte

# Stage 3: Runtime image
FROM node:20-alpine

# Install ca-certificates and rspack (required for runtime bundling)
RUN apk --no-cache add ca-certificates && \
    npm install -g @rspack/cli @rspack/core

WORKDIR /app

# Copy binary from builder
COPY --from=go-builder /build/runbyte .

# Copy example config
COPY runbyte.json.example ./runbyte.json.example

# Create directory for custom configs
RUN mkdir -p /etc/runbyte

# Expose default HTTP port
EXPOSE 3000

# Set environment variables
# ENV RUNBYTE_CONFIG=/etc/runbyte/config.json
ENV RUNBYTE_PORT=3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ || exit 1

# Default to HTTP mode
ENTRYPOINT ["/app/runbyte"]
CMD ["-transport", "http"]
