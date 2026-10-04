/**
 * "Licenses and thanks": renders static/licenses/credits.json, which a Go test
 * keeps complete. The license texts are embedded in the binary beside it.
 */
(function () {
    const SECTIONS = [
        ['inside', 'Built into CaddyShack', 'These are inside the program. Their license texts come with it.'],
        ['data', 'Data', 'Sources of the data CaddyShack shows.'],
    ];

    function el(tag, className, text) {
        const e = document.createElement(tag);
        if (className) e.className = className;
        if (text) e.textContent = text;
        return e;
    }

    function link(href, label) {
        const a = el('a', null, label);
        a.href = href;
        if (href.startsWith('http')) { a.target = '_blank'; a.rel = 'noopener noreferrer'; }
        return a;
    }

    function item(c) {
        const li = el('li', 'credit');
        const name = el('h3', null, c.name);
        if (c.version) name.append(' ', el('span', 'credit-version mono', c.version));
        const links = el('p', 'credit-links');
        links.append(link(c.home, 'Website'));
        if (c.support) links.append(link(c.support, 'Support the project'));
        if (c.text) links.append(link(c.text, 'License text'));
        li.append(name, el('p', 'credit-use', c.use), el('p', 'credit-license mono', c.license), links);
        return li;
    }

    async function render() {
        const host = document.getElementById('credits');
        try {
            const resp = await fetch('/licenses/credits.json');
            if (!resp.ok) throw new Error('the server answered ' + resp.status);
            const credits = await resp.json();
            host.replaceChildren(...SECTIONS.map(([key, title, note]) => {
                const section = el('section', 'credit-section');
                const list = el('ul', 'credit-list');
                list.append(...(credits[key] || []).map(item));
                section.append(el('h2', null, title), el('p', 'reading-note', note), list);
                return section;
            }));
        } catch (err) {
            host.replaceChildren(el('p', 'message', 'The list could not be read: ' + err.message + '. Reload the page to try again.'));
        }
    }

    render();
})();
