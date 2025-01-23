IMAGE_NAME=forum

build:
	# This line starts with a TAB
	docker image build -t $(IMAGE_NAME) .

run:
	# TAB again
	docker run -d -p 8433:8433 --name $(IMAGE_NAME)-container $(IMAGE_NAME)

stop:
	docker stop $(IMAGE_NAME)-container

remove:
	docker rm $(IMAGE_NAME)-container

clean: stop remove
	docker rmi $(IMAGE_NAME)

all: build run
