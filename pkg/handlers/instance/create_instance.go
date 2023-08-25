package instance

import (
	"encoding/json"
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
	ec22 "github.com/hobbyfarm/hf-provisioner-ec2/pkg/ec2"
)

func CreateInstance(req router.Request, resp router.Response) error {
	instance := req.Object.(*v1alpha1.Instance)

	rii := &ec2.RunInstancesInput{}
	err := json.Unmarshal(instance.Spec.Instance.Raw, rii)
	if err != nil {
		v1alpha1.ConditionInstanceExists.False(instance)
		v1alpha1.ConditionInstanceExists.Message(instance, fmt.Sprintf("error unmarshalling json during create instance: %s", err.Error()))
		return fmt.Errorf("error unmarshalling json during create instance: %s", err.Error())
	}

	ec2Client, err := ec22.NewEC2Client(instance.Spec.Machine, req)
	if err != nil {
		v1alpha1.ConditionInstanceExists.False(instance)
		v1alpha1.ConditionInstanceExists.Message(instance, fmt.Sprintf("error creating ec2 client: %s", err.Error()))
		return fmt.Errorf("error creating ec2 client: %s", err.Error())
	}

	reservation, err := ec2Client.RunInstances(rii)
	if err != nil {
		v1alpha1.ConditionInstanceExists.False(instance)
		v1alpha1.ConditionInstanceExists.Message(instance, fmt.Sprintf("error during ec2 runinstances request: %s", err.Error()))
		return fmt.Errorf("error during ec2 runinstances request: %s", err.Error())
	}

	if len(reservation.Instances) > 0 {
		instanceJson, err := json.Marshal(reservation.Instances[0])
		if err != nil {
			v1alpha1.ConditionInstanceExists.True(instance)
			v1alpha1.ConditionInstanceExists.Message(instance, fmt.Sprintf("error marshalling instance json: %s", err.Error()))
			return fmt.Errorf("error marshalling instance json: %s", err.Error())
		}

		v1alpha1.ConditionInstanceExists.True(instance)
		v1alpha1.ConditionInstanceExists.Message(instance, fmt.Sprintf("instance created: %s", *reservation.Instances[0].InstanceId))

		instance.Status.Instance.Raw = instanceJson

		return req.Client.Status().Update(req.Ctx, instance)
	}

	return fmt.Errorf("ec2 runinstances returned reservation with no instances")
}
