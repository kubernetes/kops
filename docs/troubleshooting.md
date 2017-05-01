# Troubleshooting

Here is where we store common troubleshooting "gotchas" with any notes we have.

All issues need a title, so we can link to them.

## ELBs in us-east-1b and us-east-1c

There is a known issue with trying to run ELBs in these regions.

```
When launching a ELB in K8s, Error creating load balancer (will retry): Failed to create load balancer for service default/servicename: ValidationError: The requested Availability Zone us-east-1b is constrained and cannot be used together with us-east-1c. Please retry your request by not choosing us-east-1b and us-east-1c together
```