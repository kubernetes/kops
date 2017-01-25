### Getting the value of your subdomain NS records in AWS

 - Note your hosted zone ID

```bash
aws route53 list-hosted-zones | jq '.HostedZones[] | select(.Name=="subdomain.kubernetes.com.") | .Id' 

```

  - Note your nameservers for the subdomain

```bash
aws route53 get-hosted-zone --id $HZC | jq .DelegationSet.NameServers
```