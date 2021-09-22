run:
	docker run \
		-p 5432:5432 \
		-e POSTGRES_HOST_AUTH_METHOD=trust \
		-v `pwd`/initdb:/docker-entrypoint-initdb.d \
		--rm \
		postgres:13-alpine
