/**
 * CaddyShack — main application logic.
 */
(function () {
    const uploadZone = document.getElementById('upload-zone');
    const fileInput = document.getElementById('file-input');
    const dashboard = document.getElementById('dashboard');
    const loading = document.getElementById('loading');
    const hostSelect = document.getElementById('host-select');
    const fileSelector = document.getElementById('file-selector');
    const fileSelect = document.getElementById('file-select');
    let fullReport = null;
    let currentHost = 'all';
    let currentFilter = 'all';
    let uploadedFile = null;  // { name, size, formData }

    // Drag-and-drop
    uploadZone.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadZone.classList.add('dragover');
    });
    uploadZone.addEventListener('dragleave', () => {
        uploadZone.classList.remove('dragover');
    });
    uploadZone.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadZone.classList.remove('dragover');
        if (e.dataTransfer.files.length > 0) {
            uploadFile(e.dataTransfer.files[0]);
        }
    });

    // File input
    fileInput.addEventListener('change', () => {
        if (fileInput.files.length > 0) {
            uploadFile(fileInput.files[0]);
        }
    });

    async function uploadFile(file) {
        const form = new FormData();
        form.append('logfile', file);
        uploadedFile = { name: file.name, size: file.size, form };
        rebuildFileSelector();
        fileSelect.value = 'uploaded';
        await loadReport(() => fetch('/api/upload', { method: 'POST', body: form }), file.name);
    }

    async function loadLocalFile(name) {
        await loadReport(() => fetch('/api/analyze-local?name=' + encodeURIComponent(name)), name);
    }

    async function loadReport(fetchFn, label) {
        loading.classList.remove('hidden');
        dashboard.classList.add('hidden');
        try {
            const resp = await fetchFn();
            if (!resp.ok) {
                const text = await resp.text();
                throw new Error(text || resp.statusText);
            }
            fullReport = await resp.json();
            currentHost = 'all';
            currentFilter = 'all';
            populateHostDropdown(fullReport.hosts || []);
            document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
            document.querySelector('.filter-btn[data-filter="all"]').classList.add('active');
            renderDashboard(fullReport.by_host['all']['all']);
        } catch (err) {
            alert('Error analyzing log file: ' + err.message);
        } finally {
            loading.classList.add('hidden');
        }
    }

    // Build/rebuild the file selector dropdown from known server files + any uploaded file
    let serverLogFiles = [];

    function rebuildFileSelector() {
        fileSelect.innerHTML = '';

        if (serverLogFiles.length > 0) {
            const grp = document.createElement('optgroup');
            grp.label = 'Server logs';
            for (const f of serverLogFiles) {
                const opt = document.createElement('option');
                opt.value = 'server:' + f.name;
                opt.textContent = f.name + ' (' + formatBytes(f.size) + ')';
                grp.appendChild(opt);
            }
            fileSelect.appendChild(grp);
        }

        if (uploadedFile) {
            const grp = document.createElement('optgroup');
            grp.label = 'Uploaded';
            const opt = document.createElement('option');
            opt.value = 'uploaded';
            opt.textContent = uploadedFile.name + ' (' + formatBytes(uploadedFile.size) + ')';
            grp.appendChild(opt);
            fileSelect.appendChild(grp);
        }

        fileSelector.classList.toggle('hidden', fileSelect.options.length === 0);
    }

    fileSelect.addEventListener('change', async () => {
        const val = fileSelect.value;
        if (val === 'uploaded' && uploadedFile) {
            await loadReport(() => fetch('/api/upload', { method: 'POST', body: uploadedFile.form }), uploadedFile.name);
        } else if (val.startsWith('server:')) {
            await loadLocalFile(val.slice(7));
        }
    });

    // On startup, discover server log files and auto-load the most recent
    async function init() {
        try {
            const resp = await fetch('/api/logs');
            if (resp.ok) {
                serverLogFiles = await resp.json();
            }
        } catch (_) {}

        rebuildFileSelector();

        if (serverLogFiles.length > 0) {
            fileSelect.value = 'server:' + serverLogFiles[0].name;
            await loadLocalFile(serverLogFiles[0].name);
        }
    }

    init();

    function renderDashboard(data) {
        if (!data) {
            alert('Error: Invalid report data');
            return;
        }
        dashboard.classList.remove('hidden');

        // Summary cards
        document.getElementById('total-requests').textContent = (data.total_requests || 0).toLocaleString();
        document.getElementById('unique-ips').textContent = (data.unique_ips || 0).toLocaleString();
        document.getElementById('total-bytes').textContent = formatBytes(data.total_bytes || 0);
        document.getElementById('avg-response').textContent = ((data.avg_response_ms || 0).toFixed(1)) + ' ms';

        // Browser chart
        if (data.browsers && data.browsers.length > 0) {
            Charts.renderBarChart(
                'browser-chart',
                data.browsers.map(b => b.name),
                data.browsers.map(b => b.count),
                data.total_requests,
                '#4a8c3f'
            );
        }

        // OS chart
        if (data.operating_systems && data.operating_systems.length > 0) {
            Charts.renderBarChart(
                'os-chart',
                data.operating_systems.map(o => o.name),
                data.operating_systems.map(o => o.count),
                data.total_requests,
                '#2d5a27'
            );
        }

        // Status codes
        if (data.status_codes && data.status_codes.length > 0) {
            Charts.renderBarChart(
                'status-chart',
                data.status_codes.map(s => s.name),
                data.status_codes.map(s => s.count),
                data.total_requests,
                '#8bc34a'
            );
        }

        // Daily traffic
        if (data.daily_traffic && data.daily_traffic.length > 0) {
            Charts.renderVerticalBarChart(
                'daily-chart',
                data.daily_traffic.map(d => d.date),
                data.daily_traffic.map(d => d.count)
            );
        }

        // Country table
        const totalRequests = data.total_requests || 0;
        renderTable('country-table', data.countries || [], (c) => [
            c.name, c.code, (c.count || 0).toLocaleString(),
            totalRequests > 0 ? ((c.count / totalRequests) * 100).toFixed(1) + '%' : '0%'
        ]);

        // Top pages table
        renderTable('pages-table', data.top_pages || [], (p) => [
            truncate(p.name, 60, p.name),
            (p.count || 0).toLocaleString()
        ]);

        // Top visitors table
        renderTable('visitors-table', data.top_visitors || [], (v) => [
            v.ip, v.country_name + ' (' + v.country + ')', v.count.toLocaleString()
        ]);

        // World map
        if (data.countries && data.countries.length > 0) {
            WorldMap.render('map-container', data.countries);
        }
    }

    function truncate(str, maxLen, fullValue) {
        if (str.length <= maxLen) return str;
        return { text: str.slice(0, maxLen) + '…', title: fullValue };
    }

    function renderTable(tableId, items, rowFn) {
        const table = document.getElementById(tableId);
        const tbody = table.querySelector('tbody');
        tbody.innerHTML = '';
        for (const item of items) {
            const tr = document.createElement('tr');
            for (const val of rowFn(item)) {
                const td = document.createElement('td');
                if (val && typeof val === 'object' && val.text !== undefined) {
                    td.textContent = val.text;
                    td.title = val.title;
                } else {
                    td.textContent = val;
                }
                tr.appendChild(td);
            }
            tbody.appendChild(tr);
        }
    }

    function formatBytes(bytes) {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KiB';
        if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MiB';
        return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GiB';
    }

    function populateHostDropdown(hosts) {
        hostSelect.innerHTML = '<option value="all">All Sites</option>';
        for (const h of hosts) {
            const opt = document.createElement('option');
            opt.value = h;
            opt.textContent = h;
            hostSelect.appendChild(opt);
        }
        hostSelect.value = 'all';
    }

    function currentReport() {
        return fullReport && fullReport.by_host[currentHost]
            ? fullReport.by_host[currentHost][currentFilter]
            : null;
    }

    // Host dropdown
    hostSelect.addEventListener('change', () => {
        if (!fullReport) return;
        currentHost = hostSelect.value;
        renderDashboard(currentReport());
    });

    // Filter buttons
    document.querySelectorAll('.filter-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            if (!fullReport) return;
            currentFilter = btn.getAttribute('data-filter');
            document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            renderDashboard(currentReport());
        });
    });
})();
