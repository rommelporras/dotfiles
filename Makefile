.PHONY: build test lint install install-systemd uninstall-systemd clean setup-hooks

build:
	$(MAKE) -C dotctl build

test:
	$(MAKE) -C dotctl test

lint:
	$(MAKE) -C dotctl lint

install:
	$(MAKE) -C dotctl install

install-systemd:
	$(MAKE) -C dotctl install-systemd

uninstall-systemd:
	$(MAKE) -C dotctl uninstall-systemd

clean:
	$(MAKE) -C dotctl clean

setup-hooks:
	ln -sfn $(CURDIR)/hooks/pre-commit $(CURDIR)/.git/hooks/pre-commit
	@echo "Installed gitleaks pre-commit hook"
