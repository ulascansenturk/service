env:
	cp -v .env.example .env

env-test:
	cp -v .env.test .env

build-app:
	@ [ -e .env ] || cp -v .env.example .
	docker-compose build app

build-test:
	docker-compose -f docker-compose.test.yml build app

dev: build-app
	docker-compose run --service-ports --rm app 'bash'

lint:
	docker-compose -f docker-compose.yml run --rm app 'golangci-lint run -v --fix'

console:
	@ [ -e .env ] || cp -v .env.example .
	docker-compose build app
	docker-compose run --rm app 'bash'

clean:
	docker-compose down --remove-orphans --volumes

generate: build-app
	docker-compose run --rm app 'go generate ./...'
	docker-compose run --rm app './openapi/go/gen.sh'
	docker-compose run --rm app 'go run github.com/vektra/mockery/v2@v2.43.0'

create-migration:
	docker-compose run --rm app "db/scripts/create_migration.sh $(name)"

migrate:
	docker-compose run --rm app "db/scripts/migrate.sh"

schema-dump:
	docker-compose run --rm app "db/scripts/dump.sh > db/schema.sql"

migrate-up:
	/app/db/scripts/migrate.sh
