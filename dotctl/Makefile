BINARY := dotctl
INSTALL_DIR := $(HOME)/.local/bin
SYSTEMD_DIR := $(HOME)/.config/systemd/user

.PHONY: build test lint install install-systemd uninstall-systemd clean

build:
	go build -o $(BINARY) ./cmd/dotctl/

test:
	go test ./... -v

lint:
	go vet ./...

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/$(BINARY)

install-systemd: install
	mkdir -p $(SYSTEMD_DIR)
	cp deploy/dotctl-collect.service $(SYSTEMD_DIR)/
	cp deploy/dotctl-collect.timer $(SYSTEMD_DIR)/
	systemctl --user daemon-reload
	systemctl --user enable --now dotctl-collect.timer
	@echo "Timer installed. Check: systemctl --user status dotctl-collect.timer"

uninstall-systemd:
	systemctl --user disable --now dotctl-collect.timer || true
	rm -f $(SYSTEMD_DIR)/dotctl-collect.service $(SYSTEMD_DIR)/dotctl-collect.timer
	systemctl --user daemon-reload

clean:
	rm -f $(BINARY)
