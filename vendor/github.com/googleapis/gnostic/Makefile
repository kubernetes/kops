
build:	
	go install github.com/googleapis/gnostic
	go install github.com/googleapis/gnostic/apps/report
	go install github.com/googleapis/gnostic/plugins/go/gnostic_go_sample
	go install github.com/googleapis/gnostic/plugins/go/gnostic_go_generator/encode_templates
	go generate github.com/googleapis/gnostic/plugins/go/gnostic_go_generator
	go install github.com/googleapis/gnostic/plugins/go/gnostic_go_generator
	rm -f $(GOPATH)/bin/gnostic_go_client $(GOPATH)/bin/gnostic_go_server
	ln -s $(GOPATH)/bin/gnostic_go_generator $(GOPATH)/bin/gnostic_go_client
	ln -s $(GOPATH)/bin/gnostic_go_generator $(GOPATH)/bin/gnostic_go_server

