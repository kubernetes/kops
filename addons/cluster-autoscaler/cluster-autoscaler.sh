#!/bin/bash

# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

#Set all the variables in this section
CLUSTER_NAME="myfirstcluster.k8s.local"
CLOUD_PROVIDER=aws
IMAGE=k8s.gcr.io/cluster-autoscaler:v1.1.0
MIN_NODES=2
MAX_NODES=20
AWS_REGION=us-east-1
INSTANCE_GROUP_NAME="nodes"
ASG_NAME="${INSTANCE_GROUP_NAME}.${CLUSTER_NAME}"   #ASG_NAME should be the name of ASG as seen on AWS console.
IAM_ROLE="masters.${CLUSTER_NAME}"                  #Where will the cluster-autoscaler process run? Currently on the master node.
SSL_CERT_PATH="/etc/ssl/certs/ca-certificates.crt"  #(/etc/ssl/certs for gce, /etc/ssl/certs/ca-bundle.crt for RHEL7.X)
#KOPS_STATE_STORE="s3://___"        #KOPS_STATE_STORE might already be set as an environment variable, in which case it doesn't have to be changed.


#Best-effort install script prerequisites, otherwise they will need to be installed manually.
if [[ -f /usr/bin/apt-get && ! -f /usr/bin/jq ]]
then
  sudo apt-get update
  sudo apt-get install -y jq
fi
if [[ -f /bin/yum && ! -f /bin/jq ]]
then
  echo "This may fail if epel cannot be installed. In that case, correct/install epel and retry."
  sudo yum install -y epel-release
  sudo yum install -y jq || exit
fi
if [[ -f /usr/local/bin/brew && ! -f /usr/local/bin/jq ]]
then
  brew install jq || exit
fi


echo "7️⃣  Set up Autoscaling"
echo "   First, we need to update the minSize and maxSize attributes for the kops instancegroup."
echo "   The next command will open the instancegroup config in your default editor, please save and exit the file once you're done…"
sleep 1
kops edit ig $INSTANCE_GROUP_NAME --state ${KOPS_STATE_STORE} --name ${CLUSTER_NAME}
echo "   Running kops update cluster --yes"
kops update cluster --yes --state ${KOPS_STATE_STORE} --name ${CLUSTER_NAME}
printf "\n"

printf "   a) Creating IAM policy to allow aws-cluster-autoscaler access to AWS autoscaling groups…\n"
cat > asg-policy.json << EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "autoscaling:DescribeAutoScalingGroups",
                "autoscaling:DescribeAutoScalingInstances",
                "autoscaling:DescribeLaunchConfigurations",
                "autoscaling:DescribeTags",
                "autoscaling:SetDesiredCapacity",
                "autoscaling:TerminateInstanceInAutoScalingGroup"
            ],
            "Resource": "*"
        }
    ]
}
EOF

ASG_POLICY_NAME=aws-cluster-autoscaler
unset TESTOUTPUT
TESTOUTPUT=$(aws iam list-policies --output json | jq -r '.Policies[] | select(.PolicyName == "aws-cluster-autoscaler") | .Arn')
if [[ $? -eq 0 && -n "$TESTOUTPUT" ]]
then
  printf " ✅  Policy already exists\n"
  ASG_POLICY_ARN=$TESTOUTPUT
else
  printf " ✅  Policy does not yet exist, creating now.\n"
  ASG_POLICY=$(aws iam create-policy --policy-name $ASG_POLICY_NAME --policy-document file://asg-policy.json --output json)
  ASG_POLICY_ARN=$(echo $ASG_POLICY | jq -r '.Policy.Arn')
  printf " ✅ \n"
fi

printf "   b) Attaching policy to IAM Role…\n"
aws iam attach-role-policy --policy-arn $ASG_POLICY_ARN --role-name $IAM_ROLE
printf " ✅ \n"

addon=cluster-autoscaler.yml
manifest_url=https://raw.githubusercontent.com/kubernetes/kops/master/addons/cluster-autoscaler/v1.8.0.yaml

if [[ $(which wget) ]]; then
  wget -O ${addon} ${manifest_url}
elif [[ $(which curl) ]]; then
  curl -s -o ${addon} ${manifest_url}
else
  echo "No curl or wget available. Can't get the manifest."
  exit 1
fi

sed -i -e "s@{{CLOUD_PROVIDER}}@${CLOUD_PROVIDER}@g" "${addon}"
sed -i -e "s@{{IMAGE}}@${IMAGE}@g" "${addon}"
sed -i -e "s@{{MIN_NODES}}@${MIN_NODES}@g" "${addon}"
sed -i -e "s@{{MAX_NODES}}@${MAX_NODES}@g" "${addon}"
sed -i -e "s@{{GROUP_NAME}}@${ASG_NAME}@g" "${addon}"
sed -i -e "s@{{AWS_REGION}}@${AWS_REGION}@g" "${addon}"
sed -i -e "s@{{SSL_CERT_PATH}}@${SSL_CERT_PATH}@g" "${addon}"

kubectl apply -f ${addon}

printf "Done\n"
