package instance

import (
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
)

func InstanceCreated(wantCreated bool) router.Middleware {
	return func(next router.Handler) router.Handler {
		return router.HandlerFunc(func(req router.Request, resp router.Response) error {
			instance := req.Object.(*v1alpha1.Instance)

			if wantCreated && v1alpha1.ConditionInstanceExists.IsTrue(instance) {
				// if we want it to exist, and it does, handle
				return next.Handle(req, resp)
			}

			if !wantCreated && !v1alpha1.ConditionInstanceExists.IsTrue(instance) {
				// if we don't want it to exist, and it doesn't, handle
				return next.Handle(req, resp)
			}

			return nil
		})
	}
}

func EnsureInstanceStatus(req router.Request, resp router.Response) error {
	instance := req.Object.(*v1alpha1.Instance)

	if len(instance.Status.Conditions) == 0 {
		v1alpha1.ConditionInstanceExists.Unknown(instance)
		v1alpha1.ConditionInstanceRunning.Unknown(instance)
		v1alpha1.ConditionInstanceUpdated.Unknown(instance)
	}

	resp.Objects(instance)

	return nil
}
