# assign-lb-ip

Assigns `loadBalancerIP` address to a [Kubernetes](https://kubernetes.io/docs/concepts/services-networking/#loadbalancer)
service for testing purposes.

This is normally done by the cloud provider or the
[metallb](https://github.com/danderson/metallb) "controller".  It is
not possible to set the `Status.loadBalancer.Ingress` with `kubectl`
(AFAIK), so this utility is needed.

## Usage

Use `assign-lb-ip -help` to get a brief help printout.

The easiest way is to define the `loadBalancerIP` in the service
manifest;

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

`assign-lb-ip` will simply take the `loadBalancerIP` you specified in
the manifest and set it as the "real" load-balancer IP.

You can also explicitly specify the load-balancer IP using the `-ip`
option.

The service must have `type: LoadBalancer` and an explicitly specified
`-ip` must work with `net.ParseIP()`. Other than that no checks are
made.

From v2.0 it is possible to specify a comma separated list of ip's;
```
$ assign-lb-ip -ip 1000::2,1000::4 -svc mconnect-ipv6
$ kubectl get svc mconnect-ipv6
NAME            TYPE           CLUSTER-IP        EXTERNAL-IP       PORT(S)          AGE
mconnect-ipv6   LoadBalancer   fd00:4000::ada8   1000::2,1000::4   5001:30380/TCP   134m
```

## Build

```
GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o assign-lb-ip \
  -ldflags "-extldflags '-static' -X main.version=$(date +%F:%T)" \
  ./cmd/...
```

## K8s Dual-stack phase 3

Dual-stack phase 3
([KEP](https://github.com/kubernetes/enhancements/blob/master/keps/sig-network/20180612-ipv4-ipv6-dual-stack.md))
will come in K8s v1.20 (alpha). This includes an API change. The
service above will look like;

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

A minimal change as you can see. But dual-stack phase 3 also allows
one dual-stack service to be defined, not two separate services for
ipv4 and ipv6;

```
apiVersion: v1
kind: Service
metadata:
  name: mconnect-lb
spec:
  ipFamilyPolicy: PreferDualStack
  selector:
    app: mconnect
  ports:
  - port: 5001
  type: LoadBalancer
  loadBalancerIP: "1000::8,10.0.0.1"
```

Here `PreferDualStack` means that the service will be dual-stack if
deployed in a dual-stack cluster. The "loadBalancerIP" is a list of
addresses. K8s does not document that a list of addresses may be used
but K8s does not use this filed at all and does not make any syntax
check, so this works.

`Assign-lb-ip` >= v2.1.0 is able to interpret a list of addresses in
`loadBalancerIP`.

