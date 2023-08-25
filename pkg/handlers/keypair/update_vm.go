package keypair

import (
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateVM(req router.Request, resp router.Response) error {
	kp := req.Object.(*v1alpha1.KeyPair)

	// from the keypair, get to the vm
	// then update that VM with deets from the keypair
	// simple, right?

	vm := v1.VirtualMachine{}
	err := req.Client.Get(req.Ctx, client.ObjectKey{
		Name:      kp.Spec.Machine,
		Namespace: kp.Namespace,
	}, &vm)
	if err != nil {
		return fmt.Errorf("error retrieving vm %s: %s", kp.Spec.Machine, err.Error())
	}

	vm.Spec.SecretName = kp.Spec.Secret // spec as status. gross.

	return req.Client.Update(req.Ctx, &vm)
}
