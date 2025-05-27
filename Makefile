include .env
MIGRATION_PATH = "./cmd/migrate/migrations"

startdb:
	docker run golang-db-1

migrateup:
	@migrate -path $(MIGRATION_PATH) -database $(DB_ADDR) -verbose up

migratedown:
	@migrate -path $(MIGRATION_PATH) -database $(DB_ADDR) -verbose down

migration:
	@migrate create -seq -ext sql -dir $(MIGRATION_PATH) $(filter-out $@, $(MAKECMDGOALS))

.PHONY:startdb migratedown migrateup migration