package cms

import (
	"os"
	"testing"

	"github.com/denverdino/aliyungo/common"
)

const (
	Region = common.Hangzhou
)

var (
	UT_ACCESSKEYID     = os.Getenv("AccessKeyId")
	UT_ACCESSKEYSECRET = os.Getenv("AccessKeySecret")
	UT_SECURITY_TOKEN  = os.Getenv("SecurityToken")
)

func TestCresateAlert(t *testing.T) {
	if UT_ACCESSKEYID == "" {
		t.SkipNow()
	}
	client := NewClient(UT_ACCESSKEYID, UT_ACCESSKEYSECRET)
	if UT_SECURITY_TOKEN != "" {
		client.SetSecurityToken(UT_SECURITY_TOKEN)
	}
	client.SetDebug(true)

	req := `
	{
    "actions":{
        "alertActions":[
        {
                "contactGroups":[],
                "httpNotifyParam":{
                    "type":"http",
                    "method":"GET",
                    "url":"https://cs.console.aliyun.com/hook/trigger?triggerUrl===&secret=]&type=scale_out&step=1"
                },
                "level":4,

            }
        ],
        "effective":"* * 8-22 * * ?",
        "failure":{
            "contactGroups":["云账号报警联系人"],
            "id":"failActionID"
        },
        "ok":{
            "contactGroups":[]
        },
        "silence":"120"
    },
    "condition":{
        "metricName":"CpuUtilization",
        "project":"acs_containerservice",
        "sourceType":"METRIC",
        "dimensionKeys":["userId","clusterId","serviceId"]
    },
    "deepDives":[
        {
            "text":"您的站点信息如下："
        },
        {
            "condition":{
                "metricName":"CpuUtilization"
            }
        }
    ],
    "enable":true,
    "escalations":[
        {
            "expression":"$Average>0.7",
            "level":4,
            "times":1
        }
    ],
    "interval":120,
    "name":"test_alert2",
    "template":true
}
	`

	result, err := client.CreateAlert4Json("acs_custom_xxxx", req)
	if err != nil {
		t.Errorf("CreateAlert encounter error: %v \n", err)
	}
	t.Logf("CreateAlert result : %++v %v \n ", result, err)

	dimension := DimensionRequest{
		UserId:     "xxxx",
		AlertName:  "test_alert2",
		Dimensions: "{\"userId\":\"xxxx\",\"clusterId\":\"xxxxx\",\"serviceId\":\"acsmonitoring_acs-monitoring-agent\"}",
	}
	result, err = client.CreateAlertDimension("acs_custom_xxxx", dimension)
	if err != nil {
		t.Errorf("CreateAlertDimension encounter error: %v \n", err)
	}
	t.Logf("CreateAlertDimension result : %++v  \n ", result)

	result2, err2 := client.GetAlert("acs_custom_xxxx", "test_alert2")
	if err2 != nil {
		t.Errorf("GetAlertList encounter error: %v \n", err2)
	}
	t.Logf("GetAlert result : %++v %v \n ", result2, err2)

}

func TestGetAlertDimension(t *testing.T) {
	if UT_ACCESSKEYID == "" {
		t.SkipNow()
	}
	client := NewClient(UT_ACCESSKEYID, UT_ACCESSKEYSECRET)

	result, err := client.GetDimensions("acs_custom_xxxx", "xxxx")
	t.Logf("GetDimensionsRequest result : %++v %++v %v \n ", result, result.DataPoints[0], err)
}
