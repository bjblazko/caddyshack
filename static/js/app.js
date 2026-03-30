/**
 * CaddyShack — main application logic.
 */
(function () {
    const uploadZone    = document.getElementById('upload-zone');
    const fileInput     = document.getElementById('file-input');
    const dashboard     = document.getElementById('dashboard');
    const loading       = document.getElementById('loading');
    const hostSelect    = document.getElementById('host-select');
    const fileSelector  = document.getElementById('file-selector');
    const fileSelect    = document.getElementById('file-select');

    // Reference to the currently loaded file (used for all re-analysis calls).
    let fileRef = null; // { type: 'uploaded', id } | { type: 'local', name }
    let uploadedFile = null; // { name, size } — display only
    let serverLogFiles = [];

    // Filter state — all conditions are ANDed on the backend.
    let currentHost      = '';
    let currentStatus    = 'all';
    let currentDateStart = null;
    let currentDateEnd   = null;
    let currentCountry   = null;
    let currentBrowser   = null;
    let currentOS        = null;
    let currentPage      = null;

    let dateDebounceTimer = null;

    // ── File loading ──────────────────────────────────────────────────────────

    uploadZone.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadZone.classList.add('dragover');
    });
    uploadZone.addEventListener('dragleave', () => uploadZone.classList.remove('dragover'));
    uploadZone.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadZone.classList.remove('dragover');
        if (e.dataTransfer.files.length > 0) startUpload(e.dataTransfer.files[0]);
    });
    fileInput.addEventListener('change', () => {
        if (fileInput.files.length > 0) startUpload(fileInput.files[0]);
    });

    async function startUpload(file) {
        uploadedFile = { name: file.name, size: file.size };
        const form = new FormData();
        form.append('logfile', file);

        resetFilters();
        loading.classList.remove('hidden');
        dashboard.classList.add('hidden');
        try {
            const resp = await fetch('/api/upload', { method: 'POST', body: form });
            if (!resp.ok) throw new Error(await resp.text() || resp.statusText);
            const result = await resp.json();
            fileRef = { type: 'uploaded', id: result.file_id };
            rebuildFileSelector();
            fileSelect.value = 'uploaded';
            applyResult(result);
        } catch (err) {
            alert('Error analyzing log file: ' + err.message);
        } finally {
            loading.classList.add('hidden');
        }
    }

    async function loadLocalFile(name) {
        resetFilters();
        fileRef = { type: 'local', name };
        await doFetch();
    }

    // ── File selector ─────────────────────────────────────────────────────────

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
        if (val === 'uploaded' && fileRef && fileRef.type === 'uploaded') {
            resetFilters();
            await doFetch();
        } else if (val.startsWith('server:')) {
            await loadLocalFile(val.slice(7));
        }
    });

    async function init() {
        try {
            const resp = await fetch('/api/logs');
            if (resp.ok) serverLogFiles = await resp.json();
        } catch (_) {}

        rebuildFileSelector();

        if (serverLogFiles.length > 0) {
            fileSelect.value = 'server:' + serverLogFiles[0].name;
            await loadLocalFile(serverLogFiles[0].name);
        }
    }

    init();

    // ── Fetch & render ────────────────────────────────────────────────────────

    function buildQuery() {
        const p = new URLSearchParams();
        if (fileRef.type === 'uploaded') p.set('file', fileRef.id);
        else p.set('name', fileRef.name);
        if (currentHost)      p.set('host',    currentHost);
        if (currentStatus !== 'all') p.set('status', currentStatus);
        if (currentDateStart) p.set('start',   currentDateStart);
        if (currentDateEnd)   p.set('end',     currentDateEnd);
        if (currentCountry)   p.set('country', currentCountry);
        if (currentBrowser)   p.set('browser', currentBrowser);
        if (currentOS)        p.set('os',      currentOS);
        if (currentPage)      p.set('page',    currentPage);
        return p;
    }

    async function doFetch() {
        if (!fileRef) return;
        loading.classList.remove('hidden');
        try {
            const resp = await fetch('/api/analyze?' + buildQuery());
            if (!resp.ok) throw new Error(await resp.text() || resp.statusText);
            applyResult(await resp.json());
        } catch (err) {
            alert('Error: ' + err.message);
        } finally {
            loading.classList.add('hidden');
        }
    }

    function scheduleFetch() {
        clearTimeout(dateDebounceTimer);
        dateDebounceTimer = setTimeout(doFetch, 400);
    }

    function applyResult(result) {
        populateHostDropdown(result.hosts || []);
        dashboard.classList.remove('hidden');
        renderDashboard(result.report);
    }

    // ── Filter state reset ────────────────────────────────────────────────────

    function resetFilters() {
        currentHost      = '';
        currentStatus    = 'all';
        currentDateStart = null;
        currentDateEnd   = null;
        currentCountry   = null;
        currentBrowser   = null;
        currentOS        = null;
        currentPage      = null;

        hostSelect.value = '';
        document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
        document.querySelector('.filter-btn[data-filter="all"]').classList.add('active');
        document.getElementById('date-start').value = '';
        document.getElementById('date-end').value   = '';
        document.getElementById('country-filter').value = '';
        document.getElementById('browser-filter').value = '';
        document.getElementById('os-filter').value      = '';
        document.getElementById('page-filter').value    = '';
    }

    // ── Render ────────────────────────────────────────────────────────────────

    function renderDashboard(data) {
        if (!data) { alert('Error: Invalid report data'); return; }

        document.getElementById('total-requests').textContent = (data.total_requests || 0).toLocaleString();
        document.getElementById('unique-ips').textContent     = (data.unique_ips     || 0).toLocaleString();
        document.getElementById('total-bytes').textContent    = formatBytes(data.total_bytes || 0);
        document.getElementById('avg-response').textContent   = ((data.avg_response_ms || 0).toFixed(1)) + ' ms';

        if (data.browsers && data.browsers.length > 0) {
            Charts.renderBarChart('browser-chart',
                data.browsers.map(b => b.name), data.browsers.map(b => b.count),
                data.total_requests, '#4a8c3f');
        }
        if (data.operating_systems && data.operating_systems.length > 0) {
            Charts.renderBarChart('os-chart',
                data.operating_systems.map(o => o.name), data.operating_systems.map(o => o.count),
                data.total_requests, '#2d5a27');
        }
        if (data.status_codes && data.status_codes.length > 0) {
            Charts.renderBarChart('status-chart',
                data.status_codes.map(s => s.name), data.status_codes.map(s => s.count),
                data.total_requests, '#8bc34a');
        }
        if (data.daily_traffic && data.daily_traffic.length > 0) {
            Charts.renderVerticalBarChart('daily-chart',
                data.daily_traffic.map(d => d.date), data.daily_traffic.map(d => d.count));
        }

        const total = data.total_requests || 0;
        renderTable('country-table', data.countries || [], c => [
            c.name, c.code, (c.count || 0).toLocaleString(),
            total > 0 ? ((c.count / total) * 100).toFixed(1) + '%' : '0%'
        ]);
        if (data.countries && data.countries.length > 0) {
            WorldMap.render('map-container', data.countries);
        }

        renderTable('pages-table', data.top_pages || [], p => [
            truncate(p.name, 60, p.name), (p.count || 0).toLocaleString()
        ]);
        renderTable('visitors-table', data.top_visitors || [], v => [
            v.ip, v.country_name + ' (' + v.country + ')', v.count.toLocaleString()
        ]);

        populateDimensionDropdowns(data);
        updateFilterHints();
    }

    // ── Host dropdown ─────────────────────────────────────────────────────────

    function populateHostDropdown(hosts) {
        hostSelect.innerHTML = '<option value="">All Sites</option>';
        for (const h of hosts) {
            const opt = document.createElement('option');
            opt.value = h;
            opt.textContent = h;
            hostSelect.appendChild(opt);
        }
        // Preserve selection if still valid; otherwise reset.
        if (currentHost && hosts.includes(currentHost)) {
            hostSelect.value = currentHost;
        } else {
            hostSelect.value = '';
            currentHost = '';
        }
    }

    hostSelect.addEventListener('change', () => {
        if (!fileRef) return;
        currentHost = hostSelect.value;
        doFetch();
    });

    // ── Status filter buttons ─────────────────────────────────────────────────

    document.querySelectorAll('.filter-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            if (!fileRef) return;
            currentStatus = btn.getAttribute('data-filter');
            document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            doFetch();
        });
    });

    // ── Date range inputs ─────────────────────────────────────────────────────

    document.getElementById('date-start').addEventListener('change', function () {
        currentDateStart = this.value || null;
        scheduleFetch();
    });
    document.getElementById('date-end').addEventListener('change', function () {
        currentDateEnd = this.value || null;
        scheduleFetch();
    });
    document.getElementById('date-clear').addEventListener('click', () => {
        if (!fileRef) return;
        currentDateStart = null;
        currentDateEnd   = null;
        document.getElementById('date-start').value = '';
        document.getElementById('date-end').value   = '';
        doFetch();
    });

    // ── Dimension dropdowns ───────────────────────────────────────────────────

    function populateDimensionDropdowns(report) {
        rebuildDimSelect('country-filter', (report.countries || []).map(c => c.name),
            () => { currentCountry = null; });
        rebuildDimSelect('browser-filter', (report.browsers || []).map(b => b.name),
            () => { currentBrowser = null; });
        rebuildDimSelect('os-filter', (report.operating_systems || []).map(o => o.name),
            () => { currentOS = null; });
        rebuildDimSelect('page-filter', (report.top_pages || []).map(p => p.name),
            () => { currentPage = null; }, 45);
    }

    function rebuildDimSelect(id, values, resetFn, truncateLen) {
        const sel = document.getElementById(id);
        const prev = sel.value;
        while (sel.options.length > 1) sel.remove(1);
        for (const v of values) {
            const opt = document.createElement('option');
            opt.value = v;
            opt.textContent = truncateLen && v.length > truncateLen ? v.slice(0, truncateLen) + '…' : v;
            opt.title = v;
            sel.appendChild(opt);
        }
        if (prev && values.includes(prev)) {
            sel.value = prev;
        } else {
            sel.value = '';
            resetFn();
        }
    }

    document.getElementById('country-filter').addEventListener('change', function () {
        if (!fileRef) return;
        currentCountry = this.value || null;
        doFetch();
    });
    document.getElementById('browser-filter').addEventListener('change', function () {
        if (!fileRef) return;
        currentBrowser = this.value || null;
        doFetch();
    });
    document.getElementById('os-filter').addEventListener('change', function () {
        if (!fileRef) return;
        currentOS = this.value || null;
        doFetch();
    });
    document.getElementById('page-filter').addEventListener('change', function () {
        if (!fileRef) return;
        currentPage = this.value || null;
        doFetch();
    });

    // ── Filter hints ──────────────────────────────────────────────────────────

    function updateFilterHints() {
        const date = (currentDateStart || currentDateEnd)
            ? (currentDateStart || '…') + ' – ' + (currentDateEnd || '…')
            : null;
        const pageLabel = currentPage
            ? (currentPage.length > 40 ? currentPage.slice(0, 40) + '…' : currentPage)
            : null;

        const statusLabel = currentStatus === 'success' ? 'Success (2xx)'
            : currentStatus === 'error' ? 'Errors (4xx–5xx)' : null;

        // All filters are applied on the backend — every panel reflects all active filters.
        const allTags = [currentHost || null, statusLabel, date, currentCountry,
                         currentBrowser, currentOS, pageLabel]
            .filter(Boolean);

        for (const id of ['hint-cards', 'hint-daily', 'hint-map', 'hint-countries',
                          'hint-browsers', 'hint-oses', 'hint-status', 'hint-pages',
                          'hint-visitors']) {
            setHint(id, allTags);
        }
    }

    function setHint(id, activeTags) {
        const el = document.getElementById(id);
        if (!el) return;
        el.innerHTML = '';
        if (activeTags.length === 0) {
            const tag = document.createElement('span');
            tag.className = 'filter-tag filter-tag--none';
            tag.textContent = 'all data';
            el.appendChild(tag);
        } else {
            for (const text of activeTags) {
                const tag = document.createElement('span');
                tag.className = 'filter-tag filter-tag--active';
                tag.textContent = text;
                el.appendChild(tag);
            }
        }
    }

    // ── Utilities ─────────────────────────────────────────────────────────────

    function renderTable(tableId, items, rowFn) {
        const tbody = document.getElementById(tableId).querySelector('tbody');
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

    function truncate(str, maxLen, fullValue) {
        if (str.length <= maxLen) return str;
        return { text: str.slice(0, maxLen) + '…', title: fullValue };
    }

    function formatBytes(bytes) {
        if (bytes < 1024)             return bytes + ' B';
        if (bytes < 1024 * 1024)      return (bytes / 1024).toFixed(1) + ' KiB';
        if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MiB';
        return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GiB';
    }
})();
