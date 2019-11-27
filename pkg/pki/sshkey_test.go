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

package pki

import (
	"fmt"
	"strings"
	"testing"
)

func checkAWSFingerprintEqual(t *testing.T, publicKey string, fingerprint string) {
	actual, err := ComputeAWSKeyFingerprint(publicKey)
	if err != nil {
		t.Fatalf("Unexpected error computing AWS key fingerprint: %v", err)
	}
	if actual != fingerprint {
		t.Fatalf("Expected fingerprint %q, got %q", fingerprint, actual)
	}
}

func checkAWSFingerprintError(t *testing.T, publicKey string, message string) {
	_, err := ComputeAWSKeyFingerprint(publicKey)
	if err == nil {
		t.Fatalf("Expected error %q computing AWS key fingerprint", message)
	}
	actual := fmt.Sprintf("%v", err)
	if !strings.Contains(actual, message) {
		t.Fatalf("Expected error %q, got %q", message, actual)
	}
}

func Test_AWSFingerprint_RsaKey1(t *testing.T) {
	key := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCySdqIU+FhCWl3BNrAvPaOe5VfL2aCARUWwy91ZP+T7LBwFa9lhdttfjp/VX1D1/PVwntn2EhN079m8c2kfdmiZ/iCHqrLyIGSd+BOiCz0lT47znvANSfxYjLUuKrWWWeaXqerJkOsAD4PHchRLbZGPdbfoBKwtb/WT4GMRQmb9vmiaZYjsfdPPM9KkWI9ECoWFGjGehA8D+iYIPR711kRacb1xdYmnjHqxAZHFsb5L8wDWIeAyhy49cBD+lbzTiioq2xWLorXuFmXh6Do89PgzvHeyCLY6816f/kCX6wIFts8A2eaEHFL4rAOsuh6qHmSxGCR9peSyuRW8DxV725x justin@test"
	checkAWSFingerprintEqual(t, key, "85:a6:f4:64:b7:8f:4a:75:f1:ed:f9:26:1b:67:5f:f2")
}

func Test_AWSFingerprint_RsaKeyEncrypted(t *testing.T) {
	// The private key is encrypted; the public key isn't
	key := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCrLzpTNk5r3RWrzhRuFH8wkOQ+3mOEdaFosPzgDzQtriGU3JZ9Y3UHN4ltUOYUlapyFRaB27Pvyd48GkOSym7ZMn4/kyWn1SvXumJmW8bbX5+pTGK6p3Xu0elBPYMJHWEdZLK5gV6r15uRie9vhxknS9mOzxMcG9gdyyY3DdC3LiiRr6I8wTojP9MsWseZdPPZ5o6tMR/Zp2Q0fOb/DOhNuzunauMos+iu76YPORRFF1PaT1LoLxH7+/HwSX993JDzKytakuCoDFQ2/JvoMxkIvnVIz+MGsLKUZgmxJYQRaIL+fRR+ZBGFrOTqI72NXDmjT7aKjHHxYPfrsSggPh1J justin@machine"
	checkAWSFingerprintEqual(t, key, "c9:c5:05:5e:ea:54:fc:a4:7c:7c:75:5c:d2:71:5e:40")
}

func Test_AWSFingerprint_TrickyWhitespace(t *testing.T) {
	// No name, \r instead of whitespace
	key := "ssh-rsa\rAAAAB3NzaC1yc2EAAAADAQABAAABAQCySdqIU+FhCWl3BNrAvPaOe5VfL2aCARUWwy91ZP+T7LBwFa9lhdttfjp/VX1D1/PVwntn2EhN079m8c2kfdmiZ/iCHqrLyIGSd+BOiCz0lT47znvANSfxYjLUuKrWWWeaXqerJkOsAD4PHchRLbZGPdbfoBKwtb/WT4GMRQmb9vmiaZYjsfdPPM9KkWI9ECoWFGjGehA8D+iYIPR711kRacb1xdYmnjHqxAZHFsb5L8wDWIeAyhy49cBD+lbzTiioq2xWLorXuFmXh6Do89PgzvHeyCLY6816f/kCX6wIFts8A2eaEHFL4rAOsuh6qHmSxGCR9peSyuRW8DxV725x\r"
	checkAWSFingerprintEqual(t, key, "85:a6:f4:64:b7:8f:4a:75:f1:ed:f9:26:1b:67:5f:f2")
}

func Test_AWSFingerprint_DsaKey(t *testing.T) {
	key := "ssh-dss AAAAB3NzaC1kc3MAAACBAIcCTu3vi9rNjsnhCrHeII7jSN6/FmnIdy09pQAsMAGGvCS9HBOteCKbIyYQQ0+Gi76Oui7cJ2VQojdxOxeZPoSP+QYnA+CVYhnowVVLeRA9VBQG3ZLInoXaqe3nR4/OXhY75GmYShBBPTQ+/fWGX9ltoXfygSc4KjhBNudvj75VAAAAFQDiw8A4MhY0aHSX/mtpa7XV8+iS6wAAAIAXyQaxM/dk0o1vBV3H0V0lGhog3mF7EJPdw7jagYvXQP1tAhzNofxZVhXHr4wGfiTQv9j5plDqQzCI/15a6DRyo9zI+zdPTR41W3dGrk56O2/Qxsz3/vNip5OwpOJ88yMmBX9m36gg0WrOXcZDgErhvZWRt5cXa9QjVg/KpxYLPAAAAIB8e5M82IiRLi+k1k4LsELKArQGzVkPgynESfnEXX0TKGiR7PJvBNGaKnPJtJ0Rrc38w/hLTeklroJt9Rdey/NI9b6tc+ur2pmJdnYppnNCm03WszU4oFD/7KIqR84Hf0fMbWd1hRvznpZhngZ505KNsL+ck0+Tlq6Hdhe2baXJcA== justin@machine"
	checkAWSFingerprintError(t, key, "AWS can only import RSA keys")
}

func Test_AWSFingerprint_Ed25519Key(t *testing.T) {
	key := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFpyraYd4rUFftiEKzUO4wKFAgTkXxuJcRZwVcsuZJ8G justin@machine"
	checkAWSFingerprintError(t, key, "AWS can only import RSA keys")
}

func checkOpenSSHFingerprintEqual(t *testing.T, publicKey string, fingerprint string) {
	actual, err := ComputeOpenSSHKeyFingerprint(publicKey)
	if err != nil {
		t.Fatalf("Unexpected error computing OpenSSH key fingerprint: %v", err)
	}
	if actual != fingerprint {
		t.Fatalf("Expected fingerprint %q, got %q", fingerprint, actual)
	}
}

func Test_OpenSSHFingerprint_RsaKey1(t *testing.T) {
	key := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCySdqIU+FhCWl3BNrAvPaOe5VfL2aCARUWwy91ZP+T7LBwFa9lhdttfjp/VX1D1/PVwntn2EhN079m8c2kfdmiZ/iCHqrLyIGSd+BOiCz0lT47znvANSfxYjLUuKrWWWeaXqerJkOsAD4PHchRLbZGPdbfoBKwtb/WT4GMRQmb9vmiaZYjsfdPPM9KkWI9ECoWFGjGehA8D+iYIPR711kRacb1xdYmnjHqxAZHFsb5L8wDWIeAyhy49cBD+lbzTiioq2xWLorXuFmXh6Do89PgzvHeyCLY6816f/kCX6wIFts8A2eaEHFL4rAOsuh6qHmSxGCR9peSyuRW8DxV725x justin@test"
	checkOpenSSHFingerprintEqual(t, key, "be:ba:ec:2b:9e:a0:68:b8:19:6b:9a:26:cc:b1:58:ff")
}
