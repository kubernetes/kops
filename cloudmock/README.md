cloudmock is a mock implementation of the AWS APIs.

The goal is to let us test code that interacts with the AWS APIs, without creating actual AWS resources.

While no resources are created, we maintain state so that (for example) after you call `CreateVpc`, a subsequent
call to `DescribeVpcs` will return that VPC.  The end-goal is that we simulate the AWS APIs accurately,
so that we can quickly run test-cases that might otherwise require a lot of time or money to run with real
AWS resources.

In future, we can also do fault injection etc.

Note: The AWS API is very large, and most of it is not implemented.  Functions that are implemented may
not be implemented correctly, particularly around edge-cases (such as error handling).

Typical use: `c := &mockec2.MockEC2{}`.  `MockEC2` implements the EC2 API interface `ec2iface.EC2API`,
so can be used where otherwise you would use a real EC2 client.
