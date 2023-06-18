package helpers

import (
	"path"
	"testing"
)

func Test_cacheFilePath(t *testing.T) {
	inputs := []struct {
		kopsStateStore string
		clusterName    string
	}{
		{
			kopsStateStore: "s3://abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk",
			clusterName: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcde." +
				"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk." +
				"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk." +
				"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.com",
		},
	}

	output1 := cacheFilePath(inputs[0].kopsStateStore, inputs[0].clusterName)
	_, file := path.Split(output1)

	if len(file) > 64 {
		t.Errorf("cacheFilePath() got %v, too long(%v)", output1, len(file))
	}
}
