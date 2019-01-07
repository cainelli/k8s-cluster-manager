#!/bin/bash

helm tiller start &

sleep 3

/k8s-cluster-manager "$@"
