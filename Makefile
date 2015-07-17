all: deps test
	cd src/amproxy/amproxy && go install
deps:
	cd src/amproxy/amproxy && go get -v
package:
	$(MAKE) -C packaging
test:
	cd src/amproxy/amproxy && go test
	cd src/amproxy/message && go test