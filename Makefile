default:
	@clear
	@output=$$(go test ./...) || echo "$$output"
	@date +"[ %T ]"
short:
	@clear
	@output=$$(go test -short ./...) || echo "$$output" | grep -Ev "^(ok|\\?)"
	@date +"[ %T ]"
examples:
	@cd examples/htmxform && go test . -update
	for d in examples/*/; do (cd "$$d" && go test .); done

.PHONY: examples
