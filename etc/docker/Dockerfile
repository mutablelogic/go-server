ARG OS
ARG ARCH

# Run makefile to build all the commands
FROM --platform=${OS}/${ARCH} golang:latest AS builder
ARG OS
ARG ARCH
WORKDIR /usr/src/app
COPY . .

# Build the server
RUN \
    apt update -y && apt upgrade -y && \
    OS=${OS} ARCH=${ARCH} make build-docker 

# Copy binaries to /usr/local/bin
FROM --platform=${OS}/${ARCH} debian:bookworm-slim
ARG OS
ARG ARCH
ARG SOURCE
COPY --from=builder /usr/src/app/build/* /usr/local/bin/
RUN apt update -y && apt install -y ca-certificates 

# Labels
LABEL org.opencontainers.image.source=https://${SOURCE}

# Entrypoint when running the server
ENV PLUGIN_PATH="/usr/local/bin/*.plugin" \
    ENDPOINT="http://:80/"
EXPOSE 80 443
ENTRYPOINT [ "/usr/local/bin/server" ]
