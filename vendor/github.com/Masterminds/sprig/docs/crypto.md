# Cryptographic and Security Functions

Sprig provides a couple of advanced cryptographic functions.

## sha256sum

The `sha256sum` function receives a string, and computes it's SHA256 digest.

```
sha256sum "Hello world!"
```

The above will compute the SHA 256 sum in an "ASCII armored" format that is
safe to print.

## derivePassword

The `derivePassword` function can be used to derive a specific password based on
some shared "master password" constraints. The algorithm for this is
[well specified](http://masterpasswordapp.com/algorithm.html).

```
derivePassword 1 "long" "password" "user" "example.com"
```

Note that it is considered insecure to store the parts directly in the template.

## generatePrivateKey

The `generatePrivateKey` function generates a new private key encoded into a PEM
block.

It takes one of the values for its first param:

- `ecdsa`: Generate an elyptical curve DSA key (P256)
- `dsa`: Generate a DSA key (L2048N256)
- `rsa`: Generate an RSA 4096 key
