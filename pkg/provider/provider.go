package provider

import (
	"github.com/acorn-io/baaah/pkg/log"
	"github.com/ebauman/crder"
	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/crd"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/handlers"
	"github.com/hobbyfarm/hf-provisioner-shared/provider"
	"github.com/sirupsen/logrus"
)

const ProviderName = "aws-ec2"

type EC2 struct {
}

func (EC2) Name() string {
	return ProviderName
}

func (EC2) RouteAdders() []provider.RouteAdder {
	return []provider.RouteAdder{
		handlers.RegisterRoutes(ProviderName),
	}
}

func (EC2) SchemeAdders() []provider.SchemeAdder {
	return []provider.SchemeAdder{
		v1.AddToScheme,       // hobbyfarm
		v1alpha1.AddToScheme, // hf-provisioner-ec2
	}
}

func (EC2) CRDs() []crder.CRD {
	return crd.Setup()
}

func (EC2) Logger() log.Logger {
	return logrus.StandardLogger()
}
