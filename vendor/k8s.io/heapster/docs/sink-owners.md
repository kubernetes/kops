Sink Owners
===========

Each sink in Heapster needs to have at least one "owner".  The owner will
be responsible for doing code reviews of pull requests regarding their
sink, and will be a point of contact for issues relating to their sink.

Owners will *not* be responsible for actually triaging issues and pull
requests, but once assigned, they will be responsible for responding to
the issues.

PRs affecting a particular sink generally need to be approved by the sink
owner.  Similarly, PRs affecting a particular sink that have LGTM from the
sink owner will be considered ok-to-merge by the Heapster maintainers
(i.e. sink owners will not have official merge permissions, but the
maintainer's role in this case is just to perform the actual merge).

List of Owners
--------------

- :ok: : has owners
- :sos: : needs owners, will eventually be deprecated and removed without owners
- :new: : in development
- :no_entry: : deprecated, pending removal

| Sink            | Metric             | Event              | Owner(s)                                      | Status         |
| --------------- | ------------------ | -------------------| --------------------------------------------- | -------------- |
| ElasticSearch   | :heavy_check_mark: | :heavy_check_mark: | @AlmogBaku / @andyxning / @huangyuqi          | :ok:           |
| GCM             | :heavy_check_mark: | :x:                | @kubernetes/heapster-maintainers              | :ok:           |
| Hawkular        | :heavy_check_mark: | :x:                | @burmanm / @mwringe                           | :ok:           |
| InfluxDB        | :heavy_check_mark: | :heavy_check_mark: | @kubernetes/heapster-maintainers / @andyxning | :ok:           |
| Metric (memory) | :heavy_check_mark: | :x:                | @kubernetes/heapster-maintainers              | :ok:           |
| Kafka           | :heavy_check_mark: | :x:                | @huangyuqi                                    | :ok:           |
| OpenTSDB        | :heavy_check_mark: | :x:                | @bluebreezecf                                 | :ok:           |
| Riemann         | :heavy_check_mark: | :x: :new:          | @jamtur01 @mcorbin                            | :ok:           |
| Graphite        | :heavy_check_mark: | :x:                | @jsoriano / @theairkit                        | :new: #1341    |
| Wavefront       | :heavy_check_mark: | :x:                | @ezeev                                        | :ok:           |
