package keypair

import (
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	"github.com/hobbyfarm/hf-provisioner-shared/config"
	"github.com/hobbyfarm/hf-provisioner-shared/labels"
	"github.com/hobbyfarm/hf-provisioner-shared/namespace"
	"github.com/hobbyfarm/hf-provisioner-shared/ssh"
	corev1 "k8s.io/api/core/v1"
	labels2 "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SecretHandler creates a secret for the virtualmachine
func SecretHandler(req router.Request, resp router.Response) error {
	vm := req.Object.(*v1.VirtualMachine)

	var secret corev1.Secret
	var secretList = &corev1.SecretList{}
	err := req.Client.List(req.Ctx, secretList, &client.ListOptions{
		Namespace: namespace.ResolveNamespace(),
		LabelSelector: labels2.SelectorFromSet(map[string]string{
			labels.VirtualMachineLabel: vm.Name,
		}),
	})
	if err != nil {
		return fmt.Errorf("error getting secret list: %s", err.Error())
	}

	if len(secretList.Items) == 0 {
		secret = corev1.Secret{}

		public, private, err := ssh.GenKeyPair()
		if err != nil {
			return err
		}

		secret.Data = map[string][]byte{}

		secret.Data["public_key"] = []byte(public)
		secret.Data["private_key"] = []byte(private)

		secret.Name = fmt.Sprintf("%s-keys", vm.Name)
	} else {
		secret = secretList.Items[0]
	}

	if len(secret.Labels) == 0 {
		secret.Labels = map[string]string{}
	}

	secret.Labels[labels.VirtualMachineLabel] = vm.Name
	secret.Namespace = vm.Namespace
	if password, err := config.ResolveConfigItem(vm, req, "password"); err == nil {
		secret.Data["password"] = []byte(password)
	}

	if secret.UID == "" {
		err = req.Client.Create(req.Ctx, &secret)
	} else {
		err = req.Client.Update(req.Ctx, &secret)
	}

	if err != nil {
		return fmt.Errorf("error creating/updating secret: %s", err.Error())
	}

	return nil
}
