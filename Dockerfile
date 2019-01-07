FROM golang:1.11-alpine3.8 as builder

WORKDIR /go/src/github.com/getyourguide/k8s-cluster-manager/
COPY . /go/src/github.com/getyourguide/k8s-cluster-manager/
RUN apk add --no-cache  \
      bash              \
      gcc			\
      git 			\
      musl-dev          \
      dep               \
      zlib-dev


ENV GOOS linux
ENV GOARCH amd64

RUN go get . 
# TODO move to dep instead.
# RUN dep ensure 

RUN go build -ldflags "-X main.GitCommit=$(git rev-parse --short HEAD)" -tags static_all .

FROM ubuntu

RUN apt-get update && apt-get install curl git -y

RUN curl https://storage.googleapis.com/kubernetes-helm/helm-v2.12.1-linux-amd64.tar.gz | tar -xvz \
    &&  mv linux-amd64/helm /usr/local/bin/ \
    &&  mv linux-amd64/tiller /usr/local/bin/ \
    && rm -rf /var/lib/apt/lists/*


ENV HELM_HOME /root/.helm
RUN mkdir -p /root/.helm/plugins

RUN helm plugin install https://github.com/rimusz/helm-tiller

COPY --from=builder /go/src/github.com/getyourguide/k8s-cluster-manager/k8s-cluster-manager .
COPY entrypoint.sh /
COPY k8s-cluster-manager /

WORKDIR /

ENTRYPOINT [ "/entrypoint.sh" ]
# CMD ["/app/k8s-cluster-manager"]
