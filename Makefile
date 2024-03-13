.PHONY: run-gin run-react run

run-gin:
	@echo "Starting Golang Gin server..."
	go run server/main.go &

run-react:
	@echo "Starting React development server..."
	cd client && npm start

run: run-gin run-react
