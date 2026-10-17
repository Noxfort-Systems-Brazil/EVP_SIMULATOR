---
tags: [csv, data-exchange, controllers, fleet, export, import]
aliases: [CSV Data Exchange, Controller Import, Trajectory Export]
---

# 📊 CSV Data Exchange (Controller Inventories & Fleet Telemetry)

This document specifies the tabular CSV data schemas and utilities implemented in `internal/infrastructure/csv/`.

⬅️ [Documentation Hub](README.md) | 🏛️ [Architecture](architecture.md) | 🔌 [SNMP Emulation](utmc_and_snmp_emulation.md) | 📖 [API Reference](api_and_protocols_reference.md)

---

## 1. Controller Inventory Schema (`controllers_inventory.csv`)

```csv
controller_id,junction_id,ip_address,port,protocol,phase_count,notes
CTRL_CENTRO_01,J_CENTRO_01,127.0.0.1,16100,UTMC2,4,Main Avenue & 1st St Intersection
CTRL_CENTRO_02,J_CENTRO_02,127.0.0.1,16101,UTMC2,2,Main Avenue & 2nd St Intersection
CTRL_NORTE_01,J_NORTE_01,127.0.0.1,16102,NTCIP1202,3,Hospital Emergency Approach
```

---

## 2. Fleet Trajectory Export Schema (`fleet_telemetry_export.csv`)

```csv
timestamp_ms,vehicle_id,corporation,lat,lon,speed_kmh,siren,preemption_engaged,target_junction
1729188000100,SAMU_01,SAMU,-23.310284,-51.162781,58.4,true,false,J_CENTRO_01
1729188000200,SAMU_01,SAMU,-23.310350,-51.162781,64.2,true,true,J_CENTRO_01
1729188000300,SAMU_01,SAMU,-23.310420,-51.162781,68.0,true,true,J_CENTRO_01
```

---

<div align="center">
  <img src="assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Intelligent Mobility Engineering • EVP Simulator v1.0.0</i>
  <small>SPDX-License-Identifier: <a href="../LICENSE">Apache-2.0</a> • &copy; 2026 Noxfort Systems</small>
</div>
