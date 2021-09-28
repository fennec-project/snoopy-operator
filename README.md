# snoopy-operator

---

### A Cloud Native Way for Pod Deep Inspection

How can we get packet captures, network flow information or monitor deeply network communications in multiple pods at the same time and have all that information centralized? How can we troubleshoot certain network performance related issues having the "view" of a pod instead? As if we can see the packets before an external encryption occurs? What if we need to check certain destinations or sessions or connections success rate from pod to pod or from pod to external services?

I'm pretty sure a bunch of tools come up to mind when talking about those challenges. Tools like tcpdump, iperf, tc, iproute suite, eBPF tracing etc. Just using one of them for a single Pod is already a challenge. But how can we use them spread over a large cluster on specific pods, namespaces and nodes? Doing that seamlessly without disturbing or changing kubernetes resources at all?

That's the motivation behind the snoopy-operator. Snoopy, for short, manages multiple jobs running special tools against selected or labeled pods across multiple nodes collecting valuable information without changing the pod's object or affecting the pod's cpu budget for example.

### Architecture In a Nutshell

In order to achieve it's results it makes use of [podtracer](https://github.com/fennec-project/podtracer) a utility that makes incredibly easy running all those mentioned tools and many others targeting pods and, more specifically, containers inside those pods. By using podtracer instances as Scheduled Job instances many vital pieces of information can be captured and transferred to a central location from many different nodes at the same time.

<img src='docs/img/snoopy-operator.png'></img>

### Road Map

At this moment snoopy-operator can run the jobs with podtracer and use tcpdump as jobs logging out packets. Both pieces of software must evolve to include other troubleshooting, monitoring and deep inspection tools. Among many desired features we highlight a few that are part of our community discussions below:

- Send the retrieved data to a central server (Kafka for ex.) to be consumed by specialized processes as part of a data processing pipeline.
- Including tools like iperf to run specialized performance tests at scale.
- Including eBPF filters for security monitoring.
- The creation of a data pipeline and dashboard to analyze and publish results.

### Install Instructions

    New docs coming up soon!

### Contribution

Regular meetings, slack channel and YouTube channel coming soon.


