postgres:
	docker run --name postgresbank -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres
createdb:
	docker exec -it postgresbank createdb --username=root --owner=root bank
dropdb:
	docker exec -it postgresbank dropdb bank
migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose up
migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose up 1
migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose down
migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose down 1
new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)
sqlc:
	sqlc generate
test:
	go test -v -cover ./...
psql:
	docker exec -it postgresbank psql -U root -d bank
mockgen:
	mockgen -package mockdb -destination db/mock/store.go main/db/sqlc Store
proto:
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt paths=source_relative \
	--openapiv2_out doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
    proto/*.proto
evans: 
	evans --host 0.0.0.0 --port 5001 -r repl 
redis:
	docker run --name redis -p 6379:6379 -d redis:latest 
.PHONY: createdb dropdb postgres migrateup migrateup1 migratedown migratedown1 new_migration sqlc test psql mockgen proto evans redis