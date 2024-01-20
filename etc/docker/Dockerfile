ARG PLATFORM
ARG ARCH
FROM --platform=${PLATFORM}/${ARCH} golang:1.21-bullseye AS builder

# Run makefile to build all the commands
WORKDIR /usr/src/app
COPY . .
RUN make

# Copy server to /usr/local/bin
FROM --platform=${PLATFORM}/${ARCH} debian:bullseye-slim
RUN apt -y update && apt -y upgrade
COPY --from=builder /usr/src/app/build/* /usr/local/bin/
COPY etc/docker/server-entrypoint.sh .

# Labels
LABEL org.opencontainers.image.source=https://github.com/mutablelogic/go-server

# Entrypoint when running the server
ENTRYPOINT [ "/server-entrypoint.sh" ]
CMD [ "/usr/local/bin/server" ]

