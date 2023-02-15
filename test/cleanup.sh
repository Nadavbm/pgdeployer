#!/bin/sh
kubectl delete crd pgdeployers.pgdeployer.example.com
kubectl delete clusterrolebinding pgdeploy-operator
kubectl delete clusterrole pgdeploy-operator
kubectl delete ns pgdeploy-operator
kubectl delete ns pgdep-test