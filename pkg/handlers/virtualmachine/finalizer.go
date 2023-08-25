package virtualmachine

import (
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
	"github.com/hobbyfarm/hf-provisioner-shared/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func ProvisionerFinalizer(req router.Request, resp router.Response) error {
	vm := req.Object.(*v1.VirtualMachine)

	// before deleting the VM, make sure the instance and key are gone
	instanceList := new(v1alpha1.InstanceList)
	err := req.Client.List(req.Ctx, instanceList, client.MatchingLabels{
		labels.VirtualMachineLabel: vm.GetName(),
	})
	if err != nil {
		resp.RetryAfter(5 * time.Second)
		return fmt.Errorf("error retrieving list of instances for vm %s: %s", vm.GetName(), err.Error())
	}

	for _, instance := range instanceList.Items {
		err = req.Client.Delete(req.Ctx, &instance)
		resp.RetryAfter(5 * time.Second)
		if err != nil {
			return fmt.Errorf("error deleting instance %s: %s", instance.GetName(), err.Error())
		}
	}

	keyPairList := &v1alpha1.KeyPairList{}
	err = req.Client.List(req.Ctx, keyPairList, client.MatchingLabels{
		labels.VirtualMachineLabel: vm.GetName(),
	})
	if err != nil {
		return fmt.Errorf("error retrieving list of keypairs for vm %s: %s", vm.GetName(), err.Error())
	}

	for _, keyPair := range keyPairList.Items {
		err = req.Client.Delete(req.Ctx, &keyPair)
		resp.RetryAfter(5 * time.Second)
		if err != nil {
			return fmt.Errorf("error deleting keypair %s: %s", keyPair.GetName(), err.Error())
		}
	}

	return nil
}
