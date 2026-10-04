/**
 * Charts module — Canvas 2D bar charts. Colors come from the CSS tokens at
 * draw time, so light and dark mode need no code; redraw on theme change.
 */

const Charts = (() => {
    // Fixed color per entity, so a browser keeps its color when filters change
    // the ranking. Non-browser clients and "Other" stay neutral on purpose.
    const ENTITY_SLOTS = {
        Chrome: 1, Firefox: 2, Safari: 3, Brave: 4, Opera: 5, Vivaldi: 6, Edge: 7,
        Windows: 1, macOS: 3, Linux: 4, ChromeOS: 5, Android: 6, iOS: 7,
    };

    function token(name) {
        return getComputedStyle(document.documentElement).getPropertyValue('--' + name).trim();
    }

    function entityColor(label) {
        const slot = ENTITY_SLOTS[label];
        return slot ? token('series-' + slot) : token('fg-3');
    }

    // 2xx confirmed, 4xx/5xx need attention, 3xx neutral; the code is the label.
    function statusColor(code) {
        if (code.startsWith('2')) return token('confirm');
        if (code.startsWith('4') || code.startsWith('5')) return token('warning');
        return token('fg-3');
    }

    function font(size, family, weight) {
        return (weight || 400) + ' ' + size + 'px ' + token(family || 'font-sans');
    }

    // Size the canvas to its CSS width at device resolution; returns the context.
    function setup(canvas, height) {
        const dpr = window.devicePixelRatio || 1;
        const width = canvas.clientWidth || canvas.parentElement.clientWidth;
        canvas.width = Math.round(width * dpr);
        canvas.height = Math.round(height * dpr);
        canvas.style.height = height + 'px';
        const ctx = canvas.getContext('2d');
        ctx.scale(dpr, dpr);
        return { ctx, width };
    }

    function roundedBar(ctx, x, y, w, h, r) {
        if (w <= 0) return;
        r = Math.min(r, w / 2, h / 2);
        ctx.beginPath();
        ctx.moveTo(x, y);
        ctx.lineTo(x + w - r, y);
        ctx.arcTo(x + w, y, x + w, y + r, r);
        ctx.lineTo(x + w, y + h - r);
        ctx.arcTo(x + w, y + h, x + w - r, y + h, r);
        ctx.lineTo(x, y + h);
        ctx.closePath();
        ctx.fill();
    }

    /**
     * Ranked horizontal bars: label · bar · count and share.
     * @param {string} canvasId
     * @param {string[]} labels
     * @param {number[]} values
     * @param {number} total - for the share column (0 to omit)
     * @param {(label: string) => string} colorFor - bar color per label
     */
    function renderBarChart(canvasId, labels, values, total, colorFor) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) return;
        const row = 28, bar = 12;
        const { ctx, width } = setup(canvas, Math.max(row, labels.length * row));

        ctx.font = font(13);
        const labelW = Math.min(width * 0.4,
            Math.max(...labels.map(l => ctx.measureText(l).width), 0) + 12);
        ctx.font = font(13, 'font-mono');
        const valueTexts = values.map(v => v.toLocaleString() +
            (total > 0 ? '  ' + ((v / total) * 100).toFixed(1).padStart(4) + '%' : ''));
        const valueW = Math.max(...valueTexts.map(t => ctx.measureText(t).width), 0) + 12;
        const trackW = Math.max(0, width - labelW - valueW);
        const maxVal = Math.max(...values, 1);

        ctx.textBaseline = 'middle';
        labels.forEach((label, i) => {
            const mid = i * row + row / 2;
            ctx.font = font(13);
            ctx.fillStyle = token('fg');
            ctx.textAlign = 'left';
            ctx.fillText(label, 0, mid, labelW - 12);

            ctx.fillStyle = token('bg-2');
            ctx.fillRect(labelW, mid - bar / 2, trackW, bar);
            ctx.fillStyle = (colorFor || entityColor)(label);
            roundedBar(ctx, labelW, mid - bar / 2, (values[i] / maxVal) * trackW, bar, 4);

            ctx.font = font(13, 'font-mono');
            ctx.fillStyle = token('fg-2');
            ctx.textAlign = 'right';
            ctx.fillText(valueTexts[i], width, mid);
        });
    }

    /**
     * Requests per day as vertical bars with a hover tooltip.
     * @param {string} canvasId
     * @param {string[]} labels - YYYY-MM-DD
     * @param {number[]} values
     */
    function renderVerticalBarChart(canvasId, labels, values) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) return;
        const height = 240;
        const { ctx, width } = setup(canvas, height);
        const padLeft = 48, padBottom = 28, padTop = 8;
        const chartW = width - padLeft, chartH = height - padTop - padBottom;
        const maxVal = niceMax(Math.max(...values, 1));
        const slot = chartW / labels.length;
        const barW = Math.max(2, Math.min(slot - 2, 32));

        // Recessive grid and axis labels
        ctx.font = font(11, 'font-mono');
        ctx.textBaseline = 'middle';
        ctx.textAlign = 'right';
        for (let i = 0; i <= 4; i++) {
            const v = (maxVal / 4) * i;
            const y = padTop + chartH - (v / maxVal) * chartH;
            ctx.fillStyle = token('line');
            ctx.fillRect(padLeft, Math.round(y) + 0.5, chartW, 1);
            ctx.fillStyle = token('fg-3');
            ctx.fillText(Math.round(v).toLocaleString(), padLeft - 8, y);
        }

        const bars = labels.map((label, i) => {
            const x = padLeft + i * slot + (slot - barW) / 2;
            const h = (values[i] / maxVal) * chartH;
            return { x, y: padTop + chartH - h, w: barW, h, label, value: values[i] };
        });
        ctx.fillStyle = token('series-1');
        for (const b of bars) {
            ctx.save();
            ctx.translate(b.x, b.y + b.h);
            ctx.rotate(-Math.PI / 2);
            roundedBar(ctx, 0, 0, b.h, b.w, 4); // drawn sideways: rounded top only
            ctx.restore();
        }

        // Date labels without collisions: every n-th day
        ctx.fillStyle = token('fg-3');
        ctx.textAlign = 'center';
        ctx.textBaseline = 'top';
        const every = Math.ceil(44 / slot);
        bars.forEach((b, i) => {
            if (i % every === 0) ctx.fillText(b.label.slice(5), b.x + b.w / 2, padTop + chartH + 8);
        });

        canvas.onmousemove = (e) => {
            const r = canvas.getBoundingClientRect();
            const i = Math.floor((e.clientX - r.left - padLeft) / slot);
            if (i < 0 || i >= bars.length) return Tooltip.hide();
            Tooltip.show(e, bars[i].label + ': ' + bars[i].value.toLocaleString() + ' requests');
        };
        canvas.onmouseleave = Tooltip.hide;
    }

    function niceMax(v) {
        const mag = Math.pow(10, Math.floor(Math.log10(v)));
        for (const m of [1, 2, 2.5, 5, 10]) if (m * mag >= v) return m * mag;
        return 10 * mag;
    }

    return { renderBarChart, renderVerticalBarChart, statusColor, entityColor, token };
})();

/** One shared tooltip for charts and the map. */
const Tooltip = (() => {
    const el = () => document.getElementById('tooltip');
    return {
        show(event, text) {
            const t = el();
            t.textContent = text;
            t.hidden = false;
            t.style.left = (event.clientX + 12) + 'px';
            t.style.top = (event.clientY - 32) + 'px';
        },
        hide() { el().hidden = true; },
    };
})();
