# pgdeployer

pgdeploy is a kubernetes operator build by using go-based [operator-sdk](https://sdk.operatorframework.io/)

the operator will deploy postgres to kubernetes by using crd in a namespace. 

the crd allows to configure the following postgres deployment specifications:

```
spec:
  cpu_lim: "1000m"
  cpu_req: "200m"
  mem_lim: "512Mi"
  mem_req: "256Mi"
  pg_version: "14"
  port: 5432
```

container requests and limits, postgres version and the port used to connect to the pod

in `test/` dir you'll find a read me file explaining how to test this operator in kubenretes.
