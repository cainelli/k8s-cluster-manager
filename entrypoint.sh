#!/bin/bash

helm tiller start &

/k8s-cluster-manager "$@"
