.PHONY: run-swagger-ui

run-swagger-ui:
	docker rm -f $$(docker ps -a -q -f "ancestor=swaggerapi/swagger-ui") || true
	docker run -d -p 80:8080 -e SWAGGER_JSON=/spec.yaml -v $$PWD/spec.yaml:/spec.yaml swaggerapi/swagger-ui