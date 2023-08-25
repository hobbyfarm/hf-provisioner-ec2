package handlers

import (
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/handlers/instance"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/handlers/keypair"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/handlers/virtualmachine"
	labels2 "github.com/hobbyfarm/hf-provisioner-shared/labels"
	"github.com/hobbyfarm/hf-provisioner-shared/namespace"
	provider "github.com/hobbyfarm/hf-provisioner-shared/provider"
	"github.com/hobbyfarm/hf-provisioner-shared/ssh"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	Finalizer = "provisioning.hobbyfarm.io/aws-ec2"
)

func RegisterRoutes(providerName string) provider.RouteAdder {
	return func(router *router.Router) error {
		vmRouter := router.Type(&v1.VirtualMachine{}).Namespace(namespace.ResolveNamespace()).Selector(
			labels.SelectorFromSet(map[string]string{
				labels2.ProvisionerLabel: providerName,
			}))

		vmRouter.FinalizeFunc(Finalizer, virtualmachine.ProvisionerFinalizer)
		vmRouter.HandlerFunc(keypair.SecretHandler)
		vmRouter.Middleware(ssh.RequireSecret).HandlerFunc(keypair.KeyPairHandler)
		vmRouter.Middleware(keypair.KeyPairCreated).HandlerFunc(instance.InstanceHandler)

		keyPairRouter := router.Type(&v1alpha1.KeyPair{}).Namespace(namespace.ResolveNamespace())

		keyPairRouter.HandlerFunc(keypair.EnsureKeyPairStatus)
		keyPairRouter.Middleware(keypair.KeyPairImported(false)).HandlerFunc(keypair.CreateKeyPair)
		keyPairRouter.Middleware(keypair.KeyPairImported(true)).HandlerFunc(keypair.UpdateVM)
		keyPairRouter.FinalizeFunc(Finalizer, keypair.KeyPairFinalizer)

		instanceRouter := router.Type(&v1alpha1.Instance{}).Namespace(namespace.ResolveNamespace())

		instanceRouter.HandlerFunc(instance.EnsureInstanceStatus)
		instanceRouter.Middleware(instance.InstanceCreated(false)).HandlerFunc(instance.CreateInstance)
		instanceRouter.Middleware(instance.InstanceCreated(true)).HandlerFunc(instance.PeriodicUpdate)
		instanceRouter.HandlerFunc(instance.WriteVM)
		instanceRouter.FinalizeFunc(Finalizer, instance.InstanceFinalizer)

		return nil
	}
}
