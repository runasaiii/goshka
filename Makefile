run:
	docker compose up --build

clean:
	docker compose down -v

logs:
	docker compose logs -f web-app
	