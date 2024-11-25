clean:
	@echo "Cleaning..."
	@rm -rf db tmp
	@echo "## Done ##"

post:
	@case $$((RANDOM % 3)) in \
		0) make post-info ;; \
		1) make post-error ;; \
		2) make post-warn ;; \
	esac

post-info:
	@echo "posting an info..."
	@curl -X POST -H "Content-Type: application/json" \
		-d '{"message":"This is an info log", "level":"info"}' \
		localhost:8080/api/v1/logs
	@echo "## Done ##"

post-error:
	@echo "posting an error..."
	@curl -X POST -H "Content-Type: application/json" \
		-d '{"message":"This is a error log", "level":"error"}' \
		localhost:8080/api/v1/logs
	@echo "## Done ##"

post-warn:
	@echo "posting a warning..."
	@curl -X POST -H "Content-Type: application/json" \
		-d '{"message":"This is a warning", "level":"warn"}' \
		localhost:8080/api/v1/logs
	@echo "## Done ##"


