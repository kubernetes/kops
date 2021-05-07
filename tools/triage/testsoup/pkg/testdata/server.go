package testdata

import (
	context "context"

	"k8s.io/klog/v2"
)

type Server struct {
	UnimplementedTestDataServer
}

var _ TestDataServer = &Server{}

func (s *Server) ListTestJobs(request *ListTestJobsRequest, stream TestData_ListTestJobsServer) error {
	klog.Infof("ListTestJobs %v", request)
	jobs, err := ListTestJobs()
	if err != nil {
		return err
	}
	for _, job := range jobs.Jobs {
		if err := stream.Send(job); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) ListTestJobRuns(request *ListTestJobRunsRequest, stream TestData_ListTestJobRunsServer) error {
	klog.Infof("ListTestJobRuns %v", request)
	runs, err := ListTestJobRuns(request.Job)
	if err != nil {
		return err
	}
	for _, run := range runs.Runs {
		if err := stream.Send(run); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) GetTestJobRunResults(ctx context.Context, request *GetTestJobRunResultsRequest) (*TestJobRunResults, error) {
	klog.Infof("GetTestJobRunResults %v", request)
	return GetJobRunResults(request.Run)
}
