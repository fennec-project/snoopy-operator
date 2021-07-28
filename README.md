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

To install the operator just run the `make deploy` command on the root of the project so basically:

```
git clone https://github.com/fennec-project/snoopy-operator.git

cd snoopy-operator

make deploy
```

After that you should have snoopy-operator running on the snoopy-operator namespace and the tcpdump CRD installed on your cluster.

In order to test it, by creating a CR, check the config/samples folder where you can find an example on how to capture packets from multiple pods at the same time.

Let's take a look at this example:

```
apiVersion: pcap.fennecproject.io/v1alpha1
kind: Tcpdump
metadata:
  name: tcpdump-sample
spec:
  name: test-pcap
  targetNamespace: cnf-telco
  podLabel: { 
    podNetworkConfig: podnetwork-sample-a,
    }
  interfaceName: eth0
  packetCount: 10
  pcapFilePath: /pcap-data/test.pcap

```

The kind is Tcpdump, it's meant to capture packets in the cnf-telco namespace and it uses the podLabel mentioned there to find which pods on that namespace will receive a packet capture Job.

Tree simple parameters were created for packet capture. 

`interfaceName`: pod interface to read the packets from.

`packetCount`: the number of packets to be captured. Maps to the -c option of Tcpdump.

`pcapFilePath`: just indicates the path that should be used to temporarily store the pcap file.

The jobs will be created on the snoopy-operator namespace but run on the nodes where the pods are in order to tap into their network interfaces. By running `kubectl get jobs -n snoopy-operator` we should be able to see a list of jobs.

The same way, by running `kubectl get pods -n snoopy-operator` we should be able to see not only the snoopy-operator pod but also the Job's pods which will hold in their logs the output of tcpdump for each pod targeted in that packet capture.

# What comes next after the PoC

The snoopy-operator APIs may change yet a lot. Tcpdump CRD was a quick way to test the idea. Now we can evolve to a full featured podtracer API as well as more specialized APIs using Tcpdump or other tools. This is all under development at this point.

# Contribution

Regular meetings, slack channel and YouTube channel coming soon.


