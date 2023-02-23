PLUGDIR := ${HOME}/.terraform.d/plugins/terraform.local/getenv/fly/0.0.1/linux_amd64

terraform-provider-fly_v0.0.1: plugdir
		go build -o $(PLUGDIR)/$@ .

plugdir:
	mkdir -p $(PLUGDIR)

clean:
	rm -rf */.terraform*
	rm -rf */terraform*

build:
	make clean
	make terraform-provider-fly_v0.0.1