// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wavefront

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/heapster/metrics/core"
)

const (
	sysSubContainerName = "system.slice/"
)

var excludeTagList = [...]string{"namespace_id", "host_id", "pod_id", "hostname"}

type wavefrontSink struct {
	Conn              net.Conn
	ProxyAddress      string
	ClusterName       string
	Prefix            string
	IncludeLabels     bool
	IncludeContainers bool
	testMode          bool
	testReceivedLines []string
}

func (wfSink *wavefrontSink) Name() string {
	return "Wavefront Sink"
}

func (wfSink *wavefrontSink) Stop() {
	// Do nothing.
	wfSink.Conn.Close()
}

func (wfSink *wavefrontSink) sendLine(line string) {

	if wfSink.testMode {
		wfSink.testReceivedLines = append(wfSink.testReceivedLines, line)
		glog.Infoln(line)
		return
	}
	//if the connection was closed or interrupted - don't cause a panic (we'll retry at next interval)
	defer func() {
		if r := recover(); r != nil {
			//we couldn't write the line so something is wrong with the connection
			wfSink.Conn = nil
		}
	}()
	if wfSink.Conn != nil {
		wfSink.Conn.Write([]byte(line))
	}
}

func (wfSink *wavefrontSink) sendPoint(metricName string, metricValStr string, ts string, source string, tagStr string) {
	metricLine := fmt.Sprintf("%s %s %s source=\"%s\" %s\n", wfSink.cleanMetricName(metricName), metricValStr, ts, source, tagStr)
	wfSink.sendLine(metricLine)
}

func tagsToString(tags map[string]string) string {
	tagStr := ""
	for k, v := range tags {
		//if k != "hostname" {
		if excludeTag(k) == false {
			tagStr += k + "=\"" + v + "\" "
		}
	}
	return tagStr
}

func excludeTag(a string) bool {
	for _, b := range excludeTagList {
		if b == a {
			return true
		}
	}
	return false
}

func (wfSink *wavefrontSink) cleanMetricName(metricName string) string {
	return wfSink.Prefix + strings.Replace(metricName, "/", ".", -1)
}

func (wfSink *wavefrontSink) addLabelTags(ms *core.MetricSet, tags map[string]string) {
	for _, labelName := range sortedLabelKeys(ms.Labels) {
		labelValue := ms.Labels[labelName]
		if labelName == "labels" && wfSink.IncludeLabels {
			for _, label := range strings.Split(labelValue, ",") {
				//labels = app:webproxy,version:latest
				tagParts := strings.SplitN(label, ":", 2)
				if len(tagParts) == 2 {
					tags["label."+tagParts[0]] = tagParts[1]
				}
			}
		} else {
			tags[labelName] = labelValue
		}
	}

}

func (wfSink *wavefrontSink) send(batch *core.DataBatch) {

	metricCounter := 0
	for _, key := range sortedMetricSetKeys(batch.MetricSets) {
		ms := batch.MetricSets[key]
		// Populate tag map
		tags := make(map[string]string)
		// Make sure all metrics are tagged with the cluster name
		tags["cluster"] = wfSink.ClusterName
		// Add pod labels as tags
		wfSink.addLabelTags(ms, tags)
		if strings.Contains(tags["container_name"], sysSubContainerName) {
			//don't send system subcontainers
			continue
		}
		if wfSink.IncludeContainers == false && strings.Contains(tags["type"], "pod_container") {
			// the user doesn't want to include container metrics (only pod and above)
			continue
		}
		for _, metricName := range sortedMetricValueKeys(ms.MetricValues) {
			var metricValStr string
			metricValue := ms.MetricValues[metricName]
			if core.ValueInt64 == metricValue.ValueType {
				metricValStr = fmt.Sprintf("%d", metricValue.IntValue)
			} else if core.ValueFloat == metricValue.ValueType { // W
				metricValStr = fmt.Sprintf("%f", metricValue.FloatValue)
			} else {
				//do nothing for now
				metricValStr = ""
			}
			if metricValStr != "" {
				ts := strconv.FormatInt(batch.Timestamp.Unix(), 10)
				stype := tags["type"]
				source := ""
				if stype == "cluster" {
					source = wfSink.ClusterName
				} else if stype == "ns" {
					source = tags["namespace_name"] + "-ns"
				} else {
					source = tags["hostname"]
				}
				tagStr := tagsToString(tags)
				wfSink.sendPoint(wfSink.cleanMetricName(metricName), metricValStr, ts, source, tagStr)
				metricCounter = metricCounter + 1
			}
		}
		for _, metric := range ms.LabeledMetrics {
			metricName := wfSink.cleanMetricName(metric.Name)
			metricValStr := ""
			if core.ValueInt64 == metric.ValueType {
				metricValStr = fmt.Sprintf("%d", metric.IntValue)
			} else if core.ValueFloat == metric.ValueType { // W
				metricValStr = fmt.Sprintf("%f", metric.FloatValue)
			} else {
				//do nothing for now
				metricValStr = ""
			}
			if metricValStr != "" {
				ts := strconv.FormatInt(batch.Timestamp.Unix(), 10)
				source := tags["hostname"]
				tagStr := tagsToString(tags)
				for labelName, labelValue := range metric.Labels {
					tagStr += labelName + "=\"" + labelValue + "\" "
				}
				metricCounter = metricCounter + 1
				wfSink.sendPoint(metricName, metricValStr, ts, source, tagStr)
			}
		}
	}

}

func (wfSink *wavefrontSink) ExportData(batch *core.DataBatch) {

	if wfSink.testMode {
		//clear lines from last batch
		wfSink.testReceivedLines = wfSink.testReceivedLines[:0]
		wfSink.send(batch)
		return
	}

	//make sure we're Connected before sending a real batch
	err := wfSink.connect()
	if err != nil {
		glog.Warning(err)
	}

	if wfSink.Conn != nil && err == nil {
		wfSink.send(batch)
	}
}

func (wfSink *wavefrontSink) connect() error {
	var err error
	wfSink.Conn, err = net.DialTimeout("tcp", wfSink.ProxyAddress, time.Second*10)
	if err != nil {
		glog.Warningf("Unable to connect to Wavefront proxy at address: %s", wfSink.ProxyAddress)
		return err
	} else {
		glog.Infof("Connected to Wavefront proxy at address: %s", wfSink.ProxyAddress)
		return nil
	}
}

func NewWavefrontSink(uri *url.URL) (core.DataSink, error) {

	storage := &wavefrontSink{
		ProxyAddress:      uri.Scheme + ":" + uri.Opaque,
		ClusterName:       "k8s-cluster",
		Prefix:            "heapster.",
		IncludeLabels:     false,
		IncludeContainers: true,
		testMode:          false,
	}

	vals := uri.Query()
	if len(vals["clusterName"]) > 0 {
		storage.ClusterName = vals["clusterName"][0]
	}
	if len(vals["prefix"]) > 0 {
		storage.Prefix = vals["prefix"][0]
	}
	if len(vals["includeLabels"]) > 0 {
		incLabels := false
		incLabels, err := strconv.ParseBool(vals["includeLabels"][0])
		if err != nil {
			glog.Warning("Unable to parse the includeLabels argument. This argument is a boolean, please pass \"true\" or \"false\"")
			return nil, err
		}
		storage.IncludeLabels = incLabels
	}
	if len(vals["includeContainers"]) > 0 {
		incContainers := false
		incContainers, err := strconv.ParseBool(vals["includeContainers"][0])
		if err != nil {
			glog.Warning("Unable to parse the includeContainers argument. This argument is a boolean, please pass \"true\" or \"false\"")
			return nil, err
		}
		storage.IncludeContainers = incContainers
	}
	if len(vals["testMode"]) > 0 {
		testMode := false
		testMode, err := strconv.ParseBool(vals["testMode"][0])
		if err != nil {
			glog.Warning("Unable to parse the testMode argument. This argument is a boolean, please pass \"true\" or \"false\"")
			return nil, err
		}
		storage.testMode = testMode
	}
	return storage, nil
}

func sortedMetricSetKeys(m map[string]*core.MetricSet) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func sortedLabelKeys(m map[string]string) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func sortedMetricValueKeys(m map[string]core.MetricValue) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
