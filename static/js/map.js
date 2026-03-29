/**
 * WorldMap module — D3.js world map with bubble overlay sized by request count.
 */
const WorldMap = (() => {
    let tooltip = null;
    let svgNode = null;

    // ISO 3166-1 numeric → alpha-2 mapping (for TopoJSON which uses numeric IDs)
    const numToAlpha2 = {
        4:'AF',8:'AL',12:'DZ',24:'AO',32:'AR',36:'AU',40:'AT',50:'BD',56:'BE',
        64:'BT',68:'BO',70:'BA',72:'BW',76:'BR',100:'BG',104:'MM',108:'BI',
        112:'BY',116:'KH',120:'CM',124:'CA',140:'CF',144:'LK',148:'TD',152:'CL',
        156:'CN',170:'CO',178:'CG',180:'CD',188:'CR',191:'HR',192:'CU',196:'CY',
        203:'CZ',204:'BJ',208:'DK',214:'DO',218:'EC',818:'EG',222:'SV',226:'GQ',
        232:'ER',233:'EE',231:'ET',238:'FK',242:'FJ',246:'FI',250:'FR',254:'GF',
        266:'GA',270:'GM',268:'GE',276:'DE',288:'GH',300:'GR',304:'GL',320:'GT',
        324:'GN',328:'GY',332:'HT',340:'HN',348:'HU',352:'IS',356:'IN',360:'ID',
        364:'IR',368:'IQ',372:'IE',376:'IL',380:'IT',384:'CI',388:'JM',392:'JP',
        398:'KZ',400:'JO',404:'KE',408:'KP',410:'KR',414:'KW',417:'KG',418:'LA',
        422:'LB',426:'LS',428:'LV',430:'LR',434:'LY',440:'LT',442:'LU',450:'MG',
        454:'MW',458:'MY',466:'ML',478:'MR',484:'MX',496:'MN',498:'MD',504:'MA',
        508:'MZ',512:'OM',516:'NA',524:'NP',528:'NL',540:'NC',554:'NZ',558:'NI',
        562:'NE',566:'NG',578:'NO',586:'PK',591:'PA',598:'PG',600:'PY',604:'PE',
        608:'PH',616:'PL',620:'PT',630:'PR',634:'QA',642:'RO',643:'RU',646:'RW',
        682:'SA',686:'SN',688:'RS',694:'SL',702:'SG',703:'SK',704:'VN',705:'SI',
        706:'SO',710:'ZA',716:'ZW',724:'ES',728:'SS',729:'SD',740:'SR',752:'SE',
        756:'CH',760:'SY',762:'TJ',764:'TH',768:'TG',780:'TT',788:'TN',792:'TR',
        795:'TM',800:'UG',804:'UA',784:'AE',826:'GB',834:'TZ',840:'US',858:'UY',
        860:'UZ',862:'VE',887:'YE',894:'ZM',
        // Territories / special
        10:'AQ',158:'TW',275:'PS',732:'EH',807:'MK',499:'ME',344:'HK'
    };

    // Approximate centroids (lon, lat) for countries — used for bubble placement
    const centroids = {
        AF:[67,33],AL:[20,41],DZ:[3,28],AO:[18,-12],AR:[-64,-34],AU:[134,-25],
        AT:[14,47],BD:[90,24],BE:[4,51],BT:[90,28],BO:[-65,-17],BA:[18,44],
        BW:[24,-22],BR:[-53,-10],BG:[25,43],MM:[97,20],BI:[30,-3],BY:[28,53],
        KH:[105,13],CM:[12,6],CA:[-96,62],CF:[21,7],LK:[81,8],TD:[19,15],
        CL:[-71,-35],CN:[104,35],CO:[-72,4],CG:[16,-1],CD:[24,-3],CR:[-84,10],
        HR:[16,45],CU:[-79,22],CY:[33,35],CZ:[15,50],BJ:[2,10],DK:[10,56],
        DO:[-70,19],EC:[-78,-2],EG:[30,27],SV:[-89,14],GQ:[10,2],ER:[39,15],
        EE:[26,59],ET:[40,9],FI:[26,64],FR:[2,46],GA:[12,-1],GM:[-16,13],
        GE:[44,42],DE:[10,51],GH:[-2,8],GR:[22,39],GL:[-42,72],GT:[-90,16],
        GN:[-12,11],GY:[-59,5],HT:[-72,19],HN:[-86,15],HU:[20,47],IS:[-19,65],
        IN:[79,22],ID:[118,-2],IR:[53,32],IQ:[44,33],IE:[-8,53],IL:[35,31],
        IT:[12,42],CI:[-6,8],JM:[-77,18],JP:[138,36],KZ:[67,48],JO:[36,31],
        KE:[38,0],KP:[127,40],KR:[128,36],KW:[48,29],KG:[75,41],LA:[103,18],
        LB:[36,34],LS:[29,-30],LV:[25,57],LR:[-10,7],LY:[17,27],LT:[24,56],
        LU:[6,50],MG:[47,-19],MW:[34,-14],MY:[110,4],ML:[-4,17],MR:[-11,20],
        MX:[-102,24],MN:[104,47],MD:[29,47],MA:[-6,32],MZ:[35,-18],OM:[56,21],
        NA:[18,-22],NP:[84,28],NL:[6,52],NZ:[174,-41],NI:[-85,13],NE:[8,16],
        NG:[8,10],NO:[9,62],PK:[69,30],PA:[-80,9],PG:[147,-6],PY:[-58,-23],
        PE:[-76,-10],PH:[122,12],PL:[20,52],PT:[-8,40],PR:[-66,18],QA:[51,25],
        RO:[25,46],RU:[100,60],RW:[30,-2],SA:[45,24],SN:[-14,15],RS:[21,44],
        SG:[104,1],SK:[20,49],VN:[108,16],SI:[15,46],SO:[46,6],ZA:[25,-29],
        ZW:[30,-20],ES:[-4,40],SS:[30,7],SD:[30,15],SR:[-56,4],SE:[16,62],
        CH:[8,47],SY:[38,35],TJ:[69,39],TH:[101,15],TG:[1,8],TT:[-61,11],
        TN:[9,34],TR:[35,39],TM:[59,39],UG:[32,1],UA:[32,49],AE:[54,24],
        GB:[-3,54],TZ:[35,-6],US:[-97,38],UY:[-56,-33],UZ:[65,41],VE:[-67,8],
        YE:[48,15],ZM:[28,-14],TW:[121,24],PS:[35,32],MK:[22,41],ME:[19,43],
        HK:[114,22],AQ:[0,-82],EH:[-13,24],
    };

    function render(containerId, countries) {
        const container = document.getElementById(containerId);
        if (!container) return;
        tooltip = document.getElementById('map-tooltip');

        // Build count lookup by alpha-2 code
        const countMap = {};
        let maxCount = 0;
        for (const c of countries) {
            if (c.code && c.code !== '??') {
                countMap[c.code] = c;
                if (c.count > maxCount) maxCount = c.count;
            }
        }

        // Load TopoJSON
        d3.json('/data/countries-110m.json').then(world => {
            const geoCountries = topojson.feature(world, world.objects.countries);
            const borders = topojson.mesh(world, world.objects.countries, (a, b) => a !== b);

            container.innerHTML = '';
            const width = container.clientWidth;
            const height = Math.round(width * 0.52);

            const svg = d3.select(container).append('svg')
                .attr('width', width)
                .attr('height', height)
                .attr('viewBox', `0 0 ${width} ${height}`)
                .style('background', '#f8fbff');

            svgNode = svg.node();

            const projection = d3.geoNaturalEarth1()
                .fitSize([width - 20, height - 20], { type: 'Sphere' })
                .translate([width / 2, height / 2]);

            const path = d3.geoPath().projection(projection);

            // Draw country shapes
            svg.append('g')
                .selectAll('path')
                .data(geoCountries.features)
                .join('path')
                .attr('d', path)
                .attr('fill', '#e8e8e8')
                .attr('stroke', '#ccc')
                .attr('stroke-width', 0.5);

            // Draw borders
            svg.append('path')
                .datum(borders)
                .attr('d', path)
                .attr('fill', 'none')
                .attr('stroke', '#bbb')
                .attr('stroke-width', 0.5);

            // Graticule
            svg.append('path')
                .datum(d3.geoGraticule10())
                .attr('d', path)
                .attr('fill', 'none')
                .attr('stroke', '#e0e8f0')
                .attr('stroke-width', 0.3);

            // Outline
            svg.append('path')
                .datum({ type: 'Sphere' })
                .attr('d', path)
                .attr('fill', 'none')
                .attr('stroke', '#aaa')
                .attr('stroke-width', 0.8);

            // Bubbles
            const bubbleData = [];
            for (const code in countMap) {
                const coords = centroids[code];
                if (coords) {
                    const projected = projection(coords);
                    if (projected) {
                        bubbleData.push({
                            code: code,
                            name: countMap[code].name,
                            count: countMap[code].count,
                            x: projected[0],
                            y: projected[1],
                        });
                    }
                }
            }

            // Size scale: sqrt for area-proportional bubbles
            const maxRadius = Math.min(width, height) * 0.06;
            const radiusScale = d3.scaleSqrt()
                .domain([0, maxCount])
                .range([3, maxRadius]);

            // Sort: larger bubbles behind smaller ones
            bubbleData.sort((a, b) => b.count - a.count);

            svg.append('g')
                .selectAll('circle')
                .data(bubbleData)
                .join('circle')
                .attr('cx', d => d.x)
                .attr('cy', d => d.y)
                .attr('r', d => radiusScale(d.count))
                .attr('fill', 'rgba(74, 140, 63, 0.55)')
                .attr('stroke', '#2d5a27')
                .attr('stroke-width', 0.8)
                .on('mouseenter', (event, d) => {
                    tooltip.textContent = `${d.name} (${d.code}): ${d.count.toLocaleString()} requests`;
                    tooltip.classList.remove('hidden');
                    d3.select(event.target).attr('fill', 'rgba(139, 195, 74, 0.75)');
                })
                .on('mousemove', (event) => {
                    tooltip.style.left = (event.clientX + 14) + 'px';
                    tooltip.style.top = (event.clientY - 10) + 'px';
                })
                .on('mouseleave', (event) => {
                    tooltip.classList.add('hidden');
                    d3.select(event.target).attr('fill', 'rgba(74, 140, 63, 0.55)');
                });

            // Legend
            if (bubbleData.length > 0) {
                const legendValues = [
                    Math.round(maxCount),
                    Math.round(maxCount / 4),
                    Math.round(maxCount / 16) || 1
                ].filter((v, i, a) => a.indexOf(v) === i && v > 0);

                const legend = svg.append('g')
                    .attr('transform', `translate(${width - 80}, ${height - 20 - legendValues.length * 28})`);

                legend.selectAll('circle')
                    .data(legendValues)
                    .join('circle')
                    .attr('cx', 0)
                    .attr('cy', (d, i) => i * 28)
                    .attr('r', d => radiusScale(d))
                    .attr('fill', 'none')
                    .attr('stroke', '#2d5a27')
                    .attr('stroke-width', 0.8);

                legend.selectAll('text')
                    .data(legendValues)
                    .join('text')
                    .attr('x', maxRadius + 8)
                    .attr('y', (d, i) => i * 28 + 4)
                    .text(d => d.toLocaleString())
                    .attr('font-size', '10px')
                    .attr('fill', '#666');
            }
        }).catch(() => {
            container.innerHTML = '<p style="color:#999;text-align:center;padding:2rem;">World map data not available</p>';
        });
    }

    return { render };
})();
