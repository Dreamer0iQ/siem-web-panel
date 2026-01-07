import { useEffect, useRef } from 'react';
import * as d3 from 'd3';

export interface NetworkNode {
    id: string;
    group: string;
    label: string;
    value: number;
}

export interface NetworkLink {
    source: string;
    target: string;
    value: number;
}

interface Props {
    nodes: NetworkNode[];
    links: NetworkLink[];
    width?: number;
    height?: number;
}

export const NetworkGraph: React.FC<Props> = ({
    nodes,
    links,
    width = 800,
    height = 600
}) => {
    const svgRef = useRef<SVGSVGElement>(null);

    useEffect(() => {
        if (!svgRef.current || nodes.length === 0) return;

        // Очистка предыдущего графа
        d3.select(svgRef.current).selectAll('*').remove();

        const svg = d3.select(svgRef.current)
            .attr('width', width)
            .attr('height', height)
            .attr('viewBox', [0, 0, width, height]);

        // Создание группы для zoom/pan
        const g = svg.append('g');

        // Color scale для разных типов узлов
        const color = d3.scaleOrdinal(d3.schemeCategory10);

        // Создание force simulation
        const simulation = d3.forceSimulation(nodes as any)
            .force('link', d3.forceLink(links as any)
                .id((d: any) => d.id)
                .distance(100))
            .force('charge', d3.forceManyBody().strength(-300))
            .force('center', d3.forceCenter(width / 2, height / 2))
            .force('collision', d3.forceCollide().radius(30));

        // Создание связей
        const link = g.append('g')
            .selectAll('line')
            .data(links)
            .join('line')
            .attr('stroke', '#999')
            .attr('stroke-opacity', 0.6)
            .attr('stroke-width', (d: any) => Math.sqrt(d.value));

        // Создание узлов
        const node = g.append('g')
            .selectAll('circle')
            .data(nodes)
            .join('circle')
            .attr('r', (d: any) => Math.sqrt(d.value) * 3)
            .attr('fill', (d: any) => color(d.group))
            .attr('stroke', '#fff')
            .attr('stroke-width', 2)
            .call(drag(simulation) as any);

        // Добавление тултипов
        node.append('title')
            .text((d: any) => `${d.label}\n${d.value} events`);

        // Добавление лейблов
        const label = g.append('g')
            .selectAll('text')
            .data(nodes)
            .join('text')
            .text((d: any) => d.label)
            .attr('font-size', 10)
            .attr('dx', 15)
            .attr('dy', 4);

        // Обновление позиций на каждом тике
        simulation.on('tick', () => {
            link
                .attr('x1', (d: any) => d.source.x)
                .attr('y1', (d: any) => d.source.y)
                .attr('x2', (d: any) => d.target.x)
                .attr('y2', (d: any) => d.target.y);

            node
                .attr('cx', (d: any) => d.x)
                .attr('cy', (d: any) => d.y);

            label
                .attr('x', (d: any) => d.x)
                .attr('y', (d: any) => d.y);
        });

        // Zoom behavior
        const zoom = d3.zoom()
            .scaleExtent([0.5, 5])
            .on('zoom', (event: any) => {
                g.attr('transform', event.transform);
            });

        svg.call(zoom as any);

        // Cleanup
        return () => {
            simulation.stop();
        };
    }, [nodes, links, width, height]);

    // Drag behavior для узлов
    function drag(simulation: any) {
        function dragstarted(event: any) {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            event.subject.fx = event.subject.x;
            event.subject.fy = event.subject.y;
        }

        function dragged(event: any) {
            event.subject.fx = event.x;
            event.subject.fy = event.y;
        }

        function dragended(event: any) {
            if (!event.active) simulation.alphaTarget(0);
            event.subject.fx = null;
            event.subject.fy = null;
        }

        return d3.drag()
            .on('start', dragstarted)
            .on('drag', dragged)
            .on('end', dragended);
    }

    return (
        <div className="network-graph-container">
            <svg ref={svgRef}></svg>
        </div>
    );
};
