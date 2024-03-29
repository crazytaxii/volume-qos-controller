FROM quay.io/ceph/ceph:v17 AS builder
ARG GOPROXY
ARG VERSION
ARG GOVERSION
ENV GOPROXY=${GOPROXY} \
    GOROOT=/usr/local/go

RUN mkdir -p ${GOROOT} && \
    ARCH=$(uname -m|sed 's|x86_64|amd64|'|sed 's|aarch64|arm64|') && \
    curl https://storage.googleapis.com/golang/go${GOVERSION}.linux-${ARCH}.tar.gz | tar xzf - -C ${GOROOT} --strip-components=1
RUN ${GOROOT}/bin/go version && \
    ${GOROOT}/bin/go env
RUN dnf -y install librados-devel librbd-devel gcc
ENV GOPATH=/go \
    CGO_ENABLED=1 \
    PATH="${GOROOT}/bin:${GOPATH}/bin:${PATH}"
WORKDIR /go/src/app
COPY . .
RUN go build -o dist/qos-controller -a -ldflags "-X 'main.version=${VERSION}'" ./cmd/qos-controller

FROM quay.io/ceph/ceph:v17
COPY --from=builder /go/src/app/dist/qos-controller /usr/local/bin/qos-controller
ENTRYPOINT [ "/usr/local/bin/qos-controller" ]
