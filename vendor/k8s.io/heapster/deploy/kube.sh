#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/kube-config/influxdb"

start() {
  if kubectl apply -f "$DIR/" &> /dev/null; then
    echo "heapster pods have been setup"
  else
    echo "failed to setup heapster pods"
  fi
}

stop() {
  echo -n "heapster resources being removed..."
  kubectl --namespace kube-system delete svc,deployment,rc,rs -l task=monitoring &> /dev/null
  # wait for the pods to disappear.
  while kubectl --namespace kube-system get pods -l "task=monitoring" -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | grep -c . &> /dev/null; do
    echo -n "."
    sleep 2
  done
  echo
  echo "heapster pods have all been removed."
}

case "$1" in
  start)
    start
    ;;
  stop)
    stop
    ;;
  restart)
    stop
    start
    ;;
  *)
    echo "Usage: $0 {start|stop|restart}"
    ;;
esac

exit 0
