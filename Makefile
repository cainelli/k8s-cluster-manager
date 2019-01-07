SHORT_NAME = k8s-cluster-manager
GOOS ?= linux
GOARCH ?= amd64
PWD = $(shell pwd)

build:
	go build -o k8s-cluster-manager *.go
	env GOOS=${GOOS} GOARCH=${GOARCH} go build -o k8s-cluster-manager *.go

docker-build:
	docker build . -t cainelli/${SHORT_NAME}

docker-run:
	docker run --rm -it 						\
		-v ${PWD}:/app							\
		-v /:/host 							 	\
		-v /Users/cainelli/.kube/:/root/.kube	\
		cainelli/${SHORT_NAME} \
		/app/k8s-cluster-manager

docker-shell:
	docker run --rm -it 						\
		-v ${PWD}:/app							\
		-v /:/host 							 	\
		-v /Users/cainelli/.kube/:/root/.kube	\
		cainelli/${SHORT_NAME} bash
docker: build docker-run

docker-push:
	docker push cainelli/${SHORT_NAME}
