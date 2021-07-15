Simple test of (experimental) JWKS functionality

We have to use a fixed CA because the fingerprint is inserted into the AWS WebIdentity configuration.

ca.crt & ca.key generated with:

```
openssl req -new -newkey rsa:512 -days 3650 -nodes -x509 -subj "/CN=kubernetes" -keyout ca.key -out ca.crt -config <(cat /etc/ssl/openssl.cnf <(printf "[ v3_ca ]\nkeyUsage = critical,keyCertSign,cRLSign"))
```
