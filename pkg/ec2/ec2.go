package ec2

import (
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	session "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hobbyfarm/hf-provisioner-shared/config"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var AccessKeyField = "access_key"
var SecretKeyField = "secret_key"
var CredentialSecretField = "cred_secret"

func NewEC2Client(vmName string, req router.Request) (*ec2.EC2, error) {
	sess, err := newAWSSession(vmName, req)
	if err != nil {
		return nil, err
	}

	return ec2.New(sess), nil
}

func newAWSSession(vmName string, req router.Request) (*session.Session, error) {
	creds, err := newCredentials(vmName, req)
	if err != nil {
		return nil, err
	}

	var region string
	if region, err = config.ResolveConfigItemName(vmName, req, "region"); err != nil {
		return nil, fmt.Errorf("must specify valid public aws region: %s", err.Error())
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	})

	return sess, nil
}

func newCredentials(vmName string, req router.Request) (*credentials.Credentials, error) {
	// lookup access_key and secret_key on environment
	// or cred_secret and obtain access_key and secret_key from within that secret

	cred_secret, credErr := config.ResolveConfigItemName(vmName, req, CredentialSecretField)
	if credErr != nil {
		return nil, fmt.Errorf("must specify valid credential secret: %s", credErr.Error())
	}

	secret := &v1.Secret{}
	err := req.Client.Get(req.Ctx, client.ObjectKey{
		Name:      cred_secret,
		Namespace: req.Object.GetNamespace(),
	}, secret)
	if err != nil {
		return nil, fmt.Errorf("error retrieving credential secret: %s", err.Error())
	}

	accessKey := string(secret.Data[AccessKeyField])
	secretKey := string(secret.Data[SecretKeyField])

	return credentials.NewStaticCredentials(accessKey, secretKey, ""), nil
}
