# snoopy-operator

### A Networking Packet Capture and Flow Analysis Operator

How can we get packet captures, network flow information or monitor deeply network communications in multiple pods at the same time and have all that information centralized? How can we troubleshoot certain network performance related issues having the "view" of a pod instead? As if we can see the packets before an external encryption occurs? What if we need to check certain destinations or sessions or connections success rate from pod to pod or from pod to external services?

I'm pretty sure a bunch of tools come up to mind when talking about those challenges. Tools like tcpdump, iperf, tc, iproute suite, eBPF tracing etc. Just using one of them for a single Pod is already a challenge. But how can we use them spread over a large cluster on specific pods, namespaces and nodes? Doing that seamlessly without disturbing or changing kubernetes resources at all?

That's the motivation behind the snoopy-operator. Snoopy, for short, manages multiple jobs running special tools against selected or labeled pods across multiple nodes collecting valuable information without changing the pod's object or affecting the pod's cpu budget for example.

In order to achieve it's results it makes use of [podtracer](https://github.com/fennec-project/podtracer) a utility that makes incredibly easy running all those mentioned tools and many others targeting pods and, more specifically, containers inside those pods. By using podtracer instances as Scheduled Job instances many vital pieces of information can be captured and transferred to a central location.

# Road Map

At this moment snoopy-operator can run the jobs with podtracer and use tcpdump as jobs logging out packets. Both pieces of software must evolve to include other troubleshooting, monitoring and deep inspection tools. Among many desired features we highlight a few that are part of our commuity discussions below:

- Centralizing data on a Kafka topic to be consumed by specialized processes as part of a data processing pipeline.
- Including tools like iperf to run specialized performance tests at scale.
- Including eBPF filters for security monitoring.
- The creation of a data pipeline and dashboard to analyze and publish results.

# Running the PoC

A very simple PoC was created to allow people to understand what happens under the hood of this operator and here is how you can try it:

#### 1 - Clone the Project

```
git clone https://github.com/fennec-project/snoopy-operator.git

cd snoopy-operator
```
#### 2 - Label the Tenant Worker Node

We need to label one of our worker nodes to receive a couple of testing pods. The ones that will have their network interfaces tapped by snoopy-operator and have those packets captured.

For that find your nodes:

```
kubectl get nodes

ip-10-0-143-28.ca-central-1.compute.internal    Ready    master   
ip-10-0-164-241.ca-central-1.compute.internal   Ready    master   
ip-10-0-171-95.ca-central-1.compute.internal    Ready    worker   
ip-10-0-195-10.ca-central-1.compute.internal    Ready    master   
ip-10-0-216-145.ca-central-1.compute.internal   Ready    worker
```
And then label one of them with the cnf-telco=true label like below:

`oc label node ip-10-0-171-95.ca-central-1.compute.internal  cnf-telco=true`

That is to mimic a telco tenant only node for the example.

#### 3 - Deploy the Tenant Pods

Now run the sample-deployment.yaml file that has a node selector like below:

```
      nodeSelector:
        cnf-telco: "true"    
```

`oc apply -f config/samples/sample-deployment.yaml`

You should now have two pods running on the selected node:
```
oc get pods -n cnf-telco -o wide

cnf-example-pod-98b9d4df8-gmhx8   1/1     Running   0          10s   10.128.3.238   ip-10-0-171-95.ca-central-1.compute.internal

cnf-example-pod-98b9d4df8-p6qkn   1/1     Running   0          10s   10.128.3.239   ip-10-0-171-95.ca-central-1.compute.internal
```


#### 4 - Label another worker node as Management Node

That will allow us to separate snoopy-operator from a tenants node. And have it on a management utility node both for security and separation of concerns.

`kubectl label node ip-10-0-216-145.ca-central-1.compute.internal management=true`

#### 5 - Deploy Snoopy Operator

Snoopy, as you may have guessed, has the following node selector:

```
      nodeSelector:
        management: "true"
```
That should put it in a separate node.

To install the operator just run the `make deploy` command on the root of the project so basically:
```
make deploy
```
You should see something like below coming up on your terminal:

```
/Users/alex/go/src/github.com/fennec-project/snoopy-operator/bin/controller-gen "crd:trivialVersions=true,preserveUnknownFields=false" rbac:roleName=snoopy-operator-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
cd config/manager && /Users/alex/go/src/github.com/fennec-project/snoopy-operator/bin/kustomize edit set image controller=quay.io/fennec-project/snoopy-operator:0.0.1
/Users/alex/go/src/github.com/fennec-project/snoopy-operator/bin/kustomize build config/default | kubectl apply -f -
namespace/snoopy-operator created
customresourcedefinition.apiextensions.k8s.io/tcpdumps.pcap.fennecproject.io created
serviceaccount/snoopy-operator-sa created
role.rbac.authorization.k8s.io/snoopy-operator-scc-priv created
role.rbac.authorization.k8s.io/leader-election-role created
clusterrole.rbac.authorization.k8s.io/snoopy-operator-role created
rolebinding.rbac.authorization.k8s.io/rolebinding-priv-scc-snoopy-operator created
rolebinding.rbac.authorization.k8s.io/leader-election-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/snoopy-operator-rolebinding created
configmap/manager-config created
deployment.apps/snoopy-operator created
```

Finally you should see the operator running after a few seconds in its own namespace and in the management selected node like below:

```
NAME                               READY   STATUS    RESTARTS   AGE   IP            NODE                                            
snoopy-operator-76fb4d998b-cxrwz   1/1     Running   1          91s   10.131.1.17   ip-10-0-216-145.ca-central-1.compute.internal
```

#### 6 - Verify the existence of a tcpdump CRD

With the `make deploy` command the CRD created for this simple PoC was also installed in the cluster:

```
kubectl get crds | grep tcpdump
tcpdumps.pcap.fennecproject.io                                    2021-07-28T20:57:40Z
```

#### 7 - Capturing Packets from the Sample Pods

One note before deploying the sample CR to capture packets. Our sample pods also have a label that says `networkMonitor="true"` like below:

```
  labels:
    networkMonitor: "true"
```

Now let's finally take a look at our tcpdump CR sample:

```
apiVersion: pcap.fennecproject.io/v1alpha1
kind: Tcpdump
metadata:
  name: tcpdump-sample
spec:
  name: test-pcap
  targetNamespace: cnf-telco
  podLabel: { 
    networkMonitor: "true",
    }
  interfaceName: eth0
  packetCount: 50
```

From there you can see that we have targetNamespace that will indicate where to find the pods for packet capture and a podLabel that filters only the pods with that label. In our case it matches our previously deployed sample cnf tenant pods.

Only two parameters were put together for a simple test:

`interfaceName`: pod interface to read the packets from. Maps to the tcpdump option -i.

`packetCount`: the number of packets to be captured. Maps to the -c option of tcpdump.


Let's try our CR on the cnf-telco namespace as if it was a request from the CNF tenant to monitor specific Pod Networks by running:

`kubectl apply -f config/samples/pcap_v1alpha1_tcpdump.yaml -n cnf-telco`

#### 8 - Verifying Snoopy Operator Jobs

```
kubectl get jobs -n snoopy-operator

NAME                                      COMPLETIONS   DURATION   AGE
tcpdump-cnf-example-pod-98b9d4df8-gmhx8   0/1           2m13s      2m13s
tcpdump-cnf-example-pod-98b9d4df8-p6qkn   0/1           2m13s      2m13s
```


#### 9 - Verifying tcpdump pods generated by the Jobs

```
kubectl get pods -n snoopy-operator | grep tcpdump
tcpdump-cnf-example-pod-98b9d4df8-gmhx8-s75pw   0/1     Completed   0          4m1s
tcpdump-cnf-example-pod-98b9d4df8-p6qkn-q945h   1/1     Running     0          4m1s
```

#### 10 - Checking the 50 packets captured by the Job

```

```


The jobs will be created on the snoopy-operator namespace but run on the nodes where the pods are in order to tap into their network interfaces. By running `kubectl get jobs -n snoopy-operator` we should be able to see a list of jobs.

The same way, by running `kubectl get pods -n snoopy-operator` we should be able to see not only the snoopy-operator pod but also the Job's pods which will hold in their logs the output of tcpdump for each pod targeted in that packet capture.

# What comes next after the PoC

The snoopy-operator APIs may change yet a lot. Tcpdump CRD was a quick way to test the idea. Now we can evolve to a full featured podtracer API as well as more specialized APIs using Tcpdump or other tools. This is all under development at this point.

# Contribution

Regular meetings, slack channel and YouTube channel coming soon.


