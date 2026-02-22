---
title: Storage and filesystem management
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Expand beyond basic disk usage stats to full storage management. An appliance
needs to manage mounts, volumes, and filesystem health.

## API Endpoints

```
GET    /storage/disk             - List block devices and partitions
GET    /storage/disk/{name}      - Get disk details (SMART, partitions)

GET    /storage/mount            - List mounted filesystems
POST   /storage/mount            - Mount a filesystem
DELETE /storage/mount/{path}     - Unmount a filesystem

GET    /storage/lvm/vg           - List volume groups
GET    /storage/lvm/lv           - List logical volumes
```

## Operations

- `storage.disk.list.get`, `storage.disk.status.get` (query)
- `storage.mount.list.get` (query)
- `storage.mount.create`, `storage.mount.delete` (modify)
- `storage.lvm.vg.get`, `storage.lvm.lv.get` (query)

## Provider

- `internal/provider/storage/disk/` — `lsblk`, SMART data
- `internal/provider/storage/mount/` — `/proc/mounts`, `mount`/`umount`
- `internal/provider/storage/lvm/` — `vgs`, `lvs` command parsing
- Return types: `DiskInfo`, `MountInfo`, `VolumeGroupInfo`, `LogicalVolumeInfo`

## Notes

- SMART health monitoring is valuable for appliance reliability
- Mount operations are privileged and potentially destructive
- Scopes: `storage:read`, `storage:write`
- Existing `system.disk.get` can remain under `/system/status` as summary
