package awstasks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func mapEC2TagsToMap(tags []*ec2.Tag) map[string]string {
	if tags == nil {
		return nil
	}
	m := make(map[string]string)
	for _, t := range tags {
		m[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}
	return m
}

func findNameTag(tags []*ec2.Tag) *string {
	for _, tag := range tags {
		if aws.StringValue(tag.Key) == "Name" {
			return tag.Value
		}
	}
	return nil
}
