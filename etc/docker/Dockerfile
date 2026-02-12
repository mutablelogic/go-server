ARG OS
ARG ARCH

# Build stage
FROM --platform=${OS}/${ARCH} golang:1.25 AS builder
ARG OS
ARG ARCH
WORKDIR /usr/src/app
COPY . .

RUN OS=${OS} ARCH=${ARCH} make build

# Runtime stage
FROM --platform=${OS}/${ARCH} debian:trixie-slim
ARG SOURCE
RUN apt-get update -y && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/src/app/build/kaiak /usr/local/bin/kaiak

# Labels
LABEL org.opencontainers.image.source=https://${SOURCE}

# Entrypoint
EXPOSE 80 443
ENTRYPOINT ["/usr/local/bin/kaiak"]
CMD ["run"]
