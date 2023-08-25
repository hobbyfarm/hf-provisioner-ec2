package crd

import (
	"github.com/ebauman/crder"
	provisioning_hobbyfarm_io "github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
)

func Setup() []crder.CRD {
	keypair := crder.NewCRD(v1alpha1.KeyPair{}, provisioning_hobbyfarm_io.Group, func(c *crder.CRD) {
		c.WithShortNames("kp")
		c.IsNamespaced(true)
		c.AddVersion(v1alpha1.Version, v1alpha1.KeyPair{}, func(cv *crder.Version) {
			cv.
				IsStored(true).
				IsServed(true).
				WithStatus().
				WithPreserveUnknown()
		})
	})

	instance := crder.NewCRD(v1alpha1.Instance{}, provisioning_hobbyfarm_io.Group, func(c *crder.CRD) {
		c.IsNamespaced(true)
		c.AddVersion(v1alpha1.Version, v1alpha1.Instance{}, func(cv *crder.Version) {
			cv.
				IsStored(true).
				IsServed(true).
				WithStatus().
				WithPreserveUnknown()
		})
	})

	return []crder.CRD{
		*keypair,
		*instance,
	}
}
