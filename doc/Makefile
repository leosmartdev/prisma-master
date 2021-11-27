BUILD_DIRECTORY=../build/documentation_site

dist: docs
all: docs

# Serves a local copy of the documentation at localhost:8000
serve: docs-setup
	cd ../ && mkdocs serve

# Creates the static documentation site in ../build/documentation_site
docs: docs-setup
	cd ../ && export LC_ALL=C.UTF-8; export LANG=C.UTF-8; mkdocs build
	cd ${BUILD_DIRECTORY}/linked-documents && mv tms tms_original && mkdir tms
	cd ${BUILD_DIRECTORY}/linked-documents/tms_original && find . \( -name *.html -o -name *.pdf -o -name *.png -o -name *.jpg -o -name *.gif \) -exec rsync -R {} ../tms/ \;
	cd ${BUILD_DIRECTORY}/linked-documents && rm -rf tms_original

clean:
	-rm ./linked-documents/PRISMA_README.md
	-rm ./linked-documents/CHANGELOG.md
	-rm ./linked-documents/installer.md
	-rm ./linked-documents/tms
	-rmdir ./linked-documents

#
# Symbolic links for all markdown files located outside the docs directory
#

linked-documents:
	mkdir ./linked-documents

linked-documents/CHANGELOG.md:
	cd linked-documents && ln -s ../../CHANGELOG.md CHANGELOG.md

linked-documents/PRISMA_README.md:
	cd linked-documents && ln -s ../../README.md PRISMA_README.md

linked-documents/tms:
	cd linked-documents && ln -s ../../tms tms

linked-documents/installer.md:
	cd linked-documents && ln -s ../../make/installer/README.md installer.md

# Setup documentation links before building.
docs-setup: linked-documents linked-documents/CHANGELOG.md linked-documents/PRISMA_README.md linked-documents/installer.md linked-documents/tms

.PHONY: dist docs-setup serve-docs docs
