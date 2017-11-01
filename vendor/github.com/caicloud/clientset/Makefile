all: gen

clean:
	rm -rf informers kubernetes lister

gen: clean
	cp -r expansions/* ./
	sh cmd/autogenerate.sh

