ifneq (,$(wildcard ./.env))
    include .env
    export
endif

SCRIPT_FOLDER = scripts
GOLANG_MIGRATE_VERSION = 4.18.1
GOLANG_MIGRATE_LINUX_ZIP = migrate.linux-amd64.tar.gz
MIGRATION_FOLDER = migrations/postgres
POSTGRESQL_URI = postgresql://$(POSTGRES_USERNAME):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)?sslmode=disable

install-migrate:
	curl -OL https://github.com/golang-migrate/migrate/releases/download/v$(GOLANG_MIGRATE_VERSION)/$(GOLANG_MIGRATE_LINUX_ZIP)
	sudo tar xvf $(GOLANG_MIGRATE_LINUX_ZIP) -C /usr/local/bin/ migrate
	rm -f $(GOLANG_MIGRATE_LINUX_ZIP)


.PHONY: migrate-create
migrate-create:
	$(SCRIPT_FOLDER)/migrate-db create $(name)

.PHONY: migrate-up
migrate-up:
	$(SCRIPT_FOLDER)/migrate-db up

.PHONY: migrate-down
migrate-down:
	$(SCRIPT_FOLDER)/migrate-db down
