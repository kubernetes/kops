package slb

// EndpointMap Endpoint Data
var EndpointMap map[string]string

// EndpointType regional or central
var EndpointType = "regional"

// GetEndpointMap Get Endpoint Data Map
func GetEndpointMap() map[string]string {
	if EndpointMap == nil {
		EndpointMap = map[string]string{
			"cn-shanghai-internal-test-1": "slb.aliyuncs.com",
			"cn-beijing-gov-1":            "slb.aliyuncs.com",
			"cn-shenzhen-su18-b01":        "slb.aliyuncs.com",
			"cn-beijing":                  "slb.aliyuncs.com",
			"cn-shanghai-inner":           "slb.aliyuncs.com",
			"cn-shenzhen-st4-d01":         "slb.aliyuncs.com",
			"cn-haidian-cm12-c01":         "slb.aliyuncs.com",
			"cn-hangzhou-internal-prod-1": "slb.aliyuncs.com",
			"cn-north-2-gov-1":            "slb.aliyuncs.com",
			"cn-yushanfang":               "slb.aliyuncs.com",
			"cn-qingdao":                  "slb.aliyuncs.com",
			"cn-hongkong-finance-pop":     "slb.aliyuncs.com",
			"cn-shanghai":                 "slb.aliyuncs.com",
			"cn-shanghai-finance-1":       "slb.aliyuncs.com",
			"cn-hongkong":                 "slb.aliyuncs.com",
			"cn-beijing-finance-pop":      "slb.aliyuncs.com",
			"cn-wuhan":                    "slb.aliyuncs.com",
			"us-west-1":                   "slb.aliyuncs.com",
			"cn-shenzhen":                 "slb.aliyuncs.com",
			"cn-zhengzhou-nebula-1":       "slb.aliyuncs.com",
			"rus-west-1-pop":              "slb.aliyuncs.com",
			"cn-shanghai-et15-b01":        "slb.aliyuncs.com",
			"cn-hangzhou-bj-b01":          "slb.aliyuncs.com",
			"cn-hangzhou-internal-test-1": "slb.aliyuncs.com",
			"eu-west-1-oxs":               "slb.aliyuncs.com",
			"cn-zhangbei-na61-b01":        "slb.aliyuncs.com",
			"cn-beijing-finance-1":        "slb.aliyuncs.com",
			"cn-hangzhou-internal-test-3": "slb.aliyuncs.com",
			"cn-shenzhen-finance-1":       "slb.aliyuncs.com",
			"cn-hangzhou-internal-test-2": "slb.aliyuncs.com",
			"cn-hangzhou-test-306":        "slb.aliyuncs.com",
			"cn-huhehaote-nebula-1":       "slb-api.cn-qingdao-nebula.aliyuncs.com",
			"cn-shanghai-et2-b01":         "slb.aliyuncs.com",
			"cn-hangzhou-finance":         "slb.aliyuncs.com",
			"ap-southeast-1":              "slb.aliyuncs.com",
			"cn-beijing-nu16-b01":         "slb.aliyuncs.com",
			"cn-edge-1":                   "slb.aliyuncs.com",
			"us-east-1":                   "slb.aliyuncs.com",
			"cn-fujian":                   "slb.aliyuncs.com",
			"ap-northeast-2-pop":          "slb.aliyuncs.com",
			"cn-shenzhen-inner":           "slb.aliyuncs.com",
			"cn-zhangjiakou-na62-a01":     "slb.aliyuncs.com",
			"cn-hangzhou":                 "slb.aliyuncs.com",
		}
	}
	return EndpointMap
}

// GetEndpointType Get Endpoint Type Value
func GetEndpointType() string {
	return EndpointType
}
