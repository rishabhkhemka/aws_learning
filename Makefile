.PHONY: run-swagger-ui start-dynamodb stop-dynamodb

run-swagger-ui:
	docker rm -f $$(docker ps -a -q -f "ancestor=swaggerapi/swagger-ui") || true
	docker run -d -p 80:8080 -e SWAGGER_JSON=/spec.yaml -v $$PWD/spec.yaml:/spec.yaml swaggerapi/swagger-ui

start-dynamodb:
	docker rm -f $$(docker ps -a -q -f "ancestor=amazon/dynamodb-local") || true
	docker run -d -p 8000:8000 \
	  -v /home/rishabh/aws_learning/dynamodb_local_data:/home/dynamodblocal/data \
	  -w /home/dynamodblocal \
	  --name dynamodb-local \
	  amazon/dynamodb-local:latest \
	  -jar DynamoDBLocal.jar -sharedDb -dbPath ./data

stop-dynamodb:
	docker stop dynamodb-local
	docker rm dynamodb-local

docker-clean:
	docker rm -f $$(docker ps -a -q) || true