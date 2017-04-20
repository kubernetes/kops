---
title: Troubleshooting
menu_order: 60
---

The command

    weave status

reports on the current status of various Weave Net components, including
the proxy, if it is running:

```
...
        Service: proxy
        Address: tcp://127.0.0.1:12375
...
```

Information on the operation of the proxy can be obtained from the
weaveproxy container logs using:

    docker logs weaveproxy

**See Also**

 * [Troubleshooting Weave](/site/troubleshooting.md)
