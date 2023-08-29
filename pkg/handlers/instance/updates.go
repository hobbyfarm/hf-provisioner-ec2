package instance

import (
	"encoding/json"
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
	ec22 "github.com/hobbyfarm/hf-provisioner-ec2/pkg/ec2"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func WriteVM(req router.Request, resp router.Response) error {
	instance := req.Object.(*v1alpha1.Instance)

	ec2Instance := ec2.Instance{}
	if err := json.Unmarshal(instance.Status.Instance.Raw, &ec2Instance); err != nil {
		return fmt.Errorf("error unmarshalling instance status: %s", err.Error())
	}

	vm := &v1.VirtualMachine{}
	err := req.Client.Get(req.Ctx, client.ObjectKey{
		Namespace: instance.GetNamespace(),
		Name:      instance.Spec.Machine,
	}, vm)
	if err != nil {
		return fmt.Errorf("error retrieving vm %s: %s", instance.Spec.Machine, err.Error())
	}

	if ec2Instance.PrivateIpAddress != nil {
		vm.Status.PrivateIP = *ec2Instance.PrivateIpAddress
	}

	if ec2Instance.PublicIpAddress != nil {
		vm.Status.PublicIP = *ec2Instance.PublicIpAddress
	}

	if ec2Instance.InstanceId != nil {
		vm.Status.Hostname = *ec2Instance.InstanceId
	}

	if *ec2Instance.State.Name == ec2.InstanceStateNameRunning {
		if vm.Spec.Protocol == "ssh" {
			err := IsSSHReady(req, vm)
			if err != nil {
				v1alpha1.ConditionConnectionReady.False(instance)
				v1alpha1.ConditionConnectionReady.Message(instance, err.Error())
			} else {
				v1alpha1.ConditionConnectionReady.True(instance)
			}
			vm.Status.Status = v1.VmStatusRunning
		} else {
			v1alpha1.ConditionConnectionReady.Unknown(instance)
			v1alpha1.ConditionConnectionReady.Message(instance, "unsupported protocol for readiness check")

			vm.Status.Status = v1.VmStatusRunning
		}
	} else {
		switch *ec2Instance.State.Name {
		case ec2.InstanceStateNamePending:
			vm.Status.Status = v1.VmStatusProvisioned
		case ec2.InstanceStateNameShuttingDown:
		case ec2.InstanceStateNameStopping:
			vm.Status.Status = v1.VmStatusTerminating
		}
	}

	if err := req.Client.Status().Update(req.Ctx, instance); err != nil {
		return errors.Wrap(err, "error updating instance status")
	}

	return req.Client.Status().Update(req.Ctx, vm)
}

func PeriodicUpdate(req router.Request, resp router.Response) error {
	instance := req.Object.(*v1alpha1.Instance)

	ec2Client, err := ec22.NewEC2Client(instance.Spec.Machine, req)
	if err != nil {
		msg := fmt.Sprintf("error creating ec2 client: %s", err.Error())
		v1alpha1.ConditionInstanceUpdated.False(instance)
		v1alpha1.ConditionInstanceUpdated.Message(instance, msg)
		return fmt.Errorf(msg)
	}

	ec2Instance := ec2.Instance{}
	err = json.Unmarshal(instance.Status.Instance.Raw, &ec2Instance)
	if err != nil {
		msg := fmt.Sprintf("error unmarshalling ec2 instance from status: %s", err.Error())
		v1alpha1.ConditionInstanceUpdated.False(instance)
		v1alpha1.ConditionInstanceUpdated.Message(instance, msg)
		return fmt.Errorf(msg)
	}

	// get new instance update from describeinstance
	dii := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{
			*ec2Instance.InstanceId,
		}),
	}
	dio, err := ec2Client.DescribeInstances(dii)
	if err != nil {
		msg := fmt.Sprintf("error describing instances in ec2: %s", err.Error())
		v1alpha1.ConditionInstanceUpdated.False(instance)
		v1alpha1.ConditionInstanceUpdated.Message(instance, msg)
		return fmt.Errorf(msg)
	}

	if len(dio.Reservations) > 0 && len(dio.Reservations[0].Instances) > 0 {
		instanceJson, err := json.Marshal(dio.Reservations[0].Instances[0])
		if err != nil {
			msg := fmt.Sprintf("error marshalling instance response json during update: %s", err.Error())
			v1alpha1.ConditionInstanceUpdated.False(instance)
			v1alpha1.ConditionInstanceUpdated.Message(instance, msg)
			return fmt.Errorf(msg)
		}

		instance.Status.Instance.Raw = instanceJson

		switch *dio.Reservations[0].Instances[0].State.Name {
		case ec2.InstanceStateNameRunning:
			resp.RetryAfter(30 * time.Second)
		case ec2.InstanceStateNamePending:
			resp.RetryAfter(10 * time.Second)
		}

		return req.Client.Status().Update(req.Ctx, instance)
	}

	return nil
}

func IsSSHReady(req router.Request, vm *v1.VirtualMachine) error {
	// gather key
	var secret = &corev1.Secret{}
	if err := req.Client.Get(req.Ctx, client.ObjectKey{
		Namespace: vm.GetNamespace(),
		Name:      vm.Spec.SecretName,
	}, secret); err != nil {
		return err
	}

	privKey, err := ssh.ParsePrivateKey(secret.Data["private_key"])
	if err != nil {
		return errors.Wrap(err, "error parsing private key")
	}

	config := &ssh.ClientConfig{
		User: vm.Spec.SshUsername,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	_, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", vm.Status.PublicIP), config)
	if err != nil {
		return errors.Wrap(err, "error dialing ssh")
	}

	return nil
}
