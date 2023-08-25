package main

import (
	"context"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/provider"
	"github.com/hobbyfarm/hf-provisioner-shared/controller"
	"github.com/sirupsen/logrus"
	"log"
)

func main() {
	cx, err := controller.NewController(provider.EC2{})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	logrus.Info("starting controllers")
	if err := cx.Start(ctx); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
}
