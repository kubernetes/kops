#!/bin/bash

echo Starting aws setup

# this does not working us-east-2 yet
# export AWS_REGION=us-west-1

export VPC_ID=`aws ec2 create-vpc --cidr-block 172.20.0.0/16 --query Vpc.VpcId --output text`
aws ec2 create-tags --resources ${VPC_ID} --tags Key=k8s.io/role/imagebuilder,Value=1

export SUBNET_ID=`aws ec2 create-subnet --cidr-block 172.20.1.0/24 --vpc-id ${VPC_ID} --query Subnet.SubnetId --output text`
aws ec2 create-tags --resources ${SUBNET_ID} --tags Key=k8s.io/role/imagebuilder,Value=1


export IGW_ID=`aws ec2 create-internet-gateway --query InternetGateway.InternetGatewayId --output text`
aws ec2 create-tags --resources ${IGW_ID} --tags Key=k8s.io/role/imagebuilder,Value=1

aws ec2 attach-internet-gateway --internet-gateway-id ${IGW_ID} --vpc-id ${VPC_ID}

export RT_ID=`aws ec2 describe-route-tables --filters Name=vpc-id,Values=${VPC_ID} --query RouteTables[].RouteTableId --output text`

export SG_ID=`aws ec2 create-security-group --vpc-id ${VPC_ID} --group-name imagebuilder --description "imagebuilder security group" --query GroupId --output text`
aws ec2 create-tags --resources ${SG_ID} --tags Key=k8s.io/role/imagebuilder,Value=1

aws ec2 associate-route-table --route-table-id ${RT_ID} --subnet-id ${SUBNET_ID}

aws ec2 create-route --route-table-id ${RT_ID} --destination-cidr-block 0.0.0.0/0 --gateway-id ${IGW_ID}

aws ec2 authorize-security-group-ingress  --group-id ${SG_ID} --protocol tcp --port 22 --cidr 0.0.0.0/0

IMGBUILDER_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $IMGBUILDER_DIRECTORY/..

export AWS_REGION=$(aws configure get region)
imagebuilder --config aws.yaml --v=8

echo We are done, and there be dragons!
