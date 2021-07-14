package pcap

import (
	"context"
	"fmt"
	"testing"

	pcapv1alpha1 "github.com/fennec-project/snoopy-operator/apis/pcap/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)


func testTcpdumpController(t *testing.T) {
	
	testScenarios := []struct{
		name string
		ExpectedErrorMsg string
		tcpdumpName string
		tcpdumpTargetNamespace string // should be optional defaults to CR's namespace
		tcpdumpPodLabel map[string]string
		tcpdumpInterfaceName string
		tcpdumpPacketCount int64
		tcpdumpFileSize int32
		tcpdumpPcapFilePath string
		// TODO tcpdumpPauseJobs and/or tcpdumpAbortJobs for specific CR
		}{
		}

		for _, scenario := range testScenarios{
			t.Run(scenario.name, func(t *testing.T){

				tcpdump := &pcapv1alpha1.Tcpdump{}
				
				testPod := &corev1.Pod{}
				
				tcpdumpReconciler := TcpdumpReconciler{}

				// Label based filter (may be increased to other methods in the future)
				// Get pods on the CR namespace by label
				// needs to get pods in running state

				PodList, err := tcpdumpReconciler.GetRunningPodsByLabel(scenario.tcpdumpPodLabel)
				
				// needs to generate no pods matching specified label. Error for emtpy list
				if want, got := scenario.ExpectedErrorMsg, err.Error(); want != got {
					t.Errorf("Expected error %v got %v ", want, got)
					return
				}

				// verify if PodList is not empty
				if len(PodList) == 0 {
					fmt.Errorf("PodList is empty and reconciler didn't generate any errors.")
					return
				}

				// build command parameters from CR
				// tcpdump File size and packet count are mutually exclusive
				// case 1 is filse size and case 2 is packet count

				// this should be taken care by the CRD scheme validation on the CR
				// also need to test this using a real CR against k8s api
				
				// TODO: A more sophisticated schedule based on time should be built in the future
				// That would require a more complex validation rule that may or may not require a
				// validation webhook...
				
				// test error msg when both tcpdumpPacketCount and tcpdumpFileSize are zero

				for _, pod := range PodList {

					// should get the job running for the desired pod
					// any k8s errors should be passed along for ex: non-existent pods
					// No Job found should also be verified.
					job, err = tcpdumpReconciler.GetJobForPod(pod.ObjectMeta.Name)
					

					command, err = tcpdumpReconciler.GeneratePodtracerArgs(scenario.tcpdumpInterfaceName, 
					scenario.tcpdumpPacketCount, scenario.tcpdumpFileSize, Pod.ObjectMeta.Name, scenario.tcpdumpTargetNamespace)
					if err != nil {
						fmt.Printf("An error has occurred while generating arguments: %s", err.Error())
					}

					// generate the job object with defined command
					job, err = tcpdumpReconciler.GenerateJob(command, namespace, nodeName)
					if err != nil {
						fmt.Printf("An error has occurred while generating pcap job: %s", err.Error())
					}

					// create the job
					err = tcpdumpReconciler.Client.Create(context.Background(),job)
					if err != nil {
						fmt.Printf("An error has occurred while creating the job: %s", err.Error())
					}

					// should update job status on main CR (progress and completion)
					// should put together pods and job IDs for verification and status update

					err = tcpdumpReconciler.UpdateJobStatusOnCR(job)

					// Maybe mock a Job creation and progress changes to guarantee behavior
					// Read CR status and check on changes
					// verify job completion and update status field on CR

					// query pcap file on destination and ensure its integrity and existence
					err = tcpdumpReconciler.EnsurePcapFileAfterJobCompletion(job)

				}
			}
		}
	}