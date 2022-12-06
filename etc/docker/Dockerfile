ARG PLATFORM
ARG ARCH
ARG VERSION
FROM --platform=${PLATFORM}/${ARCH} golang:1.19-bullseye AS builder

# Run makefile to build server and plugins
WORKDIR /usr/src/app
COPY . .
RUN make

# Copy server and plugins to /usr/local/bin
FROM --platform=${PLATFORM}/${ARCH} library/nginx:${VERSION}
COPY --from=builder /usr/src/app/build/* /usr/local/bin/
COPY etc/docker/server-entrypoint.sh .

# Create directories for go-server, the json directory is used to store the
# configuration files
COPY etc/json/ /var/lib/go-server/json/

# Create directories for nginx, the available and enabled directories are
# used to store the nginx configuration files
RUN mkdir -p /var/lib/nginx/available
RUN mkdir -p /var/lib/nginx/enabled
COPY etc/nginx/ /var/lib/nginx/

# Entrypoint when running the server
ENTRYPOINT [ "/server-entrypoint.sh" ]
CMD [ "/usr/local/bin/server", "-plugins=/usr/local/bin/*.plugin", "/var/lib/go-server/json/*.json" ]
