package instance

import (
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/handlers/keypair"
	"github.com/hobbyfarm/hf-provisioner-shared/config"
	"github.com/hobbyfarm/hf-provisioner-shared/labels"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var InstanceTypeField = "instance_type"
var AMIField = "ami"
var SubnetField = "subnet"
var InstanceProfileField = "instance_profile"
var UserDataField = "user_data"
var SecurityGroupField = "security_group" // @TODO - Support array of SG ids

func InstanceHandler(req router.Request, resp router.Response) error {
	vm := req.Object.(*v1.VirtualMachine)

	key, err := keypair.GetKeyPair(req)
	if err != nil {
		return fmt.Errorf("error retrieving keypair for virtualmachine %s: %s", vm.GetName(), err.Error())
	}

	// retrieve existing? instance
	instance := &v1alpha1.Instance{}
	err = req.Client.Get(req.Ctx, client.ObjectKey{
		Namespace: vm.GetNamespace(),
		Name:      vm.GetName(),
	}, instance)

	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("error retrieving instance: %s", err.Error())
	}

	if instance.UID == "" {
		instance.Name = fmt.Sprintf("%s-%s", vm.GetName(), "instance")
		instance.Namespace = vm.GetNamespace()
	}

	instance.Spec.Machine = vm.GetName()

	if instance.GetLabels() == nil {
		instance.SetLabels(map[string]string{})
	}

	if l, ok := instance.GetLabels()[labels.VirtualMachineLabel]; !ok || l != vm.GetName() {
		instanceLabels := instance.GetLabels()
		instanceLabels[labels.VirtualMachineLabel] = vm.GetName()

		instance.SetLabels(instanceLabels)
	}

	required := map[string]string{
		InstanceTypeField:  "",
		AMIField:           "",
		SubnetField:        "",
		SecurityGroupField: "",
	}
	for k, _ := range required {
		if val, err := config.ResolveConfigItem(vm, req, k); err != nil {
			return fmt.Errorf("error resolving required config item %s: %s", k, err.Error())
		} else {
			required[k] = val
		}
	}

	rii := ec2.RunInstancesInput{
		InstanceType: aws.String(required[InstanceTypeField]),
		ImageId:      aws.String(required[AMIField]),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		SubnetId:     aws.String(required[SubnetField]),
		IamInstanceProfile: func() *ec2.IamInstanceProfileSpecification {
			if val, err := config.ResolveConfigItem(vm, req, InstanceProfileField); err == nil {
				return &ec2.IamInstanceProfileSpecification{
					Arn: aws.String(val),
				}
			}

			return nil
		}(),
		UserData: func() *string {
			if val, err := config.ResolveConfigItem(vm, req, UserDataField); err == nil {
				return aws.String(val)
			}

			return nil
		}(),
		SecurityGroupIds: func() []*string {
			if val, err := config.ResolveConfigItem(vm, req, SecurityGroupField); err == nil {
				return aws.StringSlice([]string{val})
			}

			return nil
		}(),
		KeyName: aws.String(key.GetName()),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeInstance),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(vm.GetName()),
					},
				},
			},
		},
	}

	jsonRii, err := json.Marshal(rii)
	if err != nil {
		return fmt.Errorf("error marshalling instance to json: %s", err.Error())
	}

	instance.Spec.Instance.Raw = jsonRii

	resp.Objects(instance)

	return nil
}
