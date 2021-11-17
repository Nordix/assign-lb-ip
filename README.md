# assign-lb-ip

Assigns `loadBalancerIP` address to a [Kubernetes](https://kubernetes.io/docs/concepts/services-networking/#loadbalancer)
service for testing purposes.

This is normally done by the cloud provider or the
[metallb](https://github.com/metallb/metallb) "controller".  It is
not possible to set the `Status.loadBalancer.Ingress` with `kubectl`
(AFAIK), so this utility is needed.

## Usage

Dual-stack phase 3 released in K8s v1.20 (alpha) is assumed.

Use `assign-lb-ip -help` to get a brief help printout.

The easiest way is to define the `loadBalancerIP` in the service
manifest;

```
apiVersion: v1
kind: Service
metadata:
  name: mconnect-ipv6-lb
spec:
  ipFamilies:
  - IPv6
  selector:
    app: mconnect
  ports:
  - port: 5001
  type: LoadBalancer
  loadBalancerIP: 1000::8
```

The EXTERNAL-IP will still be in `<pending>` but you can now simply
run `assign-lb-ip` to set it. Example;

```
$ kubectl get svc mconnect-ipv6-lb
NAME               TYPE           CLUSTER-IP        EXTERNAL-IP   PORT(S)          AGE
mconnect-ipv6-lb   LoadBalancer   fd00:4000::250b   <pending>     5001:32030/TCP   5m5s
$ assign-lb-ip -svc mconnect-ipv6-lb
$ kubectl get svc mconnect-ipv6-lb
NAME               TYPE           CLUSTER-IP        EXTERNAL-IP   PORT(S)          AGE
mconnect-ipv6-lb   LoadBalancer   fd00:4000::250b   1000::8       5001:32030/TCP   5m36s
```

`assign-lb-ip` will take the `loadBalancerIP` you specified in the
manifest and set it as the "real" load-balancer IP.

You can also explicitly specify the load-balancer IP using the `-ip`
option.

The service must have `type: LoadBalancer` and an explicitly specified
`-ip` must work with `net.ParseIP()`. Other than that no checks are
made.

You can also define a dual-stack service;

```
apiVersion: v1
kind: Service
metadata:
  name: mserver-preferdual-lb
spec:
  ipFamilyPolicy: PreferDualStack
  selector:
    app: mserver
  ports:
  - port: 5001
  type: LoadBalancer
```

Here `PreferDualStack` means that the service will be dual-stack if
deployed in a dual-stack cluster. The "loadBalancerIP" field can not
be used but a new `loadBalancerIPs` fiels is proposed
([PR](https://github.com/kubernetes/enhancements/pull/1992)).

For dual-stack services a comma separated list of ip's is specified;

```
$ assign-lb-ip -svc mserver-preferdual-lb -ip 10.0.0.2,1000::2
$ kubectl get svc mserver-preferdual-lb
NAME                    TYPE           CLUSTER-IP    EXTERNAL-IP        PORT(S)                         AGE
mserver-preferdual-lb   LoadBalancer   12.0.22.230   10.0.0.2,1000::2   5001:31686/TCP,5003:31406/TCP   2m16s
```

Use the `-clear` option to remove LoadBalancerIPs.

### Older dual-stack versions

Before K8s v1.20.0 only single-stack services could be specified but
for both families. Syntax is sligtly different but `assign-lb-ip`
works the same way.

```
apiVersion: v1
kind: Service
metadata:
  name: mconnect-ipv6-lb
spec:
  ipFamily: IPv6
  selector:
    app: mconnect
  ports:
  - port: 5001
  type: LoadBalancer
  loadBalancerIP: 1000::8
```

## Build

```
GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o assign-lb-ip \
  -ldflags "-extldflags '-static' -X main.version=$(date +%F:%T)" \
  ./cmd/...
```


