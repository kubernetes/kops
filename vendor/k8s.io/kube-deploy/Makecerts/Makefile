
.PHONY: all clean
all: all.tar

clean:
	rm -f *.tar
	rm -f *.pem
	rm -f *.csr
	rm -f ca.srl
	rm -f openssl.cnf

# config
openssl.cnf: openssl.cnf.in san-extras
	cat openssl.cnf.in san-extras > openssl.cnf

# root CA
ca-key.pem:
	openssl genrsa -out ca-key.pem 2048

ca.pem: ca-key.pem
	openssl req -x509 -new -nodes -key ca-key.pem -days 10000 -out ca.pem -subj "/CN=kube-ca"

root.tar: ca-key.pem ca.pem
	tar cfv root.tar ca-key.pem ca.pem

# apiserver key pair
apiserver-key.pem:
	openssl genrsa -out apiserver-key.pem 2048

apiserver.csr: apiserver-key.pem openssl.cnf
	openssl req -new -key apiserver-key.pem -out apiserver.csr -subj "/CN=kube-apiserver" -config openssl.cnf

apiserver.pem: apiserver.csr ca.pem ca-key.pem openssl.cnf
	openssl x509 -req -in apiserver.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out apiserver.pem -days 365 -extensions v3_req -extfile openssl.cnf

apiserver.tar: apiserver-key.pem apiserver.pem
	tar cfv apiserver.tar apiserver-key.pem apiserver.pem

# kubelet key pair
kubelet-key.pem:
	openssl genrsa -out kubelet-key.pem 2048

kubelet.csr: kubelet-key.pem
	openssl req -new -key kubelet-key.pem -out kubelet.csr -subj "/CN=kubelet"

kubelet.pem: kubelet.csr ca.pem ca-key.pem
	openssl x509 -req -in kubelet.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out kubelet.pem -days 365

kubelet.tar: kubelet-key.pem kubelet.pem
	tar cfv kubelet.tar kubelet-key.pem kubelet.pem

# tars
all.tar: root.tar apiserver.tar kubelet.tar
	tar cfv all.tar root.tar apiserver.tar kubelet.tar
