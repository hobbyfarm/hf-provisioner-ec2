package keypair

import (
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
	"github.com/hobbyfarm/hf-provisioner-shared/errors"
	"github.com/hobbyfarm/hf-provisioner-shared/labels"
	"github.com/hobbyfarm/hf-provisioner-shared/ssh"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func KeyPairHandler(req router.Request, resp router.Response) error {
	secret, err := ssh.GetSecret(req)
	if err != nil {
		return err
	}

	// try to get existing key
	var key *v1alpha1.KeyPair
	key, err = GetKeyPair(req)
	if errors.IsNotFound(err) {
		// need to create
		key = &v1alpha1.KeyPair{}
	} else if err != nil {
		return err
	}

	name := fmt.Sprintf("%s-key", req.Object.GetName())
	key.Name = name
	key.Namespace = req.Object.GetNamespace()

	if len(key.Labels) == 0 {
		key.Labels = map[string]string{}
	}

	key.Labels[labels.VirtualMachineLabel] = req.Object.GetName()
	key.Spec.Machine = req.Object.GetName()
	key.Spec.Secret = secret.Name

	pubKey, ok := secret.Data["public_key"]
	if !ok {
		return fmt.Errorf("could not retrieve public_key from secret %s", secret.Name)
	}

	// ec2 key specific stuff here
	ikpi := ec2.ImportKeyPairInput{
		KeyName:           pointer.String(name),
		PublicKeyMaterial: pubKey,
	}

	ikpiJson, err := json.Marshal(ikpi)
	if err != nil {
		return fmt.Errorf("error marshalling ImportKeyPairInput into JSON: %s", err.Error())
	}
	key.Spec.Key.Raw = ikpiJson
	// end ec2 key specific stuff

	if key.UID == "" {
		err = req.Client.Create(req.Ctx, key)
	} else {
		err = req.Client.Update(req.Ctx, key)
	}

	if err != nil {
		return err
	}

	return nil
}

func GetKeyPair(req router.Request) (*v1alpha1.KeyPair, error) {
	keyList := &v1alpha1.KeyPairList{}
	lo := &client.ListOptions{
		Namespace:     req.Object.GetNamespace(),
		LabelSelector: labels.VMLabelSelector(req.Object.GetName()),
	}

	err := req.Client.List(req.Ctx, keyList, lo)
	if err != nil {
		return nil, err
	}

	if len(keyList.Items) > 0 {
		return &keyList.Items[0], nil
	}

	return nil, errors.NewNotFoundError("could not find any keys for virtualmachine %s", req.Object.GetName())
}
