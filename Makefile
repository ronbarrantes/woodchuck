clean:
	@echo "Cleaning..."
	@rm -rf db tmp
	@echo "## Done ##"

post:
	@echo "posting..."
	curl -X POST -H "Content-Type: application/json" -d '{"message":"Error", "level":"error"}' localhost:8080/api/v1/log
	@echo "## Done ##"
