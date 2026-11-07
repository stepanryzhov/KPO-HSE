
.PHONY: help run demo test clean clean-csv clean-data

help:
	@printf '%s\n' \
		'make run        - запустить программу' \
		'make demo       - открыть демонстрационную ведомость' \
		'make test       - запустить тесты' \
		'make clean-csv  - удалить CSV внутри проекта' \
		'make clean-data - удалить рабочую папку data/' \
		'make clean      - удалить CSV и рабочие данные'

run:
	go run ./cmd/app

demo:
	go run ./cmd/app -demo

test:
	go test ./... -count=1

clean: clean-csv
	rm -rf data

clean-csv:
	find . -type f -name '*.csv' ! -path './.git/*' ! -path './cmd/app/demo/*' -print -delete

clean-data:
	rm -rf data

