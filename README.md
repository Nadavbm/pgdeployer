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


### Problem
There is a bug currently in the operator while creating all objects for postgres deployment (configMap, Secret, Service and Deployment). The problem is that reconcile loop will create only configMap as it will return result{}, nil (empty result which will not reconcile):
```
{"level":"INFO","time":"2023-02-21T22:40:50.492Z","message":"Start reconcile","namespace":"pgdep-test"}
{"level":"INFO","time":"2023-02-21T22:40:50.492Z","message":"create object","namespace":"pgdep-test","object":"pg-cm"}
{"level":"INFO","time":"2023-02-21T22:40:50.537Z","message":"create object","namespace":"pgdep-test","object":"pg-secret"}
{"level":"INFO","time":"2023-02-21T22:40:50.544Z","message":"create object","namespace":"pgdep-test","object":"pgdeployment"}
{"level":"INFO","time":"2023-02-21T22:40:50.554Z","message":"create object","namespace":"pgdep-test","object":"pg-service"}
{"level":"INFO","time":"2023-02-21T22:40:50.597Z","message":"Start reconcile","namespace":"pgdep-test"}
{"level":"INFO","time":"2023-02-21T22:40:50.597Z","message":"create object","namespace":"pgdep-test","object":"pg-cm"}
{"level":"INFO","time":"2023-02-21T22:40:50.630Z","message":"Start reconcile","namespace":"pgdep-test"}
{"level":"INFO","time":"2023-02-21T22:40:50.630Z","message":"create object","namespace":"pgdep-test","object":"pg-cm"}
{"level":"INFO","time":"2023-02-21T22:40:50.669Z","message":"Start reconcile","namespace":"pgdep-test"}
{"level":"INFO","time":"2023-02-21T22:40:50.670Z","message":"create object","namespace":"pgdep-test","object":"pg-cm"}
{"level":"INFO","time":"2023-02-21T22:40:51.844Z","message":"Start reconcile","namespace":"pgdep-test"}
{"level":"INFO","time":"2023-02-21T22:40:51.844Z","message":"create object","namespace":"pgdep-test","object":"pg-cm"}
```

### Solution
Recreate each object in another function, move specs to reconciler (controller) and restructure the order of object creation in reconcile loop 
