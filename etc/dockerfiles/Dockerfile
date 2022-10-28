ARG PLATFORM
ARG VERSION
FROM --platform=${PLATFORM} library/nginx:${VERSION}

# update the base packages and add a non-sudo user
ENV DEBIAN_FRONTEND="noninteractive" TZ="Europe/Berlin"
RUN apt-get update -y && apt-get upgrade -y && useradd -m nginx

