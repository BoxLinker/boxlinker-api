
registry:
	docker rm -f boxlinker-registry-registry || true
	docker run -p 5000:5000 -p 5001:5001 --name boxlinker-registry-registry -v `pwd`/config.yml:/etc/docker/registry/config.yml registry:2
