ARG PLATFORM
ARG ARCH
ARG VERSION
FROM --platform=${PLATFORM}/${ARCH} golang:1.19-bullseye AS builder

# Run makefile to build server and plugins
WORKDIR /usr/src/app
COPY . .
RUN make

FROM --platform=${PLATFORM}/${ARCH} library/nginx:${VERSION}
COPY --from=builder /usr/src/app/build/* /usr/local/bin/
COPY etc/docker/server-entrypoint.sh .
COPY etc/json/ /var/lib/go-server/json/

# Entrypoint when running the server
ENTRYPOINT [ "/server-entrypoint.sh" ]
CMD [ "/usr/local/bin/server", "-plugins=/usr/local/bin/*.plugin", "/var/lib/go-server/json/*.json" ]
