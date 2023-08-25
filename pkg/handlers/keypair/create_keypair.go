package keypair

import (
	"encoding/json"
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hobbyfarm/hf-provisioner-ec2/pkg/apis/provisioning.hobbyfarm.io/v1alpha1"
	ec22 "github.com/hobbyfarm/hf-provisioner-ec2/pkg/ec2"
)

func CreateKeyPair(req router.Request, resp router.Response) error {
	kp := req.Object.(*v1alpha1.KeyPair)

	ikpi := &ec2.ImportKeyPairInput{}
	err := json.Unmarshal(kp.Spec.Key.Raw, ikpi)
	if err != nil {
		msg := fmt.Sprintf("error unmarshalling ImportKeyPairInput from JSON: %s", err.Error())
		v1alpha1.ConditionKeyPairImported.False(kp)
		v1alpha1.ConditionKeyPairImported.Message(kp, msg)
		_ = req.Client.Status().Update(req.Ctx, kp)
		return fmt.Errorf(msg)
	}

	// use ikpi plus ec2 client to import keypair
	ec2Client, err := ec22.NewEC2Client(kp.Spec.Machine, req)
	if err != nil {
		msg := fmt.Sprintf("error creating aws ec2 client: %s", err.Error())
		v1alpha1.ConditionKeyPairImported.False(kp)
		v1alpha1.ConditionKeyPairImported.Message(kp, msg)
		_ = req.Client.Status().Update(req.Ctx, kp)
		return fmt.Errorf(msg)
	}

	_, err = ec2Client.ImportKeyPair(ikpi)
	if err != nil && err.(awserr.Error).Code() != "InvalidKeyPair.Duplicate" {
		msg := fmt.Sprintf("error importing keypair: %s", err.Error())
		v1alpha1.ConditionKeyPairImported.False(kp)
		v1alpha1.ConditionKeyPairImported.Message(kp, msg)
		_ = req.Client.Status().Update(req.Ctx, kp)
		return fmt.Errorf(msg)
	}

	var kpInfo *ec2.KeyPairInfo
	res, err := ec2Client.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{
		KeyNames: aws.StringSlice([]string{*ikpi.KeyName}),
	})
	if err != nil {
		msg := fmt.Sprintf("error describing keypair: %s", err.Error())
		v1alpha1.ConditionKeyPairImported.False(kp)
		v1alpha1.ConditionKeyPairImported.Message(kp, msg)
		_ = req.Client.Status().Update(req.Ctx, kp)
		return fmt.Errorf(msg)
	}

	if res.KeyPairs[0] != nil {
		kpInfo = res.KeyPairs[0]
	}

	// keypair imported. marshal response and store
	jsonKp, err := json.Marshal(kpInfo)
	if err != nil {
		v1alpha1.ConditionKeyPairImported.True(kp)
		v1alpha1.ConditionKeyPairImported.Message(kp, fmt.Sprintf("keypair imported, error storing in k8s: %s", err.Error()))
	} else {
		kp.Status.Key.Raw = jsonKp
		v1alpha1.ConditionKeyPairImported.True(kp)
	}

	return req.Client.Status().Update(req.Ctx, kp)
}
