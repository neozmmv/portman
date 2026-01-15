# Portman

**Portman** is a simple, safe, and opinionated wrapper for managing `iptables` ports on Linux servers, with first-class support for **Oracle Cloud Infrastructure** and `iptables-persistent`.

It edits `/etc/iptables/rules.v4` in a deterministic way, keeps rules organized, creates automatic backups, and optionally applies changes immediately using `iptables-restore`.

I made this project to simplify the process, especially in Oracle Cloud VPS, where some other solutions simply don't work and can break the firewall completely.

---

## Features

- Simple CLI (`open`, `close`, `status`)
- Manages rules inside a dedicated `#PORTMAN` block
- Idempotent operations (no duplicate rules)
- Automatic backups before every change
- Optional immediate apply with validation
- Works with `iptables-persistent`
- Supports **amd64** and **arm64**

---

## Installation (Ubuntu)

### One-line install

```bash
curl -fsSL https://raw.githubusercontent.com/neozmmv/portman/main/install.sh | sudo bash
```

## Usage

All commands must be run as `root` (or using `sudo`) when modifying or applying firewall rules.

### Open a port

Open a TCP port and apply the rule immediately:

```bash
sudo portman open 443 tcp --apply
```

### Close a port

Close a TCP port and apply the rule immediately:

```bash
sudo portman close 443 tcp --apply
```

### Full instructions

Run `sudo portman` to see the full usage.
