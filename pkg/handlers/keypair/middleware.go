package keypair

import (
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
)

func KeyPairCreated(next router.Handler) router.Handler {
	return router.HandlerFunc(func(req router.Request, resp router.Response) error {
		vm := req.Object.(*v1.VirtualMachine)

		kp, err := GetKeyPair(req)
		if err != nil {
			return fmt.Errorf("error retrieving keypair for virtualmachine %s: %s", vm.GetName(), err.Error())
		}

		if kp == nil {
			return nil
		}

		return next.Handle(req, resp)
	})

}

// KeyPairImported ensures that the keypair we are operating on
// has not yet been created in ec2. This is tracked by the ConditionKeyPairImported
// condition in the status of the object.
func KeyPairImported(wantImported bool) router.Middleware {
	return func(next router.Handler) router.Handler {
		return router.HandlerFunc(func(req router.Request, resp router.Response) error {
			kp := req.Object.(*v1alpha1.KeyPair)

			if v1alpha1.ConditionKeyPairImported.IsTrue(kp) && !wantImported {
				// already imported, and we don't want it to be
				return nil
			} else if v1alpha1.ConditionKeyPairImported.IsFalse(kp) && wantImported {
				// not imported, and we want it to be
				return nil
			}

			return next.Handle(req, resp)
		})
	}
}

// EnsureKeyPairStatus ensures that the keypair status has been initialized with
// default values.
func EnsureKeyPairStatus(req router.Request, resp router.Response) error {
	kp := req.Object.(*v1alpha1.KeyPair)

	if len(kp.Status.Conditions) == 0 {
		v1alpha1.ConditionKeyPairImported.SetStatus(kp, "unknown")
	}

	resp.Objects(kp)

	return nil
}
