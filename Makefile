.PHONY: compose-up compose-down run migrations migration-create migration-rollback sqlc

# prepare database and dependencies on docker compose
compose-up:
	docker compose -f compose.yml up -d

# remove docker containers and volumes
compose-down:
	docker compose -f compose.yml down -v

# run project
run:
	air --build.cmd "go build -o ./bin/api ./cmd/api" --build.bin "./bin/api"

# run migrations from go script, in order to load local environments
migrations:
	go run ./cmd/terndotenv

# create a new migration using tern
migration-create:
	tern new --migrations ./internal/store/pgstore/migrations $(name)

# rollback migration
migration-rollback:
	tern migrate --migrations ./internal/store/pgstore/migrations --config ./internal/store/pgstore/migrations/tern.conf --destination -1

# generate queries using SQLc
sqlc:
	sqlc generate -f ./internal/store/pgstore/sqlc.yml