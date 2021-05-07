package testdata

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kops/tools/triage/testsoup/pkg/caching"
	"k8s.io/kops/tools/triage/testsoup/pkg/gcs"
)

const TestBucketName = "kubernetes-jenkins"

func ListTestJobs() (*TestJobList, error) {
	var result TestJobList
	err := caching.Default.Get("ListTestJobs").GetOrEval(&result, func() error {
		prefixes, err := gcs.ListPrefixes(TestBucketName, "logs/e2e-kops-grid-")
		if err != nil {
			return err
		}
		for _, prefix := range prefixes {
			result.Jobs = append(result.Jobs, &TestJob{
				Bucket: TestBucketName,
				Prefix: prefix,
				Name:   strings.TrimPrefix(prefix, "logs/e2e-"),
			})
		}
		return nil
	})
	return &result, err
}

func ListTestJobRuns(test *TestJob) (*TestJobRunList, error) {
	var result TestJobRunList
	err := caching.Default.Get("ListTestJobRuns", test.Bucket, test.Prefix).GetOrEval(&result, func() error {
		prefixes, err := gcs.ListPrefixes(test.Bucket, test.Prefix)
		if err != nil {
			return err
		}
		for _, prefix := range prefixes {
			result.Runs = append(result.Runs, &TestJobRun{
				Bucket:  test.Bucket,
				Prefix:  prefix,
				JobName: test.Name,
				RunName: strings.TrimPrefix(prefix, test.Prefix),
			})
		}
		return nil
	})
	return &result, err
}

func GetJobRunFile(run *TestJobRun, name string) ([]byte, error) {
	return gcs.ReadObject(run.Bucket, run.Prefix+name)
}

func GetJobRunResults(run *TestJobRun) (*TestJobRunResults, error) {
	var results *TestJobRunResults

	{
		finishedBytes, err := gcs.ReadObject(run.Bucket, run.Prefix+"finished.json")
		if err != nil {
			return nil, fmt.Errorf("error reading finished.json: %w", err)
		}

		var data finishedJSONData
		if err := json.Unmarshal(finishedBytes, &data); err != nil {
			return nil, fmt.Errorf("error parsing finished.json: %w", err)
		}

		results = &TestJobRunResults{
			JobName:   run.JobName,
			RunName:   run.RunName,
			Timestamp: data.Timestamp,
			Passed:    data.Passed,
			Result:    data.Result,
		}
	}

	{
		prowjobBytes, err := gcs.ReadObject(run.Bucket, run.Prefix+"prowjob.json")
		if err != nil {
			return nil, fmt.Errorf("error reading prowjob.json: %w", err)
		}
		u := &unstructured.Unstructured{}
		if err := u.UnmarshalJSON(prowjobBytes); err != nil {
			return nil, fmt.Errorf("failed to parse prowjob.json: %w", err)
		}

		for k, v := range u.GetAnnotations() {
			if strings.HasPrefix(k, "test.kops.k8s.io/") {
				results.Features = append(results.Features, &TestJobFeature{Key: k, Value: v})
			}
		}
	}

	return results, nil
}

type finishedJSONData struct {
	Timestamp uint64 `json:"timestamp"`
	Passed    bool   `json:"passed"`
	// "metadata":{"job-version":"v1.20.2","revision":"v1.20.2"}
	Result string `json:"result"`
}
