build:
	go build -o ./dist/ .
	chmod +x ./dist/bmoc

tidy:
	go mod tidy -compat=1.17

install: build
	cp ./dist/bmoc ~/bin