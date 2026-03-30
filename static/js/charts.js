/**
 * Charts module — Canvas 2D bar charts for CaddyShack dashboard.
 */

const Charts = (() => {
    const COLORS = {
        green: '#4a8c3f',
        greenLight: '#8bc34a',
        greenDark: '#2d5a27',
        gray: '#e0e0e0',
        text: '#333333',
        textLight: '#666666',
    };

    /**
     * Render a horizontal bar chart on a canvas element.
     * @param {string} canvasId - Canvas element ID
     * @param {string[]} labels - Bar labels
     * @param {number[]} values - Bar values
     * @param {number} total - Total for percentage calculation (0 to skip %)
     * @param {string} color - Bar fill color
     */
    function renderBarChart(canvasId, labels, values, total, color) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        const dpr = window.devicePixelRatio || 1;

        const barHeight = 24;
        const gap = 6;
        const labelWidth = 120;
        const valueWidth = 100;
        const padTop = 8;
        const padBottom = 8;

        const height = padTop + labels.length * (barHeight + gap) + padBottom;
        const width = canvas.parentElement.clientWidth - 32;

        canvas.width = width * dpr;
        canvas.height = height * dpr;
        canvas.style.width = width + 'px';
        canvas.style.height = height + 'px';
        ctx.scale(dpr, dpr);

        const maxVal = Math.max(...values, 1);
        const barAreaWidth = width - labelWidth - valueWidth;

        ctx.font = '13px -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif';
        ctx.textBaseline = 'middle';

        for (let i = 0; i < labels.length; i++) {
            const y = padTop + i * (barHeight + gap);
            const barW = (values[i] / maxVal) * barAreaWidth;
            const pct = total > 0 ? ((values[i] / total) * 100).toFixed(1) + '%' : '';

            // Label
            ctx.fillStyle = COLORS.text;
            ctx.textAlign = 'right';
            ctx.fillText(labels[i], labelWidth - 10, y + barHeight / 2);

            // Bar background
            ctx.fillStyle = COLORS.gray;
            ctx.fillRect(labelWidth, y, barAreaWidth, barHeight);

            // Bar fill
            ctx.fillStyle = color || COLORS.green;
            ctx.fillRect(labelWidth, y, barW, barHeight);

            // Value text
            ctx.fillStyle = COLORS.textLight;
            ctx.textAlign = 'left';
            const valText = values[i].toLocaleString() + (pct ? '  ' + pct : '');
            ctx.fillText(valText, labelWidth + barAreaWidth + 8, y + barHeight / 2);
        }
    }

    /**
     * Render a vertical bar chart (for daily traffic).
     * @param {string} canvasId
     * @param {string[]} labels - Date labels
     * @param {number[]} values
     */
    function renderVerticalBarChart(canvasId, labels, values) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        const dpr = window.devicePixelRatio || 1;

        const padLeft = 60;
        const padRight = 20;
        const padTop = 20;
        const padBottom = 60;
        const width = canvas.parentElement.clientWidth - 32;
        const height = 250;

        canvas.width = width * dpr;
        canvas.height = height * dpr;
        canvas.style.width = width + 'px';
        canvas.style.height = height + 'px';
        ctx.scale(dpr, dpr);

        const chartW = width - padLeft - padRight;
        const chartH = height - padTop - padBottom;
        const maxVal = Math.max(...values, 1);
        const barWidth = Math.max(4, (chartW / labels.length) - 4);
        const barGap = (chartW - barWidth * labels.length) / (labels.length + 1);

        ctx.font = '11px -apple-system, BlinkMacSystemFont, sans-serif';

        // Y axis gridlines
        const ySteps = 5;
        ctx.strokeStyle = COLORS.gray;
        ctx.fillStyle = COLORS.textLight;
        ctx.textAlign = 'right';
        ctx.textBaseline = 'middle';
        for (let i = 0; i <= ySteps; i++) {
            const val = Math.round((maxVal / ySteps) * i);
            const y = padTop + chartH - (chartH * (val / maxVal));
            ctx.beginPath();
            ctx.moveTo(padLeft, y);
            ctx.lineTo(padLeft + chartW, y);
            ctx.stroke();
            ctx.fillText(val.toLocaleString(), padLeft - 8, y);
        }

        // Bars
        for (let i = 0; i < labels.length; i++) {
            const x = padLeft + barGap + i * (barWidth + barGap);
            const barH = (values[i] / maxVal) * chartH;
            const y = padTop + chartH - barH;

            ctx.fillStyle = COLORS.green;
            ctx.fillRect(x, y, barWidth, barH);

            // Date label
            ctx.save();
            ctx.fillStyle = COLORS.textLight;
            ctx.textAlign = 'right';
            ctx.textBaseline = 'top';
            ctx.translate(x + barWidth / 2, padTop + chartH + 8);
            ctx.rotate(-Math.PI / 4);
            // Show short date: MM-DD
            const shortLabel = labels[i].substring(5);
            ctx.fillText(shortLabel, 0, 0);
            ctx.restore();
        }
    }

    /**
     * Render a D3 donut/pie chart with a legend below.
     * @param {string} containerId - Container div element ID
     * @param {string[]} labels - Slice labels
     * @param {number[]} values - Slice values
     * @param {number} total - Total for percentage calculation
     */
    function renderPieChart(containerId, labels, values, total) {
        const container = document.getElementById(containerId);
        if (!container) return;
        container.innerHTML = '';

        const PIE_GREENS = [
            '#1a4d15', '#276120', '#3d7a35', '#4a8c3f',
            '#5ea854', '#7ab864', '#8bc34a', '#a8d467',
        ];

        const w = container.clientWidth || 300;
        const pieD = Math.min(w, 200);
        const r = pieD / 2 - 4;

        const pie = d3.pie().sort(null).value(d => d);
        const arc = d3.arc().innerRadius(r * 0.38).outerRadius(r);
        const arcs = pie(values);

        const svg = d3.select(container)
            .append('svg')
            .attr('viewBox', `0 0 ${w} ${pieD}`)
            .attr('width', '100%')
            .style('display', 'block');

        const g = svg.append('g')
            .attr('transform', `translate(${w / 2},${pieD / 2})`);

        g.selectAll('path')
            .data(arcs)
            .join('path')
            .attr('d', arc)
            .attr('fill', (d, i) => PIE_GREENS[i % PIE_GREENS.length])
            .attr('stroke', '#fff')
            .attr('stroke-width', 1.5)
            .append('title')
            .text((d, i) => {
                const pct = total > 0 ? ' (' + ((values[i] / total) * 100).toFixed(1) + '%)' : '';
                return labels[i] + ': ' + values[i].toLocaleString() + pct;
            });

        // Legend
        const legend = document.createElement('div');
        legend.className = 'pie-legend';
        for (let i = 0; i < labels.length; i++) {
            const pct = total > 0 ? ((values[i] / total) * 100).toFixed(1) + '%' : '';
            const item = document.createElement('div');
            item.className = 'pie-legend-item';
            item.innerHTML =
                `<span class="pie-legend-swatch" style="background:${PIE_GREENS[i % PIE_GREENS.length]}"></span>` +
                `<span class="pie-legend-label">${labels[i]}</span>` +
                `<span class="pie-legend-pct">${pct}</span>`;
            legend.appendChild(item);
        }
        container.appendChild(legend);
    }

    return { renderBarChart, renderVerticalBarChart, renderPieChart };
})();
