---
tags: [csv, data-exchange, controllers, fleet, export, import]
aliases: [Exportação CSV, Importação de Dados, CSV Data Exchanger]
---

# 📊 Intercâmbio de Dados em CSV (Controladores e Frotas)

Este documento especifica os esquemas e utilitários de importação e exportação de dados tabulares em formato CSV implementados em `internal/infrastructure/csv/`.

⬅️ [Central de Documentação](README.md) | 🏛️ [Arquitetura](architecture.md) | 🔌 [Emulação SNMP](utmc_and_snmp_emulation.md) | 📖 [Referência de APIs](api_and_protocols_reference.md)

---

## 1. Inventário de Controladores Semafóricos (`controllers_inventory.csv`)

O arquivo de inventário associa os cruzamentos da malha SUMO a endereços de rede de controladores físicos ou emulados:

```csv
controller_id,junction_id,ip_address,port,protocol,phase_count,notes
CTRL_CENTRO_01,J_CENTRO_01,127.0.0.1,16100,UTMC2,4,Cruzamento Av Central com Rua 1
CTRL_CENTRO_02,J_CENTRO_02,127.0.0.1,16101,UTMC2,2,Cruzamento Av Central com Rua 2
CTRL_NORTE_01,J_NORTE_01,127.0.0.1,16102,NTCIP1202,3,Acesso ao Hospital Regional
```

### 1.1 Descrição dos Campos
- `controller_id`: Código alfanumérico do controlador.
- `junction_id`: ID exato do `<junction>` no arquivo `.net.xml` do SUMO.
- `ip_address`: Endereço IP do agente SNMP do controlador.
- `port`: Porta UDP do serviço SNMP (ex: `16100`).
- `protocol`: Protocolo semafórico utilizado (`UTMC2` ou `NTCIP1202`).
- `phase_count`: Quantidade de estágios/fases programadas no equipamento.
- `notes`: Observações operacionais para o operador de tráfego.

---

## 2. Exportação de Trajetórias de Emergência (`fleet_telemetry_export.csv`)

Permite a auditoria forense das missões de socorro simuladas:

```csv
timestamp_ms,vehicle_id,corporation,lat,lon,speed_kmh,siren,preemption_engaged,target_junction
1729188000100,SAMU_01,SAMU,-23.310284,-51.162781,58.4,true,false,J_CENTRO_01
1729188000200,SAMU_01,SAMU,-23.310350,-51.162781,64.2,true,true,J_CENTRO_01
1729188000300,SAMU_01,SAMU,-23.310420,-51.162781,68.0,true,true,J_CENTRO_01
```

---

<div align="center">
  <img src="../assets/noxfort-logo.png" alt="Noxfort Systems Logo" width="45" /><br/>
  <b>Noxfort Systems</b> — <i>A State Of Art Company</i><br/>
  <i>Engenharia de Mobilidade Inteligente • EVP Simulator v1.0.0</i>
</div>
