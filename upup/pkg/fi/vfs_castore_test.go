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

package fi

import (
	"math/big"
	"math/rand"
	"strings"
	"testing"
	"time"

	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/util/pkg/vfs"
)

// TestBigInt_Format tests that fmt.Sprintf on a big.Int is the same as Text
func TestBigInt_Format(t *testing.T) {
	rnd := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	var limit big.Int
	limit.Lsh(big.NewInt(1), 100)
	for i := 1; i < 100; i++ {
		var r big.Int
		r.Rand(rnd, &limit)
		s1 := r.String()
		s2 := r.Text(10)

		if s1 != s2 {
			t.Logf("%s\n", s1)
			t.Fatalf("%s not the same as %s", s1, s2)
		}
	}
}

func TestVFSCAStoreRoundTrip(t *testing.T) {
	vfs.Context.ResetMemfsContext(true)

	basePath, err := vfs.Context.BuildVfsPath("memfs://tests")
	if err != nil {
		t.Fatalf("error building vfspath: %v", err)
	}

	s := &VFSCAStore{
		basedir:   basePath,
		cachedCAs: make(map[string]*cachedEntry),
	}

	privateKeyData := "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA4JwpEprZ5n8RIEt6jT2lAh+UDgRgx/4px21gjgywQivYHVxH\nAZexVb/E9pBa9Q2G9B1Q7TCO7YsUVRQy4JMDZVt+McFnWVwexnqBYFNcVjkEmDgA\ngvCYGE0P9d/RwRL4KuLHo+u6fv7P0jXMN+CpOxyLhYZZNa0ZOZDHsSiJSQSj9WGF\nGHrbCf0KVDpKieR1uBqHrRO+mLR5zkX2L58m74kjK4dsBhmjeq/7OAoTmiG2QgJ/\nP2IjyhiA2mRqY+hl55lwEUV/0yHYEkJC8LdGkwwZz2eF77aSPGmi/A2CSKgMwDTx\n9m+P7jcpWreYw6NG9BueGoDIve/tgFKwvVFF6QIDAQABAoIBAA0ktjaTfyrAxsTI\nBezb7Zr5NBW55dvuII299cd6MJo+rI/TRYhvUv48kY8IFXp/hyUjzgeDLunxmIf9\n/Zgsoic9Ol44/g45mMduhcGYPzAAeCdcJ5OB9rR9VfDCXyjYLlN8H8iU0734tTqM\n0V13tQ9zdSqkGPZOIcq/kR/pylbOZaQMe97BTlsAnOMSMKDgnftY4122Lq3GYy+t\nvpr+bKVaQZwvkLoSU3rECCaKaghgwCyX7jft9aEkhdJv+KlwbsGY6WErvxOaLWHd\ncuMQjGapY1Fa/4UD00mvrA260NyKfzrp6+P46RrVMwEYRJMIQ8YBAk6N6Hh7dc0G\n8Z6i1m0CgYEA9HeCJR0TSwbIQ1bDXUrzpftHuidG5BnSBtax/ND9qIPhR/FBW5nj\n22nwLc48KkyirlfIULd0ae4qVXJn7wfYcuX/cJMLDmSVtlM5Dzmi/91xRiFgIzx1\nAsbBzaFjISP2HpSgL+e9FtSXaaqeZVrflitVhYKUpI/AKV31qGHf04sCgYEA6zTV\n99Sb49Wdlns5IgsfnXl6ToRttB18lfEKcVfjAM4frnkk06JpFAZeR+9GGKUXZHqs\nz2qcplw4d/moCC6p3rYPBMLXsrGNEUFZqBlgz72QA6BBq3X0Cg1Bc2ZbK5VIzwkg\nST2SSux6ccROfgULmN5ZiLOtdUKNEZpFF3i3qtsCgYADT/s7dYFlatobz3kmMnXK\nsfTu2MllHdRys0YGHu7Q8biDuQkhrJwhxPW0KS83g4JQym+0aEfzh36bWcl+u6R7\nKhKj+9oSf9pndgk345gJz35RbPJYh+EuAHNvzdgCAvK6x1jETWeKf6btj5pF1U1i\nQ4QNIw/QiwIXjWZeubTGsQKBgQCbduLu2rLnlyyAaJZM8DlHZyH2gAXbBZpxqU8T\nt9mtkJDUS/KRiEoYGFV9CqS0aXrayVMsDfXY6B/S/UuZjO5u7LtklDzqOf1aKG3Q\ndGXPKibknqqJYH+bnUNjuYYNerETV57lijMGHuSYCf8vwLn3oxBfERRX61M/DU8Z\nworz/QKBgQDCTJI2+jdXg26XuYUmM4XXfnocfzAXhXBULt1nENcogNf1fcptAVtu\nBAiz4/HipQKqoWVUYmxfgbbLRKKLK0s0lOWKbYdVjhEm/m2ZU8wtXTagNwkIGoyq\nY/C1Lox4f1ROJnCjc/hfcOjcxX5M8A8peecHWlVtUPKTJgxQ7oMKcw==\n-----END RSA PRIVATE KEY-----\n"
	privateKey, err := pki.ParsePEMPrivateKey([]byte(privateKeyData))
	if err != nil {
		t.Fatalf("error from ParsePEMPrivateKey: %v", err)
	}

	certData := "-----BEGIN CERTIFICATE-----\nMIIC2DCCAcCgAwIBAgIRALJXAkVj964tq67wMSI8oJQwDQYJKoZIhvcNAQELBQAw\nFTETMBEGA1UEAxMKa3ViZXJuZXRlczAeFw0xNzEyMjcyMzUyNDBaFw0yNzEyMjcy\nMzUyNDBaMBUxEzARBgNVBAMTCmt1YmVybmV0ZXMwggEiMA0GCSqGSIb3DQEBAQUA\nA4IBDwAwggEKAoIBAQDgnCkSmtnmfxEgS3qNPaUCH5QOBGDH/inHbWCODLBCK9gd\nXEcBl7FVv8T2kFr1DYb0HVDtMI7tixRVFDLgkwNlW34xwWdZXB7GeoFgU1xWOQSY\nOACC8JgYTQ/139HBEvgq4sej67p+/s/SNcw34Kk7HIuFhlk1rRk5kMexKIlJBKP1\nYYUYetsJ/QpUOkqJ5HW4GoetE76YtHnORfYvnybviSMrh2wGGaN6r/s4ChOaIbZC\nAn8/YiPKGIDaZGpj6GXnmXARRX/TIdgSQkLwt0aTDBnPZ4XvtpI8aaL8DYJIqAzA\nNPH2b4/uNylat5jDo0b0G54agMi97+2AUrC9UUXpAgMBAAGjIzAhMA4GA1UdDwEB\n/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBVGR2r\nhzXzRMU5wriPQAJScszNORvoBpXfZoZ09FIupudFxBVU3d4hV9StKnQgPSGA5XQO\nHE97+BxJDuA/rB5oBUsMBjc7y1cde/T6hmi3rLoEYBSnSudCOXJE4G9/0f8byAJe\nrN8+No1r2VgZvZh6p74TEkXv/l3HBPWM7IdUV0HO9JDhSgOVF1fyQKJxRuLJR8jt\nO6mPH2UX0vMwVa4jvwtkddqk2OAdYQvH9rbDjjbzaiW0KnmdueRo92KHAN7BsDZy\nVpXHpqo1Kzg7D3fpaXCf5si7lqqrdJVXH4JC72zxsPehqgi8eIuqOBkiDWmRxAxh\n8yGeRx9AbknHh4Ia\n-----END CERTIFICATE-----\n"
	cert, err := pki.ParsePEMCertificate([]byte(certData))
	if err != nil {
		t.Fatalf("error from ParsePEMCertificate: %v", err)
	}

	if err := s.StoreKeypair("ca", cert, privateKey); err != nil {
		t.Fatalf("error from StoreKeypair: %v", err)
	}

	paths, err := basePath.ReadTree()
	if err != nil {
		t.Fatalf("error from ReadTree: %v", err)
	}

	pathMap := make(map[string]vfs.Path)
	for _, p := range paths {
		pathMap[p.Path()] = p
	}

	for _, p := range []string{
		"memfs://tests/issued/ca/keyset.yaml",
		"memfs://tests/issued/ca/237054359138908419352140518924933177492.crt",
		"memfs://tests/private/ca/keyset.yaml",
		"memfs://tests/private/ca/237054359138908419352140518924933177492.key",
	} {
		if _, found := pathMap[p]; !found {
			t.Fatalf("file not found: %v", p)
		}
	}

	if len(pathMap) != 4 {
		t.Fatalf("unexpected pathMap: %v", pathMap)
	}

	// Check issued/ca/keyset.yaml round-tripped
	{
		issuedKeysetYaml, err := pathMap["memfs://tests/issued/ca/keyset.yaml"].ReadFile()
		if err != nil {
			t.Fatalf("error reading file memfs://tests/issued/ca/keyset.yaml: %v", err)
		}

		expected := `
apiVersion: kops.k8s.io/v1alpha2
kind: Keyset
metadata:
  creationTimestamp: null
  name: ca
spec:
  keys:
  - id: "237054359138908419352140518924933177492"
    publicMaterial: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMyRENDQWNDZ0F3SUJBZ0lSQUxKWEFrVmo5NjR0cTY3d01TSThvSlF3RFFZSktvWklodmNOQVFFTEJRQXcKRlRFVE1CRUdBMVVFQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4TnpFeU1qY3lNelV5TkRCYUZ3MHlOekV5TWpjeQpNelV5TkRCYU1CVXhFekFSQmdOVkJBTVRDbXQxWW1WeWJtVjBaWE13Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBCkE0SUJEd0F3Z2dFS0FvSUJBUURnbkNrU210bm1meEVnUzNxTlBhVUNINVFPQkdESC9pbkhiV0NPRExCQ0s5Z2QKWEVjQmw3RlZ2OFQya0ZyMURZYjBIVkR0TUk3dGl4UlZGRExna3dObFczNHh3V2RaWEI3R2VvRmdVMXhXT1FTWQpPQUNDOEpnWVRRLzEzOUhCRXZncTRzZWo2N3ArL3MvU05jdzM0S2s3SEl1RmhsazFyUms1a01leEtJbEpCS1AxCllZVVlldHNKL1FwVU9rcUo1SFc0R29ldEU3Nll0SG5PUmZZdm55YnZpU01yaDJ3R0dhTjZyL3M0Q2hPYUliWkMKQW44L1lpUEtHSURhWkdwajZHWG5tWEFSUlgvVElkZ1NRa0x3dDBhVERCblBaNFh2dHBJOGFhTDhEWUpJcUF6QQpOUEgyYjQvdU55bGF0NWpEbzBiMEc1NGFnTWk5NysyQVVyQzlVVVhwQWdNQkFBR2pJekFoTUE0R0ExVWREd0VCCi93UUVBd0lCQmpBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFCVkdSMnIKaHpYelJNVTV3cmlQUUFKU2Nzek5PUnZvQnBYZlpvWjA5Rkl1cHVkRnhCVlUzZDRoVjlTdEtuUWdQU0dBNVhRTwpIRTk3K0J4SkR1QS9yQjVvQlVzTUJqYzd5MWNkZS9UNmhtaTNyTG9FWUJTblN1ZENPWEpFNEc5LzBmOGJ5QUplCnJOOCtObzFyMlZnWnZaaDZwNzRURWtYdi9sM0hCUFdNN0lkVVYwSE85SkRoU2dPVkYxZnlRS0p4UnVMSlI4anQKTzZtUEgyVVgwdk13VmE0anZ3dGtkZHFrMk9BZFlRdkg5cmJEampiemFpVzBLbm1kdWVSbzkyS0hBTjdCc0RaeQpWcFhIcHFvMUt6ZzdEM2ZwYVhDZjVzaTdscXFyZEpWWEg0SkM3Mnp4c1BlaHFnaThlSXVxT0JraURXbVJ4QXhoCjh5R2VSeDlBYmtuSGg0SWEKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
  type: Keypair
`

		if strings.TrimSpace(string(issuedKeysetYaml)) != strings.TrimSpace(expected) {
			t.Fatalf("unexpected issued/ca/keyset.yaml: %q", string(issuedKeysetYaml))
		}

		pool, err := s.FindCertificatePool("ca")
		if err != nil {
			t.Fatalf("error reading certificate pool: %v", err)
		}

		if len(pool.Secondary) != 0 {
			t.Fatalf("unexpected secondary certificates: %v", pool)
		}

		if pool.Primary == nil {
			t.Fatalf("primary certificate was nil: %v", pool)
		}

		roundTrip, err := pool.Primary.AsString()
		if err != nil {
			t.Fatalf("error serializing primary cert: %v", err)
		}

		if roundTrip != certData {
			t.Fatalf("unexpected round-tripped certificate data: %q", roundTrip)
		}
	}

	// Check issued/ca/237054359138908419352140518924933177492.crt round-tripped
	{
		roundTrip, err := pathMap["memfs://tests/issued/ca/237054359138908419352140518924933177492.crt"].ReadFile()
		if err != nil {
			t.Fatalf("error reading file memfs://tests/issued/ca/237054359138908419352140518924933177492.crt: %v", err)
		}

		if string(roundTrip) != certData {
			t.Fatalf("unexpected round-tripped certificate data: %q", string(roundTrip))
		}
	}

	// Check private/ca/keyset.yaml round-tripped
	{
		privateKeysetYaml, err := pathMap["memfs://tests/private/ca/keyset.yaml"].ReadFile()
		if err != nil {
			t.Fatalf("error reading file memfs://tests/private/ca/keyset.yaml: %v", err)
		}

		expected := `
apiVersion: kops.k8s.io/v1alpha2
kind: Keyset
metadata:
  creationTimestamp: null
  name: ca
spec:
  keys:
  - id: "237054359138908419352140518924933177492"
    privateMaterial: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBNEp3cEVwclo1bjhSSUV0NmpUMmxBaCtVRGdSZ3gvNHB4MjFnamd5d1FpdllIVnhICkFaZXhWYi9FOXBCYTlRMkc5QjFRN1RDTzdZc1VWUlF5NEpNRFpWdCtNY0ZuV1Z3ZXhucUJZRk5jVmprRW1EZ0EKZ3ZDWUdFMFA5ZC9Sd1JMNEt1TEhvK3U2ZnY3UDBqWE1OK0NwT3h5TGhZWlpOYTBaT1pESHNTaUpTUVNqOVdHRgpHSHJiQ2YwS1ZEcEtpZVIxdUJxSHJSTyttTFI1emtYMkw1OG03NGtqSzRkc0JobWplcS83T0FvVG1pRzJRZ0ovClAySWp5aGlBMm1ScVkraGw1NWx3RVVWLzB5SFlFa0pDOExkR2t3d1p6MmVGNzdhU1BHbWkvQTJDU0tnTXdEVHgKOW0rUDdqY3BXcmVZdzZORzlCdWVHb0RJdmUvdGdGS3d2VkZGNlFJREFRQUJBb0lCQUEwa3RqYVRmeXJBeHNUSQpCZXpiN1pyNU5CVzU1ZHZ1SUkyOTljZDZNSm8rckkvVFJZaHZVdjQ4a1k4SUZYcC9oeVVqemdlREx1bnhtSWY5Ci9aZ3NvaWM5T2w0NC9nNDVtTWR1aGNHWVB6QUFlQ2RjSjVPQjlyUjlWZkRDWHlqWUxsTjhIOGlVMDczNHRUcU0KMFYxM3RROXpkU3FrR1BaT0ljcS9rUi9weWxiT1phUU1lOTdCVGxzQW5PTVNNS0RnbmZ0WTQxMjJMcTNHWXkrdAp2cHIrYktWYVFad3ZrTG9TVTNyRUNDYUthZ2hnd0N5WDdqZnQ5YUVraGRKditLbHdic0dZNldFcnZ4T2FMV0hkCmN1TVFqR2FwWTFGYS80VUQwMG12ckEyNjBOeUtmenJwNitQNDZSclZNd0VZUkpNSVE4WUJBazZONkhoN2RjMEcKOFo2aTFtMENnWUVBOUhlQ0pSMFRTd2JJUTFiRFhVcnpwZnRIdWlkRzVCblNCdGF4L05EOXFJUGhSL0ZCVzVuagoyMm53TGM0OEtreWlybGZJVUxkMGFlNHFWWEpuN3dmWWN1WC9jSk1MRG1TVnRsTTVEem1pLzkxeFJpRmdJengxCkFzYkJ6YUZqSVNQMkhwU2dMK2U5RnRTWGFhcWVaVnJmbGl0VmhZS1VwSS9BS1YzMXFHSGYwNHNDZ1lFQTZ6VFYKOTlTYjQ5V2RsbnM1SWdzZm5YbDZUb1J0dEIxOGxmRUtjVmZqQU00ZnJua2swNkpwRkFaZVIrOUdHS1VYWkhxcwp6MnFjcGx3NGQvbW9DQzZwM3JZUEJNTFhzckdORVVGWnFCbGd6NzJRQTZCQnEzWDBDZzFCYzJaYks1Vkl6d2tnClNUMlNTdXg2Y2NST2ZnVUxtTjVaaUxPdGRVS05FWnBGRjNpM3F0c0NnWUFEVC9zN2RZRmxhdG9iejNrbU1uWEsKc2ZUdTJNbGxIZFJ5czBZR0h1N1E4YmlEdVFraHJKd2h4UFcwS1M4M2c0SlF5bSswYUVmemgzNmJXY2wrdTZSNwpLaEtqKzlvU2Y5cG5kZ2szNDVnSnozNVJiUEpZaCtFdUFITnZ6ZGdDQXZLNngxakVUV2VLZjZidGo1cEYxVTFpClE0UU5Jdy9RaXdJWGpXWmV1YlRHc1FLQmdRQ2JkdUx1MnJMbmx5eUFhSlpNOERsSFp5SDJnQVhiQlpweHFVOFQKdDltdGtKRFVTL0tSaUVvWUdGVjlDcVMwYVhyYXlWTXNEZlhZNkIvUy9VdVpqTzV1N0x0a2xEenFPZjFhS0czUQpkR1hQS2lia25xcUpZSCtiblVOanVZWU5lckVUVjU3bGlqTUdIdVNZQ2Y4dndMbjNveEJmRVJSWDYxTS9EVThaCndvcnovUUtCZ1FEQ1RKSTIramRYZzI2WHVZVW1NNFhYZm5vY2Z6QVhoWEJVTHQxbkVOY29nTmYxZmNwdEFWdHUKQkFpejQvSGlwUUtxb1dWVVlteGZnYmJMUktLTEswczBsT1dLYllkVmpoRW0vbTJaVTh3dFhUYWdOd2tJR295cQpZL0MxTG94NGYxUk9KbkNqYy9oZmNPamN4WDVNOEE4cGVlY0hXbFZ0VVBLVEpneFE3b01LY3c9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
    publicMaterial: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMyRENDQWNDZ0F3SUJBZ0lSQUxKWEFrVmo5NjR0cTY3d01TSThvSlF3RFFZSktvWklodmNOQVFFTEJRQXcKRlRFVE1CRUdBMVVFQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4TnpFeU1qY3lNelV5TkRCYUZ3MHlOekV5TWpjeQpNelV5TkRCYU1CVXhFekFSQmdOVkJBTVRDbXQxWW1WeWJtVjBaWE13Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBCkE0SUJEd0F3Z2dFS0FvSUJBUURnbkNrU210bm1meEVnUzNxTlBhVUNINVFPQkdESC9pbkhiV0NPRExCQ0s5Z2QKWEVjQmw3RlZ2OFQya0ZyMURZYjBIVkR0TUk3dGl4UlZGRExna3dObFczNHh3V2RaWEI3R2VvRmdVMXhXT1FTWQpPQUNDOEpnWVRRLzEzOUhCRXZncTRzZWo2N3ArL3MvU05jdzM0S2s3SEl1RmhsazFyUms1a01leEtJbEpCS1AxCllZVVlldHNKL1FwVU9rcUo1SFc0R29ldEU3Nll0SG5PUmZZdm55YnZpU01yaDJ3R0dhTjZyL3M0Q2hPYUliWkMKQW44L1lpUEtHSURhWkdwajZHWG5tWEFSUlgvVElkZ1NRa0x3dDBhVERCblBaNFh2dHBJOGFhTDhEWUpJcUF6QQpOUEgyYjQvdU55bGF0NWpEbzBiMEc1NGFnTWk5NysyQVVyQzlVVVhwQWdNQkFBR2pJekFoTUE0R0ExVWREd0VCCi93UUVBd0lCQmpBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFCVkdSMnIKaHpYelJNVTV3cmlQUUFKU2Nzek5PUnZvQnBYZlpvWjA5Rkl1cHVkRnhCVlUzZDRoVjlTdEtuUWdQU0dBNVhRTwpIRTk3K0J4SkR1QS9yQjVvQlVzTUJqYzd5MWNkZS9UNmhtaTNyTG9FWUJTblN1ZENPWEpFNEc5LzBmOGJ5QUplCnJOOCtObzFyMlZnWnZaaDZwNzRURWtYdi9sM0hCUFdNN0lkVVYwSE85SkRoU2dPVkYxZnlRS0p4UnVMSlI4anQKTzZtUEgyVVgwdk13VmE0anZ3dGtkZHFrMk9BZFlRdkg5cmJEampiemFpVzBLbm1kdWVSbzkyS0hBTjdCc0RaeQpWcFhIcHFvMUt6ZzdEM2ZwYVhDZjVzaTdscXFyZEpWWEg0SkM3Mnp4c1BlaHFnaThlSXVxT0JraURXbVJ4QXhoCjh5R2VSeDlBYmtuSGg0SWEKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
  type: Keypair
`

		if strings.TrimSpace(string(privateKeysetYaml)) != strings.TrimSpace(expected) {
			t.Fatalf("unexpected private/ca/keyset.yaml: %q", string(privateKeysetYaml))
		}

		key, err := s.FindPrivateKey("ca")
		if err != nil {
			t.Fatalf("error reading certificate pool: %v", err)
		}

		if key == nil {
			t.Fatalf("private key was nil")
		}

		roundTrip, err := key.AsString()
		if err != nil {
			t.Fatalf("error serializing private key: %v", err)
		}

		if roundTrip != privateKeyData {
			t.Fatalf("unexpected round-tripped private key data: %q", roundTrip)
		}
	}

	// Check private/ca/237054359138908419352140518924933177492.key round-tripped
	{
		roundTrip, err := pathMap["memfs://tests/private/ca/237054359138908419352140518924933177492.key"].ReadFile()
		if err != nil {
			t.Fatalf("error reading file memfs://tests/private/ca/237054359138908419352140518924933177492.key: %v", err)
		}

		if string(roundTrip) != privateKeyData {
			t.Fatalf("unexpected round-tripped private key data: %q", string(roundTrip))
		}
	}

}
