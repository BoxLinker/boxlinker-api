
PREFIX=index.boxlinker.com/boxlinker

IMAGE_EMAIL=email-server
IMAGE_EMAIL_TAG=latest

IMAGE_ID=user-server
IMAGE_ID_TAG=latest

db:
	docker rm -f boxlinker-db-test || true
	docker run -d --name boxlinker-db-test -v `pwd`/db_data:/var/lib/mysql -p 3306:3306 -e MYSQL_DATABASE=boxlinker -e MYSQL_ROOT_PASSWORD=123456 mysql

rabbitmq:
	docker rm -f boxlinker-email-rabbitmq || true
	docker run -d --name boxlinker-email-rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

email: push-email

build-email:
	cd cmd/email && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o email
	docker build -t ${PREFIX}/${IMAGE_EMAIL}:${IMAGE_EMAIL_TAG} -f Dockerfile.email .

push-email: build-email
	docker push ${PREFIX}/${IMAGE_EMAIL}:${IMAGE_EMAIL_TAG}

build-user:
	cd cmd/user && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o user
	docker build -t ${PREFIX}/${IMAGE_ID}:${IMAGE_ID_TAG} -f Dockerfile.user .

push-user: build-user
	docker push ${PREFIX}/${IMAGE_ID}:${IMAGE_ID_TAG}

user: push-user