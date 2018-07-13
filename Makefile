all:
	cd provider && go build -o ../terraform-provider-scc && cd ..
	cd provisioner && go build -o ../terraform-provisioner-scc && cd ..

