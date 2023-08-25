package instance

import (
	"encoding/json"
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	ec22 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/ec2"
)

func InstanceFinalizer(req router.Request, resp router.Response) error {
	instance := req.Object.(*v1alpha1.Instance)

	ec2Client, err := ec2.NewEC2Client(instance.Spec.Machine, req)
	if err != nil {
		return fmt.Errorf("error creating ec2 client: %w", err)
	}

	ec2Instance := &ec22.Instance{}
	err = json.Unmarshal(instance.Status.Instance.Raw, ec2Instance)
	if err != nil {
		return fmt.Errorf("error unmarshalling instance info from status: %w", err)
	}

	ins, err := ec2Client.DescribeInstances(&ec22.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{*ec2Instance.InstanceId}),
	})
	if err != nil {
		if err.(awserr.Error).Code() == "InvalidInstanceID.NotFound" {
			return nil
		} else {
			return fmt.Errorf("error describing instance: %w", err)
		}
	}

	if ins.Reservations[0] != nil && ins.Reservations[0].Instances[0] != nil {
		switch *ins.Reservations[0].Instances[0].State.Name {
		case ec22.InstanceStateNameTerminated:
		case ec22.InstanceStateNameShuttingDown:
			return nil
		}
	}

	_, err = ec2Client.TerminateInstances(&ec22.TerminateInstancesInput{
		InstanceIds: aws.StringSlice([]string{*ec2Instance.InstanceId}),
	})

	return nil
}
