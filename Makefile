ifndef GITLAB_BUILD_TOKEN
$(error GITLAB_BUILD_TOKEN is not set)
endif

GITLAB_BUILD_ARGS = --build-arg=GITLAB_BUILD_TOKEN=$${GITLAB_BUILD_TOKEN}

builder:
	@docker build . -f build/Dockerfile $(GITLAB_BUILD_ARGS) -t bank-builder

lint: builder
	@docker run --rm -v $$(pwd):/src -w /src bank-builder sh -c 'go fmt ./... | xargs -r -n 1 gofmt -w'
	@docker run --rm -v $$(pwd):/src -w /src bank-builder sh -c 'go vet ./...'

start:
	@docker network create bank-tests; test 1
	@docker run -d --network bank-tests --name mariadb -e MYSQL_RANDOM_ROOT_PASSWORD=yes -e MYSQL_USER=condensat -e MYSQL_PASSWORD=condensat -e MYSQL_DATABASE=condensat -v $$(pwd)/tests/database/permissions.sql:/docker-entrypoint-initdb.d/permissions.sql:ro mariadb:10.3; test 1
	@docker run -d --network bank-tests --name redis -d redis:6-alpine; test 1
	@docker run -d --network bank-tests --name nats -d nats:2-alpine; test 1

tests: builder	
	@docker run --rm --network bank-tests -v $$(pwd):/src -w /src bank-builder sh -c 'go test -v ./... -count=1 -run=Test -parallel=4'

stop:
	@docker stop nats; docker rm nats; test 1
	@docker stop redis; docker rm redis; test 1
	@docker stop mariadb; docker rm mariadb; test 1
	@docker network rm bank-tests; test 1

.PHONY: builder lint start tests stop
