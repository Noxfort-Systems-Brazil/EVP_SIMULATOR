// EVP Simulator is an open-source emergency vehicle preemption and fleet simulation platform for urban traffic networks.
// Copyright (C) 2026 Noxfort Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// File: app.js
// Author: Gabriel Moraes
// Date: October 2026

let isRunning = false;
let isExporting = true;
let currentFilter = 'all';
let isAutoScroll = true;
let localLogs = [];
let clearedLogsCount = 0;
let currentAlertCount = 0;

// Safe Wails Bridge Call Wrapper
async function callBackend(funcName, ...args) {
    if (window.go && window.go.wails && window.go.wails.App && window.go.wails.App[funcName]) {
        try {
            return await window.go.wails.App[funcName](...args);
        } catch (err) {
            console.error(`Error calling ${funcName}:`, err);
            return null;
        }
    }
    return null;
}

// UI Elements
const simStatusPill = document.getElementById('sim-status-pill');
const simStatusText = document.getElementById('sim-status-text');
const terminal = document.getElementById('log-terminal');
const logCounter = document.getElementById('log-counter');
const checkAutoScroll = document.getElementById('check-autoscroll');
const btnClearLogs = document.getElementById('btn-clear-logs');

const statNetwork = document.getElementById('stat-network');
const statLights = document.getElementById('stat-lights');
const statVehicles = document.getElementById('stat-vehicles');

// Event Handlers
document.getElementById('btn-load-map').addEventListener('click', async () => {
    await callBackend('OpenMapDialog');
});

document.getElementById('btn-export-vehicles').addEventListener('click', async () => {
    await callBackend('ExportVehiclesCSVDialog');
});

document.getElementById('btn-export-controllers').addEventListener('click', async () => {
    await callBackend('ExportControllersCSVDialog');
});

// Endpoint Configuration Modal Handlers
const modalEndpoint = document.getElementById('modal-endpoint');
const inputTargetUrl = document.getElementById('input-target-url');
const btnConfigTarget = document.getElementById('btn-config-target');
const btnEditDest = document.getElementById('btn-edit-dest');
const btnModalClose = document.getElementById('btn-modal-close');
const btnModalCancel = document.getElementById('btn-modal-cancel');
const btnModalSave = document.getElementById('btn-modal-save');
const btnPasteClipboard = document.getElementById('btn-paste-clipboard');
const btnPresetLocal = document.getElementById('btn-preset-local');

function openEndpointModal() {
    if (!modalEndpoint) return;
    const footerTargetUrl = document.getElementById('footer-target-url');
    if (inputTargetUrl) {
        inputTargetUrl.value = (footerTargetUrl ? footerTargetUrl.textContent.trim() : '') || 'http://localhost:2201/amb/v1';
    }
    modalEndpoint.style.display = 'flex';
    setTimeout(() => {
        if (inputTargetUrl) inputTargetUrl.focus();
    }, 50);
}

function closeEndpointModal() {
    if (modalEndpoint) modalEndpoint.style.display = 'none';
}

if (btnConfigTarget) btnConfigTarget.addEventListener('click', openEndpointModal);
if (btnEditDest) btnEditDest.addEventListener('click', openEndpointModal);
if (btnModalClose) btnModalClose.addEventListener('click', closeEndpointModal);
if (btnModalCancel) btnModalCancel.addEventListener('click', closeEndpointModal);

if (modalEndpoint) {
    modalEndpoint.addEventListener('click', (e) => {
        if (e.target === modalEndpoint) closeEndpointModal();
    });
}

if (btnPresetLocal && inputTargetUrl) {
    btnPresetLocal.addEventListener('click', () => {
        inputTargetUrl.value = 'http://localhost:2201/amb/v1';
        inputTargetUrl.focus();
    });
}

if (btnPasteClipboard && inputTargetUrl) {
    btnPasteClipboard.addEventListener('click', async () => {
        try {
            const text = await navigator.clipboard.readText();
            if (text && text.trim()) {
                inputTargetUrl.value = text.trim();
            }
        } catch (err) {
            console.warn('Clipboard read failed:', err);
        }
    });
}

if (btnModalSave && inputTargetUrl) {
    btnModalSave.addEventListener('click', async () => {
        const url = inputTargetUrl.value.trim();
        if (url) {
            const savedUrl = await callBackend('SetTargetURL', url);
            if (savedUrl) {
                const footerTargetUrl = document.getElementById('footer-target-url');
                if (footerTargetUrl) footerTargetUrl.textContent = savedUrl;
            }
            closeEndpointModal();
            await pollState();
        }
    });
}

if (inputTargetUrl) {
    inputTargetUrl.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            if (btnModalSave) btnModalSave.click();
        } else if (e.key === 'Escape') {
            closeEndpointModal();
        }
    });
}

const btnToggleFleetAlert = document.getElementById('btn-toggle-fleet-alert');
if (btnToggleFleetAlert) {
    btnToggleFleetAlert.addEventListener('click', async () => {
        // Se houver qualquer viatura em alerta, desativa todas. Se todas normais, ativa todas.
        const shouldAlert = currentAlertCount === 0;
        await callBackend('ToggleAllVehiclesMode', shouldAlert);
        await pollState();
    });
}

btnClearLogs.addEventListener('click', () => {
    clearedLogsCount = localLogs.length;
    renderLogs(localLogs);
});

checkAutoScroll.addEventListener('change', (e) => {
    isAutoScroll = e.target.checked;
});

// Filter Tabs Handling
document.querySelectorAll('.filter-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        currentFilter = btn.dataset.filter;
        renderLogs(localLogs);
    });
});

// UI Sync Helpers
function updateStatusUI(mapLoaded) {
    if (isRunning) {
        simStatusPill.className = 'status-pill running';
        simStatusText.textContent = 'Executando (10 Hz)';
    } else if (mapLoaded) {
        simStatusPill.className = 'status-pill paused';
        simStatusText.textContent = 'Pausado';
    } else {
        simStatusPill.className = 'status-pill paused';
        simStatusText.textContent = 'Aguardando Mapa';
    }
}

// Parse and Categorize Log Entry
function parseLog(rawText) {
    let time = '';
    let msg = rawText;

    // Extract [HH:MM:SS] if present
    const timeMatch = rawText.match(/^\[(\d{2}:\d{2}:\d{2})\]\s*(.*)/);
    if (timeMatch) {
        time = `[${timeMatch[1]}]`;
        msg = timeMatch[2];
    }

    let category = 'system';
    let tag = 'SISTEMA';
    let entryClass = '';

    const lower = msg.toLowerCase();

    if (msg.includes('🚨') || lower.includes('alerta') || lower.includes('sirene')) {
        category = 'alert';
        tag = 'ALERTA';
        entryClass = 'log-alert';
    } else if (lower.includes('telemetria') || lower.includes('carina_core') || lower.includes('export')) {
        category = 'telemetry';
        tag = 'TELEMETRIA';
        entryClass = 'log-telemetry';
    } else if (lower.includes('semáforo') || lower.includes('controlador') || lower.includes('sinal')) {
        category = 'traffic';
        tag = 'TRÁFEGO';
        entryClass = 'log-traffic';
    } else if (lower.includes('mapa') || lower.includes('malha') || lower.includes('viatura') || lower.includes('simulação')) {
        category = 'system';
        tag = 'SISTEMA';
        entryClass = 'log-system';
    } else {
        category = 'system';
        tag = 'INFO';
        entryClass = '';
    }

    return { time, msg, category, tag, entryClass };
}

// Render Log Entries
function renderLogs(logs) {
    if (!logs) return;

    // Discard entries before clear
    const visibleLogs = logs.slice(clearedLogsCount);

    // Apply Filter
    const filtered = visibleLogs.filter(raw => {
        if (currentFilter === 'all') return true;
        const parsed = parseLog(raw);
        if (currentFilter === 'alert') return parsed.category === 'alert';
        if (currentFilter === 'telemetry') return parsed.category === 'telemetry';
        if (currentFilter === 'system') return parsed.category === 'system' || parsed.category === 'traffic';
        return true;
    });

    logCounter.textContent = `${filtered.length} evento${filtered.length === 1 ? '' : 's'}`;

    if (filtered.length === 0) {
        terminal.innerHTML = '<div class="log-empty">Nenhum evento registrado para este filtro.</div>';
        return;
    }

    terminal.innerHTML = filtered.map(raw => {
        const item = parseLog(raw);
        return `
            <div class="log-entry ${item.entryClass}">
                <span class="log-time">${item.time}</span>
                <span class="log-tag tag-${item.category}">${item.tag}</span>
                <span class="log-msg">${escapeHtml(item.msg)}</span>
            </div>
        `;
    }).join('');

    if (isAutoScroll) {
        terminal.scrollTop = terminal.scrollHeight;
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Background Periodic State Polling (5 Hz)
async function pollState() {
    const state = await callBackend('GetSimulationState');
    if (!state) return;

    isRunning = state.is_running;
    isExporting = state.is_exporting;

    updateStatusUI(state.map_loaded);

    // Update bottom stats
    if (state.map_loaded) {
        statNetwork.textContent = `Malha: ${state.total_edges} vias`;
        statLights.textContent = `Semáforos: ${state.total_lights} (${state.mapped_controllers} com IP/UTMC)`;
    } else {
        statNetwork.textContent = 'Malha: aguardando importação...';
        statLights.textContent = 'Semáforos: 0';
    }

    const vehicles = state.vehicles || [];
    const totalVehicles = vehicles.length;
    const alertCount = vehicles.filter(v => v.mode === 'ALERT').length;
    currentAlertCount = alertCount;

    // Count by emergency service corporation
    const samuCount = vehicles.filter(v => v.corporation === 'SAMU').length;
    const bombeiroCount = vehicles.filter(v => v.corporation === 'BOMBEIROS').length;
    const policiaCount = vehicles.filter(v => v.corporation === 'POLICIA').length;

    // Update highlight badge pill at top
    const vehicleTotalNum = document.getElementById('vehicle-total-num');
    const vehicleCorpBreakdown = document.getElementById('vehicle-corp-breakdown');
    if (vehicleTotalNum) {
        vehicleTotalNum.textContent = `${totalVehicles}`;
    }
    if (vehicleCorpBreakdown) {
        const parts = [];
        if (samuCount > 0) parts.push(`${samuCount} SAMU`);
        if (bombeiroCount > 0) parts.push(`${bombeiroCount} Bombeiros`);
        if (policiaCount > 0) parts.push(`${policiaCount} Polícia`);
        if (alertCount > 0) parts.push(`🚨 ${alertCount} em alerta`);
        vehicleCorpBreakdown.textContent = parts.length > 0 ? `(${parts.join(' • ')})` : '';
    }

    // Update badge chip in event console header
    const consoleVehicleBadge = document.getElementById('console-vehicle-badge');
    if (consoleVehicleBadge) {
        consoleVehicleBadge.textContent = `🚑 ${totalVehicles} viatura${totalVehicles === 1 ? '' : 's'} conectada${totalVehicles === 1 ? '' : 's'}${alertCount > 0 ? ` (${alertCount} em alerta)` : ''}`;
    }

    // Update fleet-wide alert toggle button
    const btnToggleFleetAlert = document.getElementById('btn-toggle-fleet-alert');
    if (btnToggleFleetAlert) {
        if (totalVehicles > 0) {
            btnToggleFleetAlert.style.display = 'inline-flex';
            const iconEl = document.getElementById('btn-fleet-alert-icon');
            const labelEl = document.getElementById('btn-fleet-alert-label');
            if (alertCount > 0) {
                if (iconEl) iconEl.textContent = '⚪';
                if (labelEl) labelEl.textContent = 'Desativar Todos Alertas';
                btnToggleFleetAlert.title = 'Clique para desativar o modo alerta de toda a frota (retornar a patrulha normal)';
            } else {
                if (iconEl) iconEl.textContent = '🚨';
                if (labelEl) labelEl.textContent = 'Ativar Alerta Geral';
                btnToggleFleetAlert.title = 'Clique para ativar o modo alerta de emergência para toda a frota';
            }
        } else {
            btnToggleFleetAlert.style.display = 'none';
        }
    }

    // Render or update individual vehicle quick-toggle chips
    const fleetContainer = document.getElementById('fleet-chips-container');
    if (fleetContainer) {
        if (vehicles.length === 0) {
            fleetContainer.innerHTML = '';
        } else {
            const existingButtons = fleetContainer.querySelectorAll('.fleet-chip');
            const needsRebuild = existingButtons.length !== vehicles.length ||
                Array.from(existingButtons).some((btn, idx) => btn.dataset.id !== vehicles[idx].id);

            if (needsRebuild) {
                fleetContainer.innerHTML = vehicles.map(v => {
                    const isAlert = v.mode === 'ALERT';
                    const chipClass = isAlert ? 'chip-alert' : 'chip-normal';
                    const icon = isAlert ? '🚨' : '⚪';
                    const modeLabel = isAlert ? 'ALERTA' : 'NORMAL';
                    const speed = Math.round(v.speed_kmh || 0);
                    return `
                        <button class="fleet-chip ${chipClass}" data-id="${v.id}" title="Clique para alternar o modo de ${v.id} (Alerta / Normal)">
                            <span class="chip-icon">${icon}</span>
                            <span class="chip-id">${v.id}</span>
                            <span class="chip-mode">${modeLabel}</span>
                            <span class="chip-speed">${speed} km/h</span>
                        </button>
                    `;
                }).join('');

                fleetContainer.querySelectorAll('.fleet-chip').forEach(btn => {
                    btn.addEventListener('click', async (e) => {
                        e.stopPropagation();
                        const vehId = btn.dataset.id;
                        await callBackend('ToggleVehicleMode', vehId);
                        await pollState();
                    });
                });
            } else {
                vehicles.forEach(v => {
                    const btn = fleetContainer.querySelector(`.fleet-chip[data-id="${v.id}"]`);
                    if (!btn) return;
                    const isAlert = v.mode === 'ALERT';
                    const chipClass = isAlert ? 'chip-alert' : 'chip-normal';
                    btn.className = `fleet-chip ${chipClass}`;
                    const iconEl = btn.querySelector('.chip-icon');
                    if (iconEl) iconEl.textContent = isAlert ? '🚨' : '⚪';
                    const modeEl = btn.querySelector('.chip-mode');
                    if (modeEl) modeEl.textContent = isAlert ? 'ALERTA' : 'NORMAL';
                    const speedEl = btn.querySelector('.chip-speed');
                    if (speedEl) speedEl.textContent = `${Math.round(v.speed_kmh || 0)} km/h`;
                });
            }
        }
    }

    // Update footer metrics
    statVehicles.textContent = `Viaturas: ${totalVehicles} (SAMU: ${samuCount} | Bombeiros: ${bombeiroCount} | Polícia: ${policiaCount})`;

    const footerTargetUrl = document.getElementById('footer-target-url');
    if (footerTargetUrl && state.telemetry_url) {
        footerTargetUrl.textContent = state.telemetry_url;
    }

    // Update logs
    if (state.recent_logs && state.recent_logs.length > 0) {
        localLogs = state.recent_logs;
        renderLogs(localLogs);
    }
}

setInterval(pollState, 200);
pollState();
