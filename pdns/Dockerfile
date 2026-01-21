FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive
 
# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    autoconf \
    automake \
    libtool \
    pkg-config \
    bison \
    flex \
    ragel \
    libboost-all-dev \
    libssl-dev \
    liblua5.3-dev \
    libsqlite3-dev \
    libmysqlclient-dev \
    libpq-dev \
    libcurl4-openssl-dev \
    libyaml-cpp-dev \
    libcap-dev \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Download PowerDNS release tarball
WORKDIR /build
RUN curl -LO https://downloads.powerdns.com/releases/pdns-4.9.0.tar.bz2 \
    && tar xjf pdns-4.9.0.tar.bz2

WORKDIR /build/pdns-4.9.0

# Build PowerDNS
RUN autoreconf -vi \
    && ./configure --with-modules="bind gsqlite3" --with-sqlite3 --sysconfdir=/var/lib/powerdns \
    && make -j$(nproc) \
    && make install

# Create pdns user
RUN useradd -r -s /usr/sbin/nologin pdns

# Expose DNS ports
EXPOSE 53/tcp 53/udp

CMD ["pdns_server", "--daemon=no"]
