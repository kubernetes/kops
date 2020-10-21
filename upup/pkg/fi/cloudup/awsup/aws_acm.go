package awsup

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/route53"
	"k8s.io/klog/v2"
)

// EnsureCertificate requests an ACM certificate and ensures it is validated
func EnsureCertificate(clusterName, zoneName, region string) (string, error) {

	config := aws.NewConfig().WithRegion(region)

	config = config.WithCredentialsChainVerboseErrors(true)
	config = request.WithRetryer(config, newLoggingRetryer(ClientMaxRetries))
	config.SleepDelay = func(d time.Duration) {
		klog.V(6).Infof("aws request sleeping for %v", d)
		time.Sleep(d)
	}

	requestLogger := newRequestLogger(2)

	sess, err := session.NewSession(config)
	if err != nil {
		return "", err
	}
	acmClient := acm.New(sess, config)
	acmClient.Handlers.Send.PushFront(requestLogger)
	r53Client := route53.New(sess, config)
	r53Client.Handlers.Send.PushFront(requestLogger)

	apiDomain := fmt.Sprintf("api.%v", clusterName)
	certARN, err := findExistingCert(acmClient, apiDomain)
	if err != nil {
		return "", err
	}
	if certARN != "" {
		return certARN, nil
	}

	zoneID, err := getZoneID(r53Client, zoneName)
	if err != nil {
		return "", err
	}

	klog.Infof("Requesting certificate for %v", apiDomain)
	resp, err := acmClient.RequestCertificate(&acm.RequestCertificateInput{
		DomainName:       aws.String(apiDomain),
		ValidationMethod: aws.String(acm.ValidationMethodDns),
	})
	if err != nil {
		return "", err
	}
	certARN = aws.StringValue(resp.CertificateArn)
	validationName, validationValue, err := validationRecord(acmClient, certARN)

	err = setRecord(r53Client, zoneID, validationName, validationValue)
	if err != nil {
		return "", err
	}

	klog.Info("Waiting for certificate to validate")
	acmClient.WaitUntilCertificateValidated(&acm.DescribeCertificateInput{
		CertificateArn: aws.String(certARN),
	})
	klog.Info("Certificate validated")

	return certARN, nil
}

func findExistingCert(client *acm.ACM, domain string) (string, error) {
	var certARN string
	input := &acm.ListCertificatesInput{
		CertificateStatuses: []*string{aws.String(acm.CertificateStatusIssued)},
	}
	err := client.ListCertificatesPages(input,
		func(page *acm.ListCertificatesOutput, lastPage bool) bool {
			for _, cert := range page.CertificateSummaryList {
				if aws.StringValue(cert.DomainName) == domain {
					certARN := aws.StringValue(cert.CertificateArn)
					klog.Infof("Found existing ACM certificate %v", certARN)
				}
			}
			return true
		})
	return certARN, err
}

func validationRecord(client *acm.ACM, certARN string) (string, string, error) {
	var validationName string
	var validationValue string
	tries := 6
	for {
		if tries == 0 {
			return "", "", errors.New("Could not find domain validation options")
		}
		desc, err := client.DescribeCertificate(&acm.DescribeCertificateInput{
			CertificateArn: aws.String(certARN),
		})
		if err != nil {
			return "", "", err
		}
		validation := desc.Certificate.DomainValidationOptions
		if len(validation) != 1 || validation[0].ResourceRecord == nil {
			klog.Info("Unexpected domain validation options, retrying...")
			time.Sleep(time.Duration(5) * time.Second)
			tries--
			continue
		}
		validationName = aws.StringValue(validation[0].ResourceRecord.Name)
		validationValue = aws.StringValue(validation[0].ResourceRecord.Value)
		break
	}
	return validationName, validationValue, nil
}

func getZoneID(client *route53.Route53, zone string) (string, error) {
	resp, err := client.ListHostedZonesByName(&route53.ListHostedZonesByNameInput{
		DNSName: aws.String(zone),
	})
	if err != nil {
		return "", err
	}
	if len(resp.HostedZones) != 1 {
		return "", fmt.Errorf("Unexpected hosted zones found: %v", resp.HostedZones)
	}
	zoneID := aws.StringValue(resp.HostedZones[0].Id)
	klog.Infof("Found HostedZone ID %v", zoneID)
	return zoneID, nil
}

func setRecord(client *route53.Route53, zone, name, value string) error {
	resp, err := client.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zone),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(route53.ChangeActionUpsert),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(name),
						Type: aws.String(route53.RRTypeCname),
						TTL:  aws.Int64(300),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(value),
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}
	klog.Infof("Set validation record %+v", resp.ChangeInfo)
	return nil
}
