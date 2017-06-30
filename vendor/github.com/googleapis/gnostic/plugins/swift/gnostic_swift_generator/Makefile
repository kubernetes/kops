
all:
	swift build
	.build/debug/TemplateEncoder > Sources/gnostic_swift_generator/Templates.swift
	swift build
	cp .build/debug/gnostic_swift_generator gnostic_swift_generator
	rm -f gnostic_swift_client gnostic_swift_server
	ln -s gnostic_swift_generator gnostic_swift_client
	ln -s gnostic_swift_generator gnostic_swift_server

clean:
	rm -rf .build Packages
	rm -rf gnostic_swift_client gnostic_swift_service
