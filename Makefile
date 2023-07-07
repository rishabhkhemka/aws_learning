.PHONY: run-swagger-ui start-dynamodb stop-dynamodb

run-swagger-ui:
	docker rm -f $$(docker ps -a -q -f "ancestor=swaggerapi/swagger-ui") || true
	docker run -d -p 80:8080 -e SWAGGER_JSON=/openapi-spec.yaml -v $$PWD/openapi-spec.yaml:/openapi-spec.yaml swaggerapi/swagger-ui

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

build-lambdas:
	$(MAKE) clean-lambdas
	find ./lambdas/ -mindepth 1 -maxdepth 1 -type d -exec sh -c 'cd "{}" && GOOS=linux GOARCH=amd64 go build -o main main.go && zip -r "../../$(basename "{}").zip" ./main' \;

clean-lambdas:
	rm *.zip || true
	find ./lambdas -mindepth 1 -maxdepth 1 -type d -exec sh -c 'cd "{}" && rm main' \;

upload-testdata:
	aws dynamodb batch-write-item \
	--request-items file://mock_data.json \
	--return-consumed-capacity TOTAL \
	--region ap-south-1