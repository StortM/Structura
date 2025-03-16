// web/static/js/main.js

// Global variables
let currentSchema = null;
let svg = d3.select('#diagram');
let zoom = d3.zoom()
    .scaleExtent([0.1, 4])
    .on('zoom', handleZoom);

// Initialize the diagram
function init() {
    console.log("Initializing application...");
    
    // Set up zoom behavior
    svg.call(zoom);
    
    // Set up event listeners
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
    
    // Clear any existing diagrams
    clearDiagram();
    
    console.log("Initialization complete");
}

// Handle zoom events
function handleZoom(e) {
    d3.select('#diagram g').attr('transform', e.transform);
}

// Zoom in
function zoomIn() {
    svg.transition().call(zoom.scaleBy, 1.2);
}

// Zoom out
function zoomOut() {
    svg.transition().call(zoom.scaleBy, 0.8);
}

// Reset zoom
function resetZoom() {
    svg.transition().call(zoom.transform, d3.zoomIdentity);
}

// Clear the diagram
function clearDiagram() {
    svg.selectAll('*').remove();
    svg.append('g');
}

// Handle schema import
async function handleImport(e) {
    console.log("handleImport function called");
    e.preventDefault();
    
    const sqlInput = document.getElementById('sqlInput').value.trim();
    console.log("SQL input length:", sqlInput.length);
    
    if (!sqlInput) {
        alert('Please enter SQL schema.');
        return;
    }
    
    // Show loading indicator
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
        
        // Enable buttons
        document.getElementById('applyLayoutBtn').disabled = false;
        document.getElementById('exportSVGBtn').disabled = false;
        document.getElementById('exportPNGBtn').disabled = false;
    } catch (error) {
        console.error('Error importing schema:', error);
        alert(`Error: ${error.message}`);
    } finally {
        // Hide loading indicator
        document.getElementById('loading').classList.add('d-none');
        console.log("Loading indicator hidden");
    }
}

// Apply layout algorithm
async function applyLayout() {
    if (!currentSchema) return;
    
    const layoutType = document.querySelector('input[name="layoutAlgorithm"]:checked').value;
    
    // Show loading indicator
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
        // Hide loading indicator
        document.getElementById('loading').classList.add('d-none');
    }
}

// Show the load diagram modal
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
            
            // Add event listeners to load buttons
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

// Load a saved schema
async function loadSchema(id) {
    // Show loading indicator
    document.getElementById('loading').classList.remove('d-none');
    
    try {
        const response = await fetch(`/api/schemas/${id}`);
        if (!response.ok) {
            throw new Error('Failed to load diagram');
        }
        
        currentSchema = await response.json();
        renderDiagram(currentSchema);
        
        // Enable buttons
        document.getElementById('applyLayoutBtn').disabled = false;
        document.getElementById('exportSVGBtn').disabled = false;
        document.getElementById('exportPNGBtn').disabled = false;
    } catch (error) {
        console.error('Error loading schema:', error);
        alert(`Error: ${error.message}`);
    } finally {
        // Hide loading indicator
        document.getElementById('loading').classList.add('d-none');
    }
}

// Render the diagram
function renderDiagram(schema) {
    // Clear the current diagram
    clearDiagram();
    
    const g = svg.select('g');
    
    // Draw relationships first (so they appear behind tables)
    schema.relationships.forEach(rel => {
        // Find source and target tables
        const sourceTable = schema.tables.find(t => t.id === rel.sourceTable);
        const targetTable = schema.tables.find(t => t.id === rel.targetTable);
        
        if (!sourceTable || !targetTable) return;
        
        // Draw the relationship line
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
    
    // Create a marker for the arrowhead
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
    
    // Draw tables
    const tables = g.selectAll('.table')
        .data(schema.tables)
        .enter()
        .append('g')
        .attr('class', 'table')
        .attr('transform', d => `translate(${d.position.x}, ${d.position.y})`)
        .call(d3.drag()
            .on('drag', function(event, d) {
                // Update position visually
                d.position.x += event.dx;
                d.position.y += event.dy;
                d3.select(this).attr('transform', `translate(${d.position.x}, ${d.position.y})`);
                
                // Update relationship lines
                updateRelationshipLines(schema, d.id);
                
                // Save position to server (debounced)
                debounce(saveTablePosition(d.id, d.position.x, d.position.y), 500);
            })
        );
    
    // Draw table rectangles
    tables.append('rect')
        .attr('width', d => d.size.width)
        .attr('height', d => d.size.height)
        .attr('fill', '#e6f3ff')
        .attr('stroke', '#4d94ff')
        .attr('rx', 4)
        .attr('ry', 4);
    
    // Draw table headers
    tables.append('rect')
        .attr('width', d => d.size.width)
        .attr('height', 30)
        .attr('fill', '#4d94ff')
        .attr('rx', 4)
        .attr('ry', 4);
    
    // Draw table names
    tables.append('text')
        .attr('x', d => d.size.width / 2)
        .attr('y', 20)
        .attr('text-anchor', 'middle')
        .attr('fill', 'white')
        .attr('font-weight', 'bold')
        .text(d => d.name);
    
    // Draw table columns
    tables.each(function(d) {
        const table = d3.select(this);
        
        // For each column
        for (let i = 0; i < d.columns.length; i++) {
            const column = d.columns[i];
            const y = 30 + i * 25 + 15;
            
            // Draw row background
            table.append('rect')
                .attr('x', 0)
                .attr('y', 30 + i * 25)
                .attr('width', d.size.width)
                .attr('height', 25)
                .attr('fill', i % 2 === 0 ? '#f8f9fa' : 'white')
                .attr('opacity', 0.5);
            
            // Set initial position for text
            let xPos = 10;
            
            // Add primary key icon if needed
            if (column.isPrimaryKey) {
                table.append('text')
                    .attr('x', xPos)
                    .attr('y', y)
                    .attr('fill', 'gold')
                    .attr('font-size', '14px')
                    .text('🔑');
                xPos = 30;
            }
            
            // Add foreign key icon if needed
            if (column.isForeignKey) {
                table.append('text')
                    .attr('x', xPos)
                    .attr('y', y)
                    .attr('fill', 'silver')
                    .attr('font-size', '14px')
                    .text('🔗');
                xPos = column.isPrimaryKey ? 50 : 30;
            }
            
            // Draw column name
            table.append('text')
                .attr('x', xPos)
                .attr('y', y)
                .attr('fill', 'black')
                .attr('font-size', '12px')
                .text(column.name);
            
            // Draw column type
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

// Update relationship lines when a table is moved
function updateRelationshipLines(schema, tableId) {
    const g = svg.select('g');
    
    // Find all relationships involving this table
    const relatedRels = schema.relationships.filter(
        rel => rel.sourceTable === tableId || rel.targetTable === tableId
    );
    
    relatedRels.forEach(rel => {
        // Find source and target tables
        const sourceTable = schema.tables.find(t => t.id === rel.sourceTable);
        const targetTable = schema.tables.find(t => t.id === rel.targetTable);
        
        if (!sourceTable || !targetTable) return;
        
        // Update relationship points (simplified - just connecting centers)
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
        
        // Update the visual path
        const line = d3.line()
            .x(d => d.x)
            .y(d => d.y)
            .curve(d3.curveLinear);
        
        g.selectAll('path').filter((d, i, nodes) => {
            return i === schema.relationships.indexOf(rel);
        }).attr('d', line(rel.points));
    });
}

// Save table position to the server
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

// Export diagram as SVG
function exportSVG() {
    if (!currentSchema) return;
    
    // Clone the SVG
    const svgCopy = document.getElementById('diagram').cloneNode(true);
    
    // Get the diagram's bounding box
    const g = svgCopy.querySelector('g');
    const bbox = g.getBBox();
    
    // Set the SVG dimensions to match the content
    svgCopy.setAttribute('width', bbox.width);
    svgCopy.setAttribute('height', bbox.height);
    svgCopy.setAttribute('viewBox', `${bbox.x} ${bbox.y} ${bbox.width} ${bbox.height}`);
    
    // Convert SVG to a data URL
    const svgData = new XMLSerializer().serializeToString(svgCopy);
    const svgBlob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' });
    const url = URL.createObjectURL(svgBlob);
    
    // Create a download link
    const link = document.createElement('a');
    link.href = url;
    link.download = `${currentSchema.name || 'diagram'}.svg`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
}

// Export diagram as PNG
function exportPNG() {
    if (!currentSchema) return;
    
    // Clone the SVG
    const svgCopy = document.getElementById('diagram').cloneNode(true);
    
    // Get the diagram's bounding box
    const g = svgCopy.querySelector('g');
    const bbox = g.getBBox();
    
    // Set the SVG dimensions to match the content
    svgCopy.setAttribute('width', bbox.width);
    svgCopy.setAttribute('height', bbox.height);
    svgCopy.setAttribute('viewBox', `${bbox.x} ${bbox.y} ${bbox.width} ${bbox.height}`);
    
    // Convert SVG to a data URL
    const svgData = new XMLSerializer().serializeToString(svgCopy);
    const svgBlob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' });
    const url = URL.createObjectURL(svgBlob);
    
    // Create an image from the SVG
    const img = new Image();
    img.onload = function() {
        // Create a canvas to draw the image
        const canvas = document.createElement('canvas');
        canvas.width = bbox.width;
        canvas.height = bbox.height;
        
        // Draw the image onto the canvas
        const ctx = canvas.getContext('2d');
        ctx.fillStyle = 'white';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        ctx.drawImage(img, 0, 0);
        
        // Convert canvas to PNG and download
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

// Debounce helper function
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

// Add styles for the diagram
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

// Initialize the application when the page loads
document.addEventListener('DOMContentLoaded', init);