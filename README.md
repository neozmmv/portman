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

## Releases / versioning

Releases are created from Git tags using semantic versioning.

- Create a new version tag locally:
	- `git tag v1.0.0`
- Push the tag to GitHub:
	- `git push origin v1.0.0`

This triggers the GitHub Actions release workflow and creates a release named `Portman v1.0.0` with the built artifacts.

## Why not UFW?

While **UFW** is a popular and convenient firewall manager on Ubuntu, it is not well suited for all environments, especially **Oracle Cloud Infrastructure (OCI)**.

### The core difference

UFW is a **high-level firewall manager**. It assumes full ownership of the firewall and dynamically generates and manages its own iptables chains and rules.

Portman, on the other hand, operates at a **low level**. It works directly with the system’s persistent iptables configuration and makes minimal, deterministic changes without attempting to take control of the entire firewall.

---

### Firewall behavior on Oracle Cloud

Oracle Cloud instances ship with a **preconfigured iptables setup** that includes critical system rules managed by Oracle itself. These rules are required for:

- Instance metadata access
- Internal DNS resolution
- DHCP networking
- Block volume and infrastructure services

These rules are loaded at boot time from `/etc/iptables/rules.v4` and include custom chains such as `InstanceServices`.

---

### Why UFW causes problems on OCI

When UFW is enabled on an Oracle Cloud instance, it may:

- Reorder or override existing rules
- Modify default policies
- Introduce its own chains (`ufw-*`)
- Rewrite or conflict with the persistent iptables configuration
- Break Oracle-managed chains and services

In practice, this often results in:

- Loss of network connectivity after reboot
- Broken DNS or metadata access
- Instances becoming unreachable

---

### Portman’s approach

Portman was designed specifically to avoid these issues:

- It never modifies rules outside its own managed block
- It does not change default policies
- It does not create or manage custom chains
- It uses the same mechanism as the system (`iptables-restore`)
- It validates rules before applying them

All rules managed by Portman are stored inside a dedicated block:

```text
#PORTMAN BEGIN
-A INPUT -p tcp -m tcp --dport 443 -j ACCEPT
#PORTMAN END
```
