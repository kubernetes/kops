/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

//
//import (
//	"fmt"
//
//	"crypto/x509"
//	"k8s.io/klog"
//	"github.com/spf13/cobra"
//	"k8s.io/kops/upup/pkg/fi"
//	"net"
//	"strings"
//)
//
//type CreateSecretsCommand struct {
//	Id             string
//	Type           string
//
//	Usage          string
//	Subject        string
//	AlternateNames []string
//}
//
//var createSecretsCommand CreateSecretsCommand
//
//func init() {
//	cmd := &cobra.Command{
//		Use:   "secret",
//		Short: "Create secrets",
//		Long:  `Create secrets.`,
//		Run: func(cmd *cobra.Command, args []string) {
//			err := createSecretsCommand.Run()
//			if err != nil {
//				exitWithError(err)
//			}
//		},
//	}
//
//	createCmd.AddCommand(cmd)
//
//	cmd.Flags().StringVarP(&createSecretsCommand.Type, "type", "", "", "Type of secret to create")
//	cmd.Flags().StringVarP(&createSecretsCommand.Id, "id", "", "", "Id of secret to create")
//	cmd.Flags().StringVarP(&createSecretsCommand.Usage, "usage", "", "", "Usage of secret (for SSL certificate)")
//	cmd.Flags().StringVarP(&createSecretsCommand.Subject, "subject", "", "", "Subject (for SSL certificate)")
//	cmd.Flags().StringSliceVarP(&createSecretsCommand.AlternateNames, "san", "", nil, "Alternate name (for SSL certificate)")
//}
//
//func (cmd *CreateSecretsCommand) Run() error {
//	if cmd.Id == "" {
//		return fmt.Errorf("id is required")
//	}
//
//	if cmd.Type == "" {
//		return fmt.Errorf("type is required")
//	}
//
//	// TODO: Prompt before replacing?
//	// TODO: Keep history?
//
//	if strings.ToLower(cmd.Type) == strings.ToLower(fi.SecretTypeSecret) {
//		return fmt.Errorf("creating secrets of type %q not (currently) supported", cmd.Type)
//		//{
//		//	secretStore, err := rootCommand.SecretStore()
//		//	if err != nil {
//		//		return err
//		//	}
//		//	secret, err := fi.CreateSecret()
//		//	if err != nil {
//		//		return fmt.Errorf("error creating secret: %v", err)
//		//	}
//		//	_, created, err := secretStore.GetOrCreateSecret(cmd.Id, secret)
//		//	if err != nil {
//		//		return fmt.Errorf("error creating secret: %v", err)
//		//	}
//		//	if !created {
//		//		return fmt.Errorf("secret already exists")
//		//	}
//		//	return nil
//		//}
//	}
//
//	if strings.ToLower(cmd.Type) == strings.ToLower(fi.SecretTypeKeypair) {
//		return fmt.Errorf("creating secrets of type %q not (currently) supported", cmd.Type)
//		//
//		//// TODO: Create a rotate command which keeps the same values?
//		//// Or just do it here a "replace" action - existing=fail, replace or rotate
//		//// TODO: Create a CreateKeypair class, move to fi (this is duplicated code)
//		//{
//		//	if cmd.Subject == "" {
//		//		return fmt.Errorf("subject is required")
//		//	}
//		//
//		//	subject, err := parsePkixName(cmd.Subject)
//		//	if err != nil {
//		//		return fmt.Errorf("Error parsing subject: %v", err)
//		//	}
//		//	template := &x509.Certificate{
//		//		Subject:               *subject,
//		//		BasicConstraintsValid: true,
//		//		IsCA: false,
//		//	}
//		//
//		//	if len(template.Subject.ToRDNSequence()) == 0 {
//		//		return fmt.Errorf("Subject name was empty")
//		//	}
//		//
//		//	switch cmd.Usage {
//		//	case "client":
//		//		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
//		//		template.KeyUsage = x509.KeyUsageDigitalSignature
//		//		break
//		//
//		//	case "server":
//		//		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
//		//		template.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
//		//		break
//		//
//		//	default:
//		//		return fmt.Errorf("unknown usage: %q", cmd.Usage)
//		//	}
//		//
//		//	for _, san := range cmd.AlternateNames {
//		//		san = strings.TrimSpace(san)
//		//		if san == "" {
//		//			continue
//		//		}
//		//		if ip := net.ParseIP(san); ip != nil {
//		//			template.IPAddresses = append(template.IPAddresses, ip)
//		//		} else {
//		//			template.DNSNames = append(template.DNSNames, san)
//		//		}
//		//	}
//		//
//		//	caStore, err := rootCommand.KeyStore()
//		//	if err != nil {
//		//		return err
//		//	}
//		//
//		//	// TODO: Allow resigning of the existing private key?
//		//
//		//	_, _, err = caStore.CreateKeypair(cmd.Id, template)
//		//	if err != nil {
//		//		return fmt.Errorf("error creating keypair %v", err)
//		//	}
//		//	return nil
//	}
//
//	if strings.ToLower(cmd.Type) == strings.ToLower(fi.SecretTypeSSHPublicKey) {
//		return fmt.Errorf("creating secrets of type %q not (currently) supported", cmd.Type)
//	}
//
//	return fmt.Errorf("secret type not known: %q", cmd.Type)
//}
