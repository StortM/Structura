let currentSchema = null;
let svg = d3.select('#diagram');
let zoom = d3.zoom()
    .scaleExtent([0.1, 4])
    .on('zoom', handleZoom);

function init() {
    console.log("Initializing application...");
    
    svg.call(zoom);
    
    const importForm = document.getElementById('importForm');
    console.log("Import form found:", importForm !== null);
    
    if (importForm) {
        importForm.addEventListener('submit', function(e) {
            console.log("Import form submitted");
            e.preventDefault();
            handleImport(e);
        });
    } else {
        console.error("Import form not found!");
    }
    
    document.getElementById('applyLayoutBtn').addEventListener('click', applyLayout);
    document.getElementById('loadDiagramBtn').addEventListener('click', showLoadDiagramModal);
    document.getElementById('exportSVGBtn').addEventListener('click', exportSVG);
    document.getElementById('exportPNGBtn').addEventListener('click', exportPNG);
    document.getElementById('zoomInBtn').addEventListener('click', zoomIn);
    document.getElementById('zoomOutBtn').addEventListener('click', zoomOut);
    document.getElementById('resetZoomBtn').addEventListener('click', resetZoom);
    
    clearDiagram();
    
    console.log("Initialization complete");
}

function handleZoom(e) {
    d3.select('#diagram g').attr('transform', e.transform);
}

function zoomIn() {
    svg.transition().call(zoom.scaleBy, 1.2);
}

function zoomOut() {
    svg.transition().call(zoom.scaleBy, 0.8);
}

function resetZoom() {
    svg.transition().call(zoom.transform, d3.zoomIdentity);
}

function clearDiagram() {
    svg.selectAll('*').remove();
    svg.append('g');
}

async function handleImport(e) {
    console.log("handleImport function called");
    e.preventDefault();
    
    const sqlInput = document.getElementById('sqlInput').value.trim();
    console.log("SQL input length:", sqlInput.length);
    
    if (!sqlInput) {
        alert('Please enter SQL schema.');
        return;
    }
    
    document.getElementById('loading').classList.remove('d-none');
    console.log("Loading indicator shown");
    
    try {
        console.log("Sending request to server...");
        const response = await fetch('/api/schemas', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ sql: sqlInput })
        });
        
        console.log("Response received:", response.status);
        
        if (!response.ok) {
            const error = await response.text();
            throw new Error(`Import failed: ${error}`);
        }
        
        currentSchema = await response.json();
        console.log("Schema parsed successfully:", currentSchema.id);
        renderDiagram(currentSchema);
        
        document.getElementById('applyLayoutBtn').disabled = false;
        document.getElementById('exportSVGBtn').disabled = false;
        document.getElementById('exportPNGBtn').disabled = false;
    } catch (error) {
        console.error('Error importing schema:', error);
        alert(`Error: ${error.message}`);
    } finally {
        document.getElementById('loading').classList.add('d-none');
        console.log("Loading indicator hidden");
    }
}

async function applyLayout() {
    if (!currentSchema) return;
    
    const layoutType = document.querySelector('input[name="layoutAlgorithm"]:checked').value;
    
    document.getElementById('loading').classList.remove('d-none');
    
    try {
        const response = await fetch(`/api/schemas/${currentSchema.id}/layout`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ layoutType })
        });
        
        if (!response.ok) {
            const error = await response.text();
            throw new Error(`Layout failed: ${error}`);
        }
        
        currentSchema = await response.json();
        renderDiagram(currentSchema);
    } catch (error) {
        console.error('Error applying layout:', error);
        alert(`Error: ${error.message}`);
    } finally {
        document.getElementById('loading').classList.add('d-none');
    }
}

async function showLoadDiagramModal() {
    const modal = new bootstrap.Modal(document.getElementById('loadDiagramModal'));
    
    try {
        const response = await fetch('/api/schemas');
        if (!response.ok) {
            throw new Error('Failed to fetch saved diagrams');
        }
        
        const schemas = await response.json();
        const list = document.getElementById('savedDiagramsList');
        
        if (schemas.length === 0) {
            list.innerHTML = '<p>No saved diagrams found.</p>';
        } else {
            let html = '<ul class="list-group">';
            schemas.forEach(schema => {
                const date = new Date(schema.updatedAt * 1000).toLocaleString();
                html += `
                    <li class="list-group-item d-flex justify-content-between align-items-center">
                        <div>
                            <h6>${schema.name}</h6>
                            <small>Last updated: ${date}</small>
                        </div>
                        <button class="btn btn-sm btn-primary load-schema-btn" data-id="${schema.id}">Load</button>
                    </li>
                `;
            });
            html += '</ul>';
            list.innerHTML = html;
            
            document.querySelectorAll('.load-schema-btn').forEach(btn => {
                btn.addEventListener('click', () => {
                    loadSchema(btn.dataset.id);
                    modal.hide();
                });
            });
        }
    } catch (error) {
        console.error('Error fetching saved diagrams:', error);
        document.getElementById('savedDiagramsList').innerHTML = `
            <div class="alert alert-danger">
                Error loading diagrams: ${error.message}
            </div>
        `;
    }
    
    modal.show();
}

async function loadSchema(id) {
    document.getElementById('loading').classList.remove('d-none');
    
    try {
        const response = await fetch(`/api/schemas/${id}`);
        if (!response.ok) {
            throw new Error('Failed to load diagram');
        }
        
        currentSchema = await response.json();
        renderDiagram(currentSchema);

        document.getElementById('applyLayoutBtn').disabled = false;
        document.getElementById('exportSVGBtn').disabled = false;
        document.getElementById('exportPNGBtn').disabled = false;
    } catch (error) {
        console.error('Error loading schema:', error);
        alert(`Error: ${error.message}`);
    } finally {
        document.getElementById('loading').classList.add('d-none');
    }
}

function renderDiagram(schema) {
    clearDiagram();
    
    const g = svg.select('g');
    
    schema.relationships.forEach(rel => {
        const sourceTable = schema.tables.find(t => t.id === rel.sourceTable);
        const targetTable = schema.tables.find(t => t.id === rel.targetTable);
        
        if (!sourceTable || !targetTable) return;
        
        const line = d3.line()
            .x(d => d.x)
            .y(d => d.y)
            .curve(d3.curveLinear);
        
        g.append('path')
            .attr('d', line(rel.points))
            .attr('stroke', '#999')
            .attr('stroke-width', 1.5)
            .attr('fill', 'none')
            .attr('marker-end', 'url(#arrowhead)');
    });
    
    g.append('defs').append('marker')
        .attr('id', 'arrowhead')
        .attr('viewBox', '0 -5 10 10')
        .attr('refX', 8)
        .attr('refY', 0)
        .attr('markerWidth', 6)
        .attr('markerHeight', 6)
        .attr('orient', 'auto')
        .append('path')
        .attr('d', 'M0,-5L10,0L0,5')
        .attr('fill', '#999');
    
    const tables = g.selectAll('.table')
        .data(schema.tables)
        .enter()
        .append('g')
        .attr('class', 'table')
        .attr('transform', d => `translate(${d.position.x}, ${d.position.y})`)
        .call(d3.drag()
            .on('drag', function(event, d) {
                d.position.x += event.dx;
                d.position.y += event.dy;
                d3.select(this).attr('transform', `translate(${d.position.x}, ${d.position.y})`);
                
                updateRelationshipLines(schema, d.id);
                
                debounce(saveTablePosition(d.id, d.position.x, d.position.y), 500);
            })
        );
    
    tables.append('rect')
        .attr('width', d => d.size.width)
        .attr('height', d => d.size.height)
        .attr('fill', '#e6f3ff')
        .attr('stroke', '#4d94ff')
        .attr('rx', 4)
        .attr('ry', 4);
    
    tables.append('rect')
        .attr('width', d => d.size.width)
        .attr('height', 30)
        .attr('fill', '#4d94ff')
        .attr('rx', 4)
        .attr('ry', 4);
    
    tables.append('text')
        .attr('x', d => d.size.width / 2)
        .attr('y', 20)
        .attr('text-anchor', 'middle')
        .attr('fill', 'white')
        .attr('font-weight', 'bold')
        .text(d => d.name);
    
    tables.each(function(d) {
        const table = d3.select(this);
        
        for (let i = 0; i < d.columns.length; i++) {
            const column = d.columns[i];
            const y = 30 + i * 25 + 15;
            
            table.append('rect')
                .attr('x', 0)
                .attr('y', 30 + i * 25)
                .attr('width', d.size.width)
                .attr('height', 25)
                .attr('fill', i % 2 === 0 ? '#f8f9fa' : 'white')
                .attr('opacity', 0.5);
            
            let xPos = 10;
            
            if (column.isPrimaryKey) {
                table.append('text')
                    .attr('x', xPos)
                    .attr('y', y)
                    .attr('fill', 'gold')
                    .attr('font-size', '14px')
                    .text('🔑');
                xPos = 30;
            }
            
            if (column.isForeignKey) {
                table.append('text')
                    .attr('x', xPos)
                    .attr('y', y)
                    .attr('fill', 'silver')
                    .attr('font-size', '14px')
                    .text('🔗');
                xPos = column.isPrimaryKey ? 50 : 30;
            }
            
            table.append('text')
                .attr('x', xPos)
                .attr('y', y)
                .attr('fill', 'black')
                .attr('font-size', '12px')
                .text(column.name);
            
            table.append('text')
                .attr('x', d.size.width - 10)
                .attr('y', y)
                .attr('text-anchor', 'end')
                .attr('fill', '#666')
                .attr('font-size', '12px')
                .text(column.dataType);
        }
    });
}

function updateRelationshipLines(schema, tableId) {
    const g = svg.select('g');
    
    const relatedRels = schema.relationships.filter(
        rel => rel.sourceTable === tableId || rel.targetTable === tableId
    );
    
    relatedRels.forEach(rel => {
        const sourceTable = schema.tables.find(t => t.id === rel.sourceTable);
        const targetTable = schema.tables.find(t => t.id === rel.targetTable);
        
        if (!sourceTable || !targetTable) return;
        
        rel.points = [
            {
                x: sourceTable.position.x + sourceTable.size.width / 2,
                y: sourceTable.position.y + sourceTable.size.height / 2
            },
            {
                x: targetTable.position.x + targetTable.size.width / 2,
                y: targetTable.position.y + targetTable.size.height / 2
            }
        ];
        
        const line = d3.line()
            .x(d => d.x)
            .y(d => d.y)
            .curve(d3.curveLinear);
        
        g.selectAll('path').filter((d, i, nodes) => {
            return i === schema.relationships.indexOf(rel);
        }).attr('d', line(rel.points));
    });
}

async function saveTablePosition(tableId, x, y) {
    if (!currentSchema) return;
    
    try {
        await fetch(`/api/schemas/${currentSchema.id}/tables/${tableId}/position`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ x, y })
        });
    } catch (error) {
        console.error('Error saving table position:', error);
    }
}

function exportSVG() {
    if (!currentSchema) return;
    
    const svgCopy = document.getElementById('diagram').cloneNode(true);
    
    const g = svgCopy.querySelector('g');
    const bbox = g.getBBox();
    
    svgCopy.setAttribute('width', bbox.width);
    svgCopy.setAttribute('height', bbox.height);
    svgCopy.setAttribute('viewBox', `${bbox.x} ${bbox.y} ${bbox.width} ${bbox.height}`);
    
    const svgData = new XMLSerializer().serializeToString(svgCopy);
    const svgBlob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' });
    const url = URL.createObjectURL(svgBlob);
    
    const link = document.createElement('a');
    link.href = url;
    link.download = `${currentSchema.name || 'diagram'}.svg`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
}

function exportPNG() {
    if (!currentSchema) return;
    
    const svgCopy = document.getElementById('diagram').cloneNode(true);
    
    const g = svgCopy.querySelector('g');
    const bbox = g.getBBox();
    
    svgCopy.setAttribute('width', bbox.width);
    svgCopy.setAttribute('height', bbox.height);
    svgCopy.setAttribute('viewBox', `${bbox.x} ${bbox.y} ${bbox.width} ${bbox.height}`);
    
    const svgData = new XMLSerializer().serializeToString(svgCopy);
    const svgBlob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' });
    const url = URL.createObjectURL(svgBlob);
    
    const img = new Image();
    img.onload = function() {

        const canvas = document.createElement('canvas');
        canvas.width = bbox.width;
        canvas.height = bbox.height;
        

        const ctx = canvas.getContext('2d');
        ctx.fillStyle = 'white';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        ctx.drawImage(img, 0, 0);
        

        const pngUrl = canvas.toDataURL('image/png');
        const link = document.createElement('a');
        link.href = pngUrl;
        link.download = `${currentSchema.name || 'diagram'}.png`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    };
    img.src = url;
}
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}
const style = document.createElement('style');
style.textContent = `
    #diagram-container {
        position: relative;
        overflow: hidden;
        background-color: #f8f9fa;
        border: 1px solid #dee2e6;
    }
    
    #loading {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        padding: 1rem 2rem;
        background-color: rgba(0, 0, 0, 0.7);
        color: white;
        border-radius: 4px;
        z-index: 1000;
    }
`;
document.head.appendChild(style);
document.addEventListener('DOMContentLoaded', init);