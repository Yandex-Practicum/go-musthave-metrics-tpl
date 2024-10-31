make create-db:
	docker run --rm --name postgres -p "5432:5432" -e POSTGRES_DB=admin -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=admin -d "postgres:14-bullseye"
	@sleep 2