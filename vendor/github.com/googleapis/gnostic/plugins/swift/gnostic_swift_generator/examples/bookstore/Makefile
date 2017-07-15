
all:
	rm -f gnostic_swift_generator
	ln -s ../../gnostic_swift_generator 
	gnostic bookstore.json --swift_generator_out=Sources/Bookstore
	swift build

test:
	swift test
