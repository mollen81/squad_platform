.PHONY: proto build run clean db-reset

# Параметры локальной базы платформы (переопределяются из окружения)
POSTGRES_CONTAINER ?= squad_postgres
POSTGRES_USER ?= admin
POSTGRES_DB ?= main_db

proto:
	@cd internal/core/proto && protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		event.proto
	@echo "Protobuf code generated successfully"

build:
	@go build -o bin/event-service cmd/main.go
	@echo "Build completed"

run: build
	@./bin/event-service
	@echo "Running event-service"

clean:
	@rm -rf bin/
	@echo "Clean completed"

# db-reset — пересоздать схему event-service из migrations/000001_init.up.sql.
#
# Пока сервис не в проде, миграция одна и правится на месте, а golang-migrate
# не перезапускает уже применённую версию: на базе, где схема накатана, новые
# колонки сами не появятся. Этот сброс удаляет таблицы event-service и запись
# о версии, после чего migrate накатит файл заново.
#
# ВНИМАНИЕ: данные event-service теряются. Таблицы других сервисов не
# затрагиваются. Когда выйдем в прод — вместо сброса появится 000002 и далее.
db-reset:
	@docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 \
		-c 'DROP TABLE IF EXISTS team_members, team, users, events' \
		-c "DELETE FROM schema_migrations WHERE version = 1"
	@echo "Схема event-service удалена. Дальше: docker compose up -d migrate-event event-service"
