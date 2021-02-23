# go-redis-kubernetes

> Based on [this blog post](https://www.callicoder.com/deploy-multi-container-go-redis-app-kubernetes/)

## Run Application

```bash
$ go build

$ ./go-redis-kubernetes
2021/02/23 10:43:23 Starting server
2021/02/23 10:43:53 Cache miss for date 2021-02-23
2021/02/23 10:43:53 Quote API returned: 200 OK
2021/02/23 10:44:02 Cache hit for date 2021-02-23
^C2021/02/23 10:44:14 Shutting down
2021/02/23 10:44:14 http: Server closed
```

## Build Docker image

```bash
$ docker build -t go-redis-kubernetes .
$ docker tag go-redis-kubernetes m99coder/go-redis-app:1.0.0
$ docker login
$ docker push m99coder/go-redis-app:1.0.0
```

## Create Kubernetes manifests

```bash
$ kubectl apply -f .
deployment.apps/go-redis-app created
service/go-redis-app-service created

$ kubectl get service/go-redis-app-service
NAME                   TYPE       CLUSTER-IP    EXTERNAL-IP   PORT(S)          AGE
go-redis-app-service   NodePort   10.96.94.54   <none>        9090:31831/TCP   2m21s

$ kubectl port-forward service/go-redis-app-service 9090:9090
Forwarding from 127.0.0.1:9090 -> 8080
Forwarding from [::1]:9090 -> 8080
Handling connection for 9090

$ kubectl logs service/go-redis-app-service
Found 3 pods, using pod/go-redis-app-5b6dd654c6-mxxrc
2021/02/23 10:02:01 Starting server
2021/02/23 10:03:13 Cache hit for date 2021-02-23
2021/02/23 10:03:14 Cache hit for date 2021-02-23
2021/02/23 10:03:16 Cache hit for date 2021-02-23
```
