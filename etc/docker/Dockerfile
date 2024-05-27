ARG OS
ARG ARCH

# Run makefile to build all the commands
FROM --platform=${OS}/${ARCH} golang:1.22 AS builder
ARG OS
ARG ARCH
WORKDIR /usr/src/app
COPY . .
RUN OS=${OS} ARCH=${ARCH} make

# Copy server and plugins to /usr/local/bin
FROM --platform=${OS}/${ARCH} debian:bookworm-slim
ARG OS
ARG ARCH
ARG SOURCE

# Install nginx and dependencies
RUN set -x \
  && groupadd --system --gid 101 nginx \
  && useradd --system --gid nginx --no-create-home --home /nonexistent --comment "nginx user" --shell /bin/false --uid 101 nginx \
  && apt update \
  && apt install --no-install-recommends --no-install-suggests -y gnupg1 curl gnupg2 ca-certificates lsb-release debian-archive-keyring \
  && curl https://nginx.org/keys/nginx_signing.key | gpg --dearmor | sudo tee /usr/share/keyrings/nginx-archive-keyring.gpg >/dev/null  \
  && echo "deb [signed-by=/usr/share/keyrings/nginx-archive-keyring.gpg] http://nginx.org/packages/debian `lsb_release -cs` nginx" | sudo tee /etc/apt/sources.list.d/nginx.list \
  && apt update \
  && apt install --no-install-recommends --no-install-suggests -y nginx

# Copy server and plugins to /usr/local/bin
COPY --from=builder /usr/src/app/build/* /usr/local/bin/
COPY etc/docker/entrypoint.sh .
RUN chmod +x entrypoint.sh

# Labels
LABEL org.opencontainers.image.source=${SOURCE}

# Entrypoint when running the server
ENTRYPOINT [ "/entrypoint.sh" ]
EXPOSE 80
STOPSIGNAL SIGQUIT
CMD [ "/usr/local/bin/nginx-server" ]
