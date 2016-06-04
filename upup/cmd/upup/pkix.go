package main

import (
	"bytes"
	"crypto/x509/pkix"
	"fmt"
	"strings"
)

func pkixNameToString(name *pkix.Name) string {
	seq := name.ToRDNSequence()
	var s bytes.Buffer
	for _, rdnSet := range seq {
		for _, rdn := range rdnSet {
			if s.Len() != 0 {
				s.WriteString(",")
			}
			key := ""
			t := rdn.Type
			if len(t) == 4 && t[0] == 2 && t[1] == 5 && t[2] == 4 {
				switch t[3] {
				case 3:
					key = "cn"
				case 5:
					key = "serial"
				case 6:
					key = "c"
				case 7:
					key = "l"
				case 10:
					key = "o"
				case 11:
					key = "ou"
				}
			}
			if key == "" {
				key = t.String()
			}
			s.WriteString(fmt.Sprintf("%v=%v", key, rdn.Value))
		}
	}
	return s.String()
}

func parsePkixName(s string) (*pkix.Name, error) {
	name := new(pkix.Name)

	tokens := strings.Split(s, ",")
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		kv := strings.SplitN(token, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("unrecognized token (expected k=v): %q", token)
		}
		k := strings.ToLower(kv[0])
		v := kv[1]

		switch k {
		case "cn":
			name.CommonName = v
		default:
			return nil, fmt.Errorf("unrecognized key %q in token %q", k, token)
		}
	}

	return name, nil
}
