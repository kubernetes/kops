# Grafana Dashboards using InfluxDB fields

If you're using `withfields=true` parameters in InfluxDB sink URL,
the storage schema changes in InfluxDB.
So you need to use those Grafana Dashboards:
* If you're using heapster:
    * pods.json
    * cluster.json
* If you're using eventer:
    * events.json


_NOTE: These files are **not** built into the grafana image, they are only for reference_

More info [here](/docs/storage-schema.md)
