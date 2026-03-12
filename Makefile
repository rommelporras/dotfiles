.PHONY: build test lint install install-systemd uninstall-systemd clean

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
