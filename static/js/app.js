/**
 * CaddyShack — main application logic.
 */
(function () {
    const $ = (id) => document.getElementById(id);
    const fileInput    = $('file-input');
    const dashboard    = $('dashboard');
    const emptyState   = $('empty-state');
    const hostSelect   = $('host-select');
    const fileSelector = $('file-selector');
    const fileSelect   = $('file-select');

    // Reference to the currently loaded file (used for all re-analysis calls).
    let fileRef = null; // { type: 'uploaded', id } | { type: 'local', name }
    let uploadedFile = null; // { name, size } — display only
    let serverLogFiles = [];
    let lastReport = null; // redrawn on resize and theme change

    // Filter state — all conditions are ANDed on the backend.
    let currentHost      = '';
    let currentStatus    = 'all';
    let currentDateStart = null;
    let currentDateEnd   = null;
    let currentCountry   = null;
    let currentBrowser   = null;
    let currentOS        = null;
    let currentPage      = null;
    let currentMethod    = null;
    let currentSearch    = '';
    // All exclusions are on by default: the first view shows page traffic
    // from people. Scripts have no exclusion, so probes always stay visible.
    let ignoreStatic     = true;
    let ignoreImages     = true;
    let ignoreMonitors   = true;
    let ignoreBots       = true;

    let dateDebounceTimer = null;

    // Tab & events state.
    let activeTab      = 'statistics';
    let eventsOffset   = 0;
    let eventsTotal    = 0;
    let eventsLoading  = false;
    let eventsDirty    = true;

    // ── Status line and messages ──────────────────────────────────────────────

    function setBusy(text) {
        $('status').textContent = text || '';
        dashboard.classList.toggle('is-loading', Boolean(text));
    }

    function showMessage(text) {
        $('message').textContent = text;
        $('message').hidden = false;
    }

    function clearMessage() { $('message').hidden = true; }

    function fileLabel() {
        if (!fileRef) return '';
        return fileRef.type === 'uploaded' ? uploadedFile.name : fileRef.name;
    }

    // ── File loading ──────────────────────────────────────────────────────────

    // The whole page accepts a dropped log file.
    document.addEventListener('dragover', (e) => {
        e.preventDefault();
        document.body.classList.add('dragging');
    });
    document.addEventListener('dragleave', (e) => {
        if (!e.relatedTarget) document.body.classList.remove('dragging');
    });
    document.addEventListener('drop', (e) => {
        e.preventDefault();
        document.body.classList.remove('dragging');
        if (e.dataTransfer.files.length > 0) startUpload(e.dataTransfer.files[0]);
    });
    fileInput.addEventListener('change', () => {
        if (fileInput.files.length > 0) startUpload(fileInput.files[0]);
        fileInput.value = '';
    });

    async function startUpload(file) {
        resetFilters();
        clearMessage();
        setBusy('Uploading and analyzing ' + file.name + '…');
        try {
            const resp = await fetch('/api/upload?' + buildFilterQuery(), { method: 'POST', body: form(file) });
            if (!resp.ok) throw new Error((await resp.text()).trim() || resp.statusText);
            const result = await resp.json();
            uploadedFile = { name: file.name, size: file.size };
            fileRef = { type: 'uploaded', id: result.file_id };
            rebuildFileSelector();
            fileSelect.value = 'uploaded';
            applyResult(result);
        } catch (err) {
            showMessage('Could not analyze ' + file.name + ': ' + err.message + '. Check that it is a Caddy JSON access log.');
        } finally {
            setBusy('');
        }
    }

    function form(file) {
        const f = new FormData();
        f.append('logfile', file);
        return f;
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

        fileSelector.hidden = fileSelect.options.length === 0;
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
        activateTab(location.hash === '#events' ? 'events' : 'statistics');
        try {
            const resp = await fetch('/api/logs');
            if (resp.ok) serverLogFiles = (await resp.json()) || [];
        } catch (_) {}

        rebuildFileSelector();

        if (serverLogFiles.length > 0) {
            fileSelect.value = 'server:' + serverLogFiles[0].name;
            await loadLocalFile(serverLogFiles[0].name);
        } else {
            emptyState.hidden = false;
        }
    }

    init();

    // ── About ─────────────────────────────────────────────────────────────────

    const about = $('about');
    $('about-open').addEventListener('click', openAbout);
    if (location.hash === '#about') openAbout(); // deep link, like the view tabs
    async function openAbout() {
        if (!about.open) about.showModal();
        if ($('about-version').textContent) return;
        try {
            const resp = await fetch('/api/health');
            const { version } = await resp.json();
            $('about-version').textContent = version === 'dev' ? 'Development build' : 'Version ' + version;
        } catch (_) {}
    }
    // A click on the backdrop (the dialog element itself, outside its box) closes it.
    about.addEventListener('click', (e) => { if (e.target === about) about.close(); });

    // ── Fetch & render ────────────────────────────────────────────────────────

    function buildQuery() {
        const p = buildFilterQuery();
        if (fileRef.type === 'uploaded') p.set('file', fileRef.id);
        else p.set('name', fileRef.name);
        return p;
    }

    function buildFilterQuery() {
        const p = new URLSearchParams();
        if (currentHost)      p.set('host',    currentHost);
        if (currentStatus !== 'all') p.set('status', currentStatus);
        if (currentDateStart) p.set('start',   currentDateStart);
        if (currentDateEnd)   p.set('end',     currentDateEnd);
        if (currentCountry)   p.set('country', currentCountry);
        if (currentBrowser)   p.set('browser', currentBrowser);
        if (currentOS)        p.set('os',      currentOS);
        if (currentPage)      p.set('page',    currentPage);
        if (currentMethod)    p.set('method',  currentMethod);
        if (currentSearch.trim()) p.set('search', currentSearch.trim());
        if (ignoreStatic)     p.set('ignore_static',   '1');
        if (ignoreImages)     p.set('ignore_images',   '1');
        if (ignoreMonitors)   p.set('ignore_monitors', '1');
        if (ignoreBots)       p.set('ignore_bots',     '1');
        return p;
    }

    async function doFetch() {
        if (!fileRef) return;
        setBusy('Analyzing ' + fileLabel() + '…');
        try {
            const resp = await fetch('/api/analyze?' + buildQuery());
            if (!resp.ok) throw new Error((await resp.text()).trim() || resp.statusText);
            applyResult(await resp.json());
        } catch (err) {
            showMessage('Could not analyze ' + fileLabel() + ': ' + err.message + '.');
        } finally {
            setBusy('');
        }
        // Mark events as stale; reload immediately if on events tab.
        eventsDirty = true;
        if (activeTab === 'events') loadEvents(true);
    }

    function scheduleFetch() {
        clearTimeout(dateDebounceTimer);
        dateDebounceTimer = setTimeout(doFetch, 400);
    }

    function applyResult(result) {
        clearMessage();
        populateHostDropdown(result.hosts || []);
        emptyState.hidden = true;
        dashboard.hidden = false;
        renderDashboard(result.report);
        if (activeTab === 'events' && eventsDirty) loadEvents(true);
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
        currentMethod    = null;
        currentSearch    = '';
        ignoreStatic     = true;
        ignoreImages     = true;
        ignoreMonitors   = true;
        ignoreBots       = true;

        hostSelect.value = '';
        setStatusButtons('all');
        for (const id of ['date-start', 'date-end', 'country-filter', 'browser-filter',
                          'os-filter', 'page-filter', 'method-filter', 'search-filter']) {
            $(id).value = '';
        }
        $('ignore-static').checked   = true;
        $('ignore-images').checked   = true;
        $('ignore-monitors').checked = true;
        $('ignore-bots').checked     = true;

        eventsOffset  = 0;
        eventsTotal   = 0;
        eventsLoading = false;
        eventsDirty   = true;
        $('events-tbody').innerHTML = '';
        setEventsStatus('');
    }

    // ── Render ────────────────────────────────────────────────────────────────

    function renderDashboard(data) {
        if (!data) { showMessage('The server returned no report.'); return; }
        lastReport = data;

        $('total-requests').textContent = (data.total_requests || 0).toLocaleString();
        $('unique-ips').textContent     = (data.unique_ips     || 0).toLocaleString();
        $('total-bytes').textContent    = formatBytes(data.total_bytes || 0);
        $('avg-response').textContent   = (data.avg_response_ms || 0).toFixed(1) + ' ms';

        // Canvas text needs the web fonts; draw once they are ready.
        document.fonts.ready.then(() => renderCharts(data));

        const total = data.total_requests || 0;
        renderTable('country-table', data.countries || [], c => [
            countryName(c.name, c.code), c.code === '??' ? '—' : c.code, (c.count || 0).toLocaleString(),
            total > 0 ? ((c.count / total) * 100).toFixed(1) + '%' : '0%'
        ]);
        // DB-IP Lite (CC BY 4.0) requires a link wherever its results are shown.
        const resolved = (data.countries || []).some(c => c.code && c.code !== '??');
        $('geoip-attribution').hidden = !resolved;
        $('geoip-missing').hidden = resolved || (data.countries || []).length === 0;

        renderTable('pages-table', data.top_pages || [], p => [
            truncate(p.name, 60), (p.count || 0).toLocaleString()
        ]);
        renderTable('visitors-table', data.top_visitors || [], v => [
            v.ip, v.country === '??' ? 'Unknown' : v.country_name + ' (' + v.country + ')', v.count.toLocaleString()
        ]);
        renderTable('referrers-table', data.top_referrers || [], r => [
            truncate(r.name, 60), (r.count || 0).toLocaleString()
        ], 'No external referrers for the current filters.');

        populateDimensionDropdowns(data);
        updateFilterSummary();
    }

    function renderCharts(data) {
        const total = data.total_requests || 0;
        const bars = (id, items, colorFor) => {
            clearEmptyChart(id);
            if (items.length === 0) return emptyChart(id);
            Charts.renderBarChart(id, items.map(x => x.name), items.map(x => x.count), total, colorFor);
        };
        bars('browser-chart', data.browsers || [], Charts.entityColor);
        bars('os-chart', data.operating_systems || [], Charts.entityColor);
        bars('status-chart', data.status_codes || [], Charts.statusColor);

        const days = data.daily_traffic || [];
        clearEmptyChart('daily-chart');
        if (days.length === 0) emptyChart('daily-chart');
        else Charts.renderVerticalBarChart('daily-chart', days.map(d => d.date), days.map(d => d.count));

        if (data.countries && data.countries.length > 0) WorldMap.render('map-container', data.countries);
        else $('map-container').innerHTML = '<p class="map-unavailable">No data for the current filters.</p>';
    }

    function emptyChart(id) {
        const canvas = $(id);
        canvas.getContext('2d').clearRect(0, 0, canvas.width, canvas.height);
        canvas.style.height = '0px';
        let note = canvas.nextElementSibling;
        if (!note || !note.classList.contains('chart-empty')) {
            note = document.createElement('p');
            note.className = 'chart-empty table-empty';
            canvas.after(note);
        }
        note.textContent = 'No data for the current filters.';
    }

    function clearEmptyChart(id) {
        const note = $(id).nextElementSibling;
        if (note && note.classList.contains('chart-empty')) note.remove();
    }

    // Redraw charts when the width or the color scheme changes.
    let resizeTimer = null;
    window.addEventListener('resize', () => {
        clearTimeout(resizeTimer);
        resizeTimer = setTimeout(() => { if (lastReport) renderCharts(lastReport); }, 150);
    });
    matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        if (lastReport) renderCharts(lastReport);
    });

    // ── Host dropdown ─────────────────────────────────────────────────────────

    function populateHostDropdown(hosts) {
        hostSelect.innerHTML = '<option value="">All sites</option>';
        for (const h of hosts) {
            const opt = document.createElement('option');
            opt.value = h;
            opt.textContent = h;
            hostSelect.appendChild(opt);
        }
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

    // ── Status segment group ──────────────────────────────────────────────────

    function setStatusButtons(value) {
        document.querySelectorAll('.status-btn').forEach(b =>
            b.setAttribute('aria-pressed', String(b.dataset.filter === value)));
    }

    document.querySelectorAll('.status-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            if (!fileRef) return;
            currentStatus = btn.dataset.filter;
            setStatusButtons(currentStatus);
            doFetch();
        });
    });

    // ── Date range inputs ─────────────────────────────────────────────────────

    $('date-start').addEventListener('change', function () {
        currentDateStart = this.value || null;
        scheduleFetch();
    });
    $('date-end').addEventListener('change', function () {
        currentDateEnd = this.value || null;
        scheduleFetch();
    });
    $('date-clear').addEventListener('click', () => {
        if (!fileRef) return;
        currentDateStart = null;
        currentDateEnd   = null;
        $('date-start').value = '';
        $('date-end').value   = '';
        doFetch();
    });
    $('filters').addEventListener('submit', (e) => e.preventDefault());

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
        rebuildDimSelect('method-filter', (report.methods || []).map(m => m.name),
            () => { currentMethod = null; });
    }

    function rebuildDimSelect(id, values, resetFn, truncateLen) {
        const sel = $(id);
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

    const onChange = (id, apply, debounce) => $(id).addEventListener(debounce ? 'input' : 'change', function () {
        if (!fileRef) return;
        apply(this);
        debounce ? scheduleFetch() : doFetch();
    });
    onChange('country-filter',  el => { currentCountry = el.value || null; });
    onChange('browser-filter',  el => { currentBrowser = el.value || null; });
    onChange('os-filter',       el => { currentOS      = el.value || null; });
    onChange('page-filter',     el => { currentPage    = el.value || null; });
    onChange('method-filter',   el => { currentMethod  = el.value || null; });
    onChange('search-filter',   el => { currentSearch  = el.value; }, true);
    onChange('ignore-static',   el => { ignoreStatic   = el.checked; });
    onChange('ignore-images',   el => { ignoreImages   = el.checked; });
    onChange('ignore-monitors', el => { ignoreMonitors = el.checked; });
    onChange('ignore-bots',     el => { ignoreBots     = el.checked; });

    // ── Filter summary (one line for all views) ───────────────────────────────

    function updateFilterSummary() {
        const date = (currentDateStart || currentDateEnd)
            ? (currentDateStart || '…') + ' – ' + (currentDateEnd || '…') : null;
        const statusLabel = currentStatus === 'success' ? 'success (2xx)'
            : currentStatus === 'error' ? 'errors (4xx–5xx)' : null;
        const page = currentPage && currentPage.length > 40 ? currentPage.slice(0, 40) + '…' : currentPage;

        const only = [currentHost || null, statusLabel, date, currentCountry, currentBrowser,
                      currentOS, page, currentMethod,
                      currentSearch.trim() ? 'matching “' + currentSearch.trim() + '”' : null].filter(Boolean);
        const without = [ignoreStatic && 'static files', ignoreImages && 'images',
                         ignoreMonitors && 'monitors', ignoreBots && 'bots'].filter(Boolean);

        let text = only.length ? 'Showing ' + only.join(' · ') : 'Showing all requests';
        if (without.length) text += ', without ' + listJoin(without);
        $('filter-summary').textContent = text + '.';
    }

    // ── Views (links, so back button and deep links work) ─────────────────────

    function activateTab(tab) {
        activeTab = tab;
        document.querySelectorAll('.tabs a').forEach(a => {
            if (a.dataset.tab === tab) a.setAttribute('aria-current', 'page');
            else a.removeAttribute('aria-current');
        });
        $('tab-statistics').hidden = tab !== 'statistics';
        $('tab-events').hidden = tab !== 'events';
        if (tab === 'events' && eventsDirty) loadEvents(true);
        if (tab === 'statistics' && lastReport) renderCharts(lastReport); // canvas width was 0 while hidden
    }

    window.addEventListener('hashchange', () => {
        activateTab(location.hash === '#events' ? 'events' : 'statistics');
    });

    // ── Single events ─────────────────────────────────────────────────────────

    async function loadEvents(reset) {
        if (!fileRef || eventsLoading) return;
        if (reset) {
            eventsOffset = 0;
            eventsTotal  = 0;
            $('events-tbody').innerHTML = '';
        }
        eventsLoading = true;
        setEventsStatus('Loading events…');

        const p = buildQuery();
        p.set('offset', eventsOffset);
        p.set('limit', 100);

        try {
            const resp = await fetch('/api/events?' + p);
            if (!resp.ok) throw new Error((await resp.text()).trim() || resp.statusText);
            const result = await resp.json();
            eventsTotal = result.total;
            appendEventRows(result.events || []);
            eventsOffset += (result.events || []).length;
            eventsDirty = false;
            updateEventsStatus();
        } catch (err) {
            setEventsStatus('Could not load events: ' + err.message + '.');
        } finally {
            eventsLoading = false;
        }
    }

    function appendEventRows(events) {
        const tbody = $('events-tbody');
        for (const ev of events) {
            const tr = document.createElement('tr');
            const isError = ev.status >= 400;
            if (isError) tr.className = 'is-error';

            const cells = [
                { text: ev.ts.replace('T', ' ').replace(/\.\d+Z$/, ''), cls: 'mono', title: ev.ts },
                { text: ev.method },
                { text: ev.host },
                { text: ev.uri, cls: 'wrap', title: ev.uri },
                { text: String(ev.status), cls: 'num' + (isError ? ' status-error' : '') },
                { text: formatBytes(ev.size), cls: 'num' },
                { text: ev.duration_ms.toFixed(1) + ' ms', cls: 'num' },
                { text: ev.ip, cls: 'mono' },
                { text: countryName(ev.country_name, ev.country) },
                { text: ev.browser },
                { text: ev.os },
                { text: ev.referer || '', cls: 'wrap', title: ev.referer || '' },
            ];

            for (const c of cells) {
                const td = document.createElement('td');
                td.textContent = c.text;
                if (c.cls) td.className = c.cls;
                if (c.title) td.title = c.title;
                tr.appendChild(td);
            }
            tbody.appendChild(tr);
        }
    }

    function updateEventsStatus() {
        if (eventsTotal === 0) {
            setEventsStatus('No events match the current filters.');
        } else if (eventsOffset >= eventsTotal) {
            setEventsStatus('All ' + eventsTotal.toLocaleString() + ' events loaded.');
        } else {
            setEventsStatus('Showing ' + eventsOffset.toLocaleString() + ' of ' + eventsTotal.toLocaleString() + ' events. Scroll to load more.');
        }
    }

    function setEventsStatus(msg) { $('events-status').textContent = msg; }

    // Lazy-load more events when the sentinel scrolls into view.
    new IntersectionObserver(entries => {
        if (entries[0].isIntersecting && eventsOffset < eventsTotal && !eventsLoading) {
            loadEvents(false);
        }
    }, { rootMargin: '200px' }).observe($('events-sentinel'));

    // ── Utilities ─────────────────────────────────────────────────────────────

    // Columns whose header has class "num" are right-aligned, tabular numbers.
    function renderTable(tableId, items, rowFn, emptyText = 'No data for the current filters.') {
        const table = $(tableId);
        const numeric = [...table.querySelectorAll('thead th')].map(th => th.classList.contains('num'));
        const tbody = table.querySelector('tbody');
        tbody.innerHTML = '';
        if (items.length === 0) {
            const td = document.createElement('td');
            td.className = 'table-empty';
            td.colSpan = numeric.length;
            td.textContent = emptyText;
            tbody.appendChild(document.createElement('tr')).appendChild(td);
            return;
        }
        for (const item of items) {
            const tr = document.createElement('tr');
            rowFn(item).forEach((val, i) => {
                const td = document.createElement('td');
                if (val && typeof val === 'object' && val.text !== undefined) {
                    td.textContent = val.text;
                    td.title = val.title;
                } else {
                    td.textContent = val;
                }
                if (numeric[i]) td.className = 'num';
                tr.appendChild(td);
            });
            tbody.appendChild(tr);
        }
    }

    // "??" is the backend's code for "no country known".
    function countryName(name, code) {
        return code === '??' || !code ? 'Unknown' : (name || code);
    }

    function listJoin(items) {
        return items.length < 2 ? items.join('') : items.slice(0, -1).join(', ') + ' and ' + items[items.length - 1];
    }

    function truncate(str, maxLen) {
        if (str.length <= maxLen) return str;
        return { text: str.slice(0, maxLen) + '…', title: str };
    }

    function formatBytes(bytes) {
        if (bytes < 1024)               return bytes + ' B';
        if (bytes < 1024 * 1024)        return (bytes / 1024).toFixed(1) + ' KiB';
        if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MiB';
        return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GiB';
    }
})();
