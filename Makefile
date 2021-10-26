all         : test cover lint

test        :
	@echo "Testing..."
	@go test -race -count=1 -bench=. -benchmem -cover -coverprofile=.coverprofile ./...
	@echo ""

cover       :
	@echo "Check coverage..."
	@go tool cover -func=.coverprofile | tail -n 1 | awk '{print "Total coverage:", $$3;}'
	@test `go tool cover -func=.coverprofile | tail -n 1 | awk '{print $$3;}' | sed 's/\..*//'` -ge 83
	@echo ""

lint        :
	@echo "Linting..."
	@golangci-lint run
	@echo "PASS"
	@echo ""
