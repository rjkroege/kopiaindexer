# A Dockerfile that knows how to make mk because the kopiaindexer uses
# mk. Build like so:
# rcpu ween docker build . --tag mkbuilder

FROM debian:12.1-slim
# FROM cpud


RUN set -eux \
    && DEBIAN_FRONTEND=noninteractive apt-get update -qq \
    && DEBIAN_FRONTEND=noninteractive apt-get install -qq -y --no-install-recommends --no-install-suggests \
        ca-certificates \
	curl \
	git \
	xz-utils


WORKDIR /tool
RUN curl https://ziglang.org/download/0.11.0/zig-linux-x86_64-0.11.0.tar.xz | tar xJf -

WORKDIR /usr/local
RUN git clone https://github.com/rjkroege/plan9port.git plan9

WORKDIR /usr/local/plan9
RUN git checkout -b hermetic-rc origin/hermetic-rc
RUN PATH=/tool/zig-linux-x86_64-0.11.0:$PATH zig build -Doptimize=ReleaseSmall

