FROM golang:1.11-alpine3.8 as builder

WORKDIR /go/src/github.com/getyourguide/k8s-cluster-manager/
COPY . /go/src/github.com/getyourguide/k8s-cluster-manager/
RUN apk add --no-cache  \
      bash              \
      gcc			\
      git 			\
      musl-dev          \
      dep               \
      make              \
      zlib-dev

RUN dep ensure 

RUN make build

FROM alpine

# RUN apt-get update && apt-get install curl git -y
RUN apk add --no-cache  \
      git               \
      bash              \
      curl			

RUN curl https://storage.googleapis.com/kubernetes-helm/helm-v2.12.1-linux-amd64.tar.gz | tar -xvz \
    &&  mv linux-amd64/helm /usr/local/bin/ \
    &&  mv linux-amd64/tiller /usr/local/bin/

ENV HELM_HOME /root/.helm
RUN mkdir -p /root/.helm/plugins

RUN helm plugin install https://github.com/rimusz/helm-tiller

COPY --from=builder /go/src/github.com/getyourguide/k8s-cluster-manager/k8s-cluster-manager .
COPY entrypoint.sh /

WORKDIR /

ENTRYPOINT [ "/entrypoint.sh" ]
# CMD ["/app/k8s-cluster-manager"]
