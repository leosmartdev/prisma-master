# Specifying the image
FROM ubuntu:xenial

# System dependencies
RUN apt-get update
RUN apt-get install -y \
    build-essential \
    dpkg-dev \
    git \
    wget \
    curl \
    python3-minimal \
    python3-pip \
    libpam-pwdfile \
    libfontconfig1 \
    libxrender1 \
    unzip \
    graphviz \
    shared-mime-info \
    rsync \
    rsyslog

# Upgrade pip
RUN pip3 install --upgrade "pip < 21.0"

# Install python packages for mkdocs
RUN pip3 install mkdocs mkdocs-material pymdown-extensions

# Install AWS command line
RUN pip3 install awscli --upgrade

# Install protobuf
RUN cd /opt ; \
    wget --quiet https://github.com/google/protobuf/releases/download/v3.5.1/protoc-3.5.1-linux-x86_64.zip ; \
    mkdir /opt/protoc ; \
    cd /opt/protoc ; \
    unzip ../protoc*.zip ; \
    find /opt/protoc -type d -exec chmod 755 {} \; ;\
    find /opt/protoc -type f -exec chmod 644 {} \; ;\
    chmod 755 /opt/protoc/bin/protoc ;

# ejdb needs to be installed from our aws deps location due to version changes
RUN wget --quiet https://prisma-c2.s3.amazonaws.com/dependencies/all-dependencies/ejdb_1.2.12-ppa1~xenial1_amd64.deb && dpkg -i ejdb*.deb && rm ejdb*.deb

# Install GO
RUN wget --quiet https://dl.google.com/go/go1.15.5.linux-amd64.tar.gz && tar -xf go*.tar.gz && mv go /usr/local && rm go*.tar.gz
RUN mkdir -p /root/go/bin
RUN mkdir -p /root/go/src
ENV GOPATH="/root/go"
ENV PATH="/root/go/bin:/opt/protoc/bin:/usr/local/go/bin:${PATH}"

# Install golangci-lint
RUN	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b ${GOPATH}/bin v1.15.0


# export GO111MODULE, and download the dependencies. 
ENV GO111MODULE=on
RUN go get -u github.com/golang/protobuf/protoc-gen-go

# pb go tag bson
RUN go get -u github.com/arkavo-com/pb-go-tag-bson

WORKDIR /root/
