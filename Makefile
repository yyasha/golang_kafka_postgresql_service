gen_docs:
	swag init http_api/*

migrate_create:
	migrate create -ext sql -dir postgres/migrations -seq ${NAME}

graphql_generate:
	rm graph/schema.resolvers.go
	go run github.com/arsmn/fastgql generate
