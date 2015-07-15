pkg = amproxy
version_base = 1.0

ifndef BUILD_NUMBER
version = 0.99.0
else
version = $(version_base).$(BUILD_NUMBER)
endif

topdir = /tmp/$(pkg)-$(version)

all: clean
	@(mkdir $(topdir) && cd $(topdir) && mkdir SRPMS BUILD RPMS SOURCES SPECS tmp)
	@(rsync -av ../ $(topdir)/SOURCES/$(pkg)-$(version))
	@(cd $(topdir)/SOURCES && tar -cvzf $(pkg)-$(version).tar.gz $(pkg)-$(version)/)
	cp redhat/$(pkg).spec $(topdir)/SPECS/
	rpmbuild -bb --define "_topdir $(topdir)" --define "_tmppath $(topdir)/tmp" --define "version $(version)" $(topdir)/SPECS/$(pkg).spec
	cp $(topdir)/RPMS/*/$(pkg)-$(version)-1.el*.*.rpm .
	rm -rf $(topdir)
clean:
	@rm -f *.rpm
	@rm -rf $(topdir) || /bin/true
