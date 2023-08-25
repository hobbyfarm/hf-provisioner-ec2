package keypair

import (
	"encoding/json"
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	ec22 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/ec2"
	"github.com/sirupsen/logrus"
	"time"
)

func KeyPairFinalizer(req router.Request, resp router.Response) error {
	kp := req.Object.(*v1alpha1.KeyPair)

	ec2Client, err := ec2.NewEC2Client(kp.Spec.Machine, req)
	if err != nil {
		return fmt.Errorf("error creating ec2 client: %w", err)
	}

	kpi := &ec22.KeyPairInfo{}
	err = json.Unmarshal(kp.Status.Key.Raw, kpi)
	if err != nil {
		logrus.Errorf("error unmarshalling keypair info from status: %v. this is not likely to "+
			"resolve on its own. releasing finalizer.", err)
		return nil
	}

	_, err = ec2Client.DescribeKeyPairs(&ec22.DescribeKeyPairsInput{
		DryRun:           nil,
		Filters:          nil,
		IncludePublicKey: nil,
		KeyNames:         nil,
		KeyPairIds:       aws.StringSlice([]string{*kpi.KeyPairId}),
	})

	if err != nil {
		if err.(awserr.Error).Code() == "InvalidKeyPair.NotFound" {
			return nil // keypair is already deleted
		} else {
			return fmt.Errorf("error describing keypair: %w", err)
		}
	}

	_, err = ec2Client.DeleteKeyPair(&ec22.DeleteKeyPairInput{
		KeyPairId: kpi.KeyPairId,
	})

	resp.RetryAfter(5 * time.Second)
	return nil
}
