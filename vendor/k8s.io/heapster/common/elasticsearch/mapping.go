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
package elasticsearch

import (
	"strings"

	"k8s.io/heapster/metrics/core"
)

func MetricFamilyTimestamp(metricFamily core.MetricFamily) string {
	return strings.Title(string(metricFamily)) + "MetricsTimestamp"
}
func metricFamilySchema(metricFamily core.MetricFamily) string {
	metricSchemas := []string{}
	for _, metric := range core.MetricFamilies[metricFamily] {
		metricSchemas = append(metricSchemas,
			`"`+metric.Name+`": {
  "properties": {
    "value": {
      "type": "double"
    }
  }
}`,
		)
	}

	return customMetricTypeSchema(string(metricFamily),
		`"`+MetricFamilyTimestamp(metricFamily)+`": {
  "type": "date",
  "format": "strict_date_optional_time||epoch_millis"
},
"Metrics": {
  "properties": {
  `+strings.Join(metricSchemas, ",\r\n")+`
  }
}
`,
	)
}

func customMetricTypeSchema(typeName string, customSchema string) string {
	return `"` + typeName + `": {
  "properties": {
    "MetricsTags": {
      "properties": {
        "container_base_image": {
          "type": "string",
          "index": "analyzed",
          "fields": {
            "raw": {
              "type": "string",
              "index": "not_analyzed"
            }
          }
        },
        "container_name": {
          "type": "string",
          "index": "analyzed",
          "fields": {
            "raw": {
              "type": "string",
              "index": "not_analyzed"
            }
          }
        },
        "host_id": {
          "type": "string",
          "index": "not_analyzed"
        },
        "hostname": {
          "type": "string",
          "index": "analyzed",
          "fields": {
            "raw": {
              "type": "string",
              "index": "not_analyzed"
            }
          }
        },
        "labels": {
          "type": "string",
          "index": "analyzed",
          "fields": {
            "raw": {
              "type": "string",
              "index": "not_analyzed"
            }
          }
        },
        "namespace_id": {
          "type": "string",
          "index": "not_analyzed"
        },
        "namespace_name": {
          "type": "string",
          "fields": {
            "raw": {
              "type": "string",
              "index": "not_analyzed"
            }
          }
        },
        "nodename": {
          "type": "string",
          "index": "analyzed",
          "fields": {
            "raw": {
              "type": "string",
              "index": "not_analyzed"
            }
          }
        },
        "cluster_name": {
          "type": "string",
          "index": "not_analyzed"
        },
        "pod_id": {
          "type": "string",
          "index": "not_analyzed"
        },
        "pod_name": {
          "type": "string",
          "index": "analyzed",
          "fields": {
            "raw": {
              "type": "string",
              "index": "not_analyzed"
            }
          }
        },
        "pod_namespace": {
          "type": "string",
          "fields": {
            "raw": {
              "type": "string",
              "index": "not_analyzed"
            }
          }
        },
        "resource_id": {
          "type": "string",
          "index": "not_analyzed"
        },
        "type": {
          "type": "string",
          "index": "not_analyzed"
        }
      }
    },
    ` + customSchema + `
  }
}`
}

var mapping = `{
  "mappings": {
    "_default_": {
      "_all": {
        "enabled": false
      }
    },
    ` + metricFamilySchema(core.MetricFamilyCpu) + `,
    ` + metricFamilySchema(core.MetricFamilyFilesystem) + `,
    ` + metricFamilySchema(core.MetricFamilyMemory) + `,
    ` + metricFamilySchema(core.MetricFamilyNetwork) + `,
    ` + customMetricTypeSchema(core.MetricFamilyGeneral,
	`"MetricsName": {
  "type": "string",
  "index": "analyzed",
  "fields": {
    "raw": {
      "type": "string",
      "index": "not_analyzed"
    }
  }
},
"GeneralMetricsTimestamp": {
"type": "date",
"format": "strict_date_optional_time||epoch_millis"
},
"MetricsValue": {
  "properties": {
    "value": {
      "type": "double"
    }
  }
}`) + `,

    "events": {
      "properties": {
        "EventTags": {
          "properties": {
            "eventID": {
              "type": "string",
              "index": "not_analyzed"
            },
            "cluster_name": {
              "type": "string",
              "index": "not_analyzed"
            },
            "hostname": {
              "type": "string",
              "index": "analyzed",
              "fields": {
                "raw": {
                  "type": "string",
                  "index": "not_analyzed"
                }
              }
            },
            "pod_id": {
              "type": "string",
              "index": "not_analyzed"
            },
            "pod_name": {
              "type": "string",
              "index": "analyzed",
              "fields": {
                "raw": {
                  "type": "string",
                  "index": "not_analyzed"
                }
              }
            }
          }
        },
        "InvolvedObject": {
          "properties": {
            "apiVersion": {
              "type": "string",
              "index": "not_analyzed"
            },
            "fieldPath": {
              "type": "string",
              "fields": {
                "raw": {
                  "type": "string",
                  "index": "not_analyzed"
                }
              }
            },
            "kind": {
              "type": "string",
              "index": "not_analyzed"
            },
            "name": {
              "type": "string",
              "fields": {
                "raw": {
                  "type": "string",
                  "index": "not_analyzed"
                }
              }
            },
            "namespace": {
              "type": "string",
              "fields": {
                "raw": {
                  "type": "string",
                  "index": "not_analyzed"
                }
              }
            },
            "resourceVersion": {
              "type": "string",
              "index": "not_analyzed"
            },
            "uid": {
              "type": "string",
              "index": "not_analyzed"
            }
          }
        },
        "FirstOccurrenceTimestamp": {
          "type": "date",
          "format": "strict_date_optional_time||epoch_millis"
        },
        "LastOccurrenceTimestamp": {
          "type": "date",
          "format": "strict_date_optional_time||epoch_millis"
        },
        "Type": {
          "type": "string",
          "index": "not_analyzed"
        },
        "Message": {
          "type": "string",
          "fields": {
            "raw": {
              "type": "string",
              "index": "not_analyzed"
            }
          }
        },
        "Reason": {
          "type": "string",
          "index": "not_analyzed"
        },
        "Count": {
          "type": "long"
        },
        "Metadata": {
          "properties": {
            "creationTimestamp": {
              "type": "date",
              "format": "strict_date_optional_time||epoch_millis"
            },
            "name": {
              "type": "string",
              "fields": {
                "raw": {
                  "type": "string",
                  "index": "not_analyzed"
                }
              }
            },
            "namespace": {
              "type": "string",
              "fields": {
                "raw": {
                  "type": "string",
                  "index": "not_analyzed"
                }
              }
            },
            "resourceVersion": {
              "type": "string",
              "index": "not_analyzed"
            },
            "selfLink": {
              "type": "string",
              "index": "not_analyzed"
            },
            "uid": {
              "type": "string",
              "index": "not_analyzed"
            }
          }
        },
        "Source": {
          "properties": {
            "component": {
              "type": "string",
              "index": "not_analyzed"
            },
            "host": {
              "type": "string",
              "fields": {
                "raw": {
                  "type": "string",
                  "index": "not_analyzed"
                }
              }
            }
          }
        }
      }
    }
  }
}`
