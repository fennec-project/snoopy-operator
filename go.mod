module github.com/fennec-project/snoopy-operator

go 1.16

require (
	github.com/containernetworking/plugins v0.9.1
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.3
	github.com/vishvananda/netlink v1.1.1-0.20201029203352-d40f9887b852
	google.golang.org/grpc v1.38.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/cri-api v0.21.1
	sigs.k8s.io/controller-runtime v0.8.3
)
