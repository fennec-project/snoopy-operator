package podnetwork

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func listPodsWithMatchingLabels(label string, value string) (*corev1.PodList, error) {

	cl, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	podList := &corev1.PodList{}

	// Get the list of pods that have a podNetworkConfig label
	err = cl.List(context.Background(), podList, client.MatchingLabels{label: value})
	if err != nil {
		fmt.Printf("failed to list pods matching labels: %v\n", err)
		os.Exit(1)
	}

	// Pods need to be at least created to proceed
	// Checking if the list is empty
	if len(podList.Items) <= 0 {
		return &corev1.PodList{}, fmt.Errorf("empty pod list")
	}
	return podList, nil
}

func getCRIOConnection() (*grpc.ClientConn, error) {

	var conn *grpc.ClientConn

	conn, err := grpc.Dial("unix:///var/run/crio/crio.sock", grpc.WithInsecure())

	if err != nil {
		fmt.Println("Connection failed: ", err)
		return nil, err
	}
	fmt.Println("Connected with CRI-O at unix:///var/run/crio/crio.sock")

	return conn, nil
}

func getCRIOContainerStatus(containerID string, grpcConn *grpc.ClientConn) (*cri.ContainerStatusResponse, error) {

	criClient := cri.NewRuntimeServiceClient(grpcConn)

	request := &cri.ContainerStatusRequest{
		ContainerId: containerID,
		Verbose:     true,
	}
	response, err := cri.RuntimeServiceClient.ContainerStatus(criClient, context.Background(), request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func parseCRIOContainerInfo(statusResponse *cri.ContainerStatusResponse) map[string]interface{} {

	var parsedContainerInfo map[string]interface{}

	containerInfo := statusResponse.Info["info"]

	json.Unmarshal([]byte(containerInfo), &parsedContainerInfo)

	return parsedContainerInfo
}

func getPid(pod corev1.Pod) (string, error) {

	// Get the container IDs for the given pod
	containerIDs := getContainerIDs(pod)

	// Connect with CRI-O's grpc endpoint
	conn, err := getCRIOConnection()
	if err != nil {
		return "", fmt.Errorf("Error getting CRIO connection: %v", err)
	}

	// Make a container status request to CRI-O
	// Here it doesn't matter which container ID inside the pod.
	// The goal is to put runtime configurations on Pod shared namespaces
	// like network and mount. Not intended for process/container specific namespaces.

	containerStatusResponse, err := getCRIOContainerStatus(containerIDs[0], conn)
	if err != nil {
		return "", fmt.Errorf("Error getting CRIO container status: %v", err)
	}

	parsedContainerInfo := parseCRIOContainerInfo(containerStatusResponse)

	return fmt.Sprintf("%.0f", parsedContainerInfo["pid"]), nil

}

func getContainerIDs(pod corev1.Pod) []string {

	containerIDs := []string{}

	// get container ID list
	for _, containerStatus := range pod.Status.ContainerStatuses {

		containerIDs = append(containerIDs, containerStatus.ContainerID[8:])

	}
	return containerIDs
}
