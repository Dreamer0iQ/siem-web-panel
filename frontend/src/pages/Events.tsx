import { useEffect, useState } from 'react';
import { Header } from '../components/common/Header';
import { DownloadIcon, XIcon, SnowflakeIcon } from '../components/common/Icons';
import { eventsAPI, type Event, type EventFilters } from '../services/api';
import './Events.scss';

export const Events = () => {
    const [events, setEvents] = useState<Event[]>([]);
    const [total, setTotal] = useState(0);
    const [page, setPage] = useState(1);
    const [loading, setLoading] = useState(true);
    const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);

    const [filters, setFilters] = useState<EventFilters>({
        page: 1,
        limit: 20,
        type: '',
        severity: '',
        user: '',
        process: '',
    });

    const fetchEvents = async () => {
        setLoading(true);
        try {
            const response = await eventsAPI.getEvents(filters);
            setEvents(response.data.events);
            setTotal(response.data.total);
            setPage(response.data.page);
        } catch (error) {
            console.error('Failed to fetch events:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchEvents();
    }, [filters]);

    const handleFilterChange = (key: keyof EventFilters, value: string) => {
        setFilters(prev => ({ ...prev, [key]: value, page: 1 }));
    };

    const clearFilters = () => {
        setFilters({
            page: 1,
            limit: 20,
            type: '',
            severity: '',
            user: '',
            process: '',
        });
    };

    const handleExport = async (format: 'json' | 'csv') => {
        try {
            const response = await eventsAPI.exportEvents(format, filters);
            const url = window.URL.createObjectURL(new Blob([response.data]));
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', `events.${format}`);
            document.body.appendChild(link);
            link.click();
            link.remove();
        } catch (error) {
            console.error('Failed to export events:', error);
        }
    };

    const getSeverityClass = (severity: string) => {
        return `badge-${severity}`;
    };

    const totalPages = Math.ceil(total / (filters.limit || 20));

    return (
        <div className="events-page">
            <div className="events-background">
                <SnowflakeIcon size={80} className="snowflake snowflake-1" />
                <SnowflakeIcon size={60} className="snowflake snowflake-2" />
                <SnowflakeIcon size={90} className="snowflake snowflake-3" />
                <SnowflakeIcon size={70} className="snowflake snowflake-4" />
                <SnowflakeIcon size={100} className="snowflake snowflake-5" />
                <SnowflakeIcon size={55} className="snowflake snowflake-6" />
                <SnowflakeIcon size={85} className="snowflake snowflake-7" />
                <SnowflakeIcon size={75} className="snowflake snowflake-8" />
            </div>

            <Header />

            <div className="events-container">
                <div className="events-header">
                    <div>
                        <h2>Events Registry</h2>
                        <p>Total: {total} events</p>
                    </div>

                    <div className="export-buttons">
                        <button className="btn btn-secondary" onClick={() => handleExport('json')}>
                            <DownloadIcon size={18} />
                            Export JSON
                        </button>
                        <button className="btn btn-secondary" onClick={() => handleExport('csv')}>
                            <DownloadIcon size={18} />
                            Export CSV
                        </button>
                    </div>
                </div>

                <div className="filter-row">
                    <select
                        className="input"
                        value={filters.type}
                        onChange={(e) => handleFilterChange('type', e.target.value)}
                    >
                        <option value="">All Types</option>
                        <option value="authentication">Authentication</option>
                        <option value="file_access">File Access</option>
                        <option value="process_execution">Process Execution</option>
                        <option value="network_connection">Network Connection</option>
                        <option value="system_event">System Event</option>
                    </select>

                    <select
                        className="input"
                        value={filters.severity}
                        onChange={(e) => handleFilterChange('severity', e.target.value)}
                    >
                        <option value="">All Severities</option>
                        <option value="critical">Critical</option>
                        <option value="high">High</option>
                        <option value="medium">Medium</option>
                        <option value="low">Low</option>
                    </select>

                    <input
                        type="text"
                        className="input"
                        placeholder="Filter by user..."
                        value={filters.user}
                        onChange={(e) => handleFilterChange('user', e.target.value)}
                    />

                    <input
                        type="text"
                        className="input"
                        placeholder="Filter by process..."
                        value={filters.process}
                        onChange={(e) => handleFilterChange('process', e.target.value)}
                    />

                    <button className="btn btn-ghost" onClick={clearFilters} style={{"display" : "flex", "alignItems": "center"}}>
                        <XIcon size={18} />
                        Clear
                    </button>
                </div>

                <div className="events-table-container">
                    {loading ? (
                        <div className="loading-state">
                            <div className="skeleton" style={{ height: '50px', marginBottom: '8px' }} />
                            <div className="skeleton" style={{ height: '50px', marginBottom: '8px' }} />
                            <div className="skeleton" style={{ height: '50px', marginBottom: '8px' }} />
                        </div>
                    ) : (
                        <table className="table events-table">
                            <thead>
                                <tr>
                                    <th>Timestamp</th>
                                    <th>Type</th>
                                    <th>Severity</th>
                                    <th>Host</th>
                                    <th>User</th>
                                    <th>Description</th>
                                </tr>
                            </thead>
                            <tbody>
                                {events.map((event) => (
                                    <tr key={event.id} onClick={() => setSelectedEvent(event)}>
                                        <td>{new Date(event.timestamp).toLocaleString()}</td>
                                        <td>{event.type}</td>
                                        <td>
                                            <span className={`badge ${getSeverityClass(event.severity)}`}>
                                                {event.severity}
                                            </span>
                                        </td>
                                        <td>{event.host}</td>
                                        <td>{event.user}</td>
                                        <td className="description-cell">{event.description}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    )}
                </div>

                <div className="pagination">
                    <button
                        className="btn btn-secondary"
                        disabled={page === 1}
                        onClick={() => setFilters(prev => ({ ...prev, page: page - 1 }))}
                    >
                        Previous
                    </button>

                    <span className="pagination-info">
                        Page {page} of {totalPages}
                    </span>

                    <button
                        className="btn btn-secondary"
                        disabled={page === totalPages}
                        onClick={() => setFilters(prev => ({ ...prev, page: page + 1 }))}
                    >
                        Next
                    </button>
                </div>
            </div>

            {selectedEvent && (
                <div className="modal-overlay" onClick={() => setSelectedEvent(null)}>
                    <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                        <div className="modal-header">
                            <h3>Event Details</h3>
                            <button className="btn btn-ghost" onClick={() => setSelectedEvent(null)}>
                                <XIcon size={20} />
                            </button>
                        </div>

                        <div className="modal-body">
                            <div className="detail-row">
                                <span className="detail-label">ID:</span>
                                <span className="detail-value">{selectedEvent.id}</span>
                            </div>
                            <div className="detail-row">
                                <span className="detail-label">Timestamp:</span>
                                <span className="detail-value">{new Date(selectedEvent.timestamp).toLocaleString()}</span>
                            </div>
                            <div className="detail-row">
                                <span className="detail-label">Type:</span>
                                <span className="detail-value">{selectedEvent.type}</span>
                            </div>
                            <div className="detail-row">
                                <span className="detail-label">Severity:</span>
                                <span className={`badge ${getSeverityClass(selectedEvent.severity)}`}>
                                    {selectedEvent.severity}
                                </span>
                            </div>
                            <div className="detail-row">
                                <span className="detail-label">Host:</span>
                                <span className="detail-value">{selectedEvent.host}</span>
                            </div>
                            <div className="detail-row">
                                <span className="detail-label">User:</span>
                                <span className="detail-value">{selectedEvent.user}</span>
                            </div>
                            <div className="detail-row">
                                <span className="detail-label">Process:</span>
                                <span className="detail-value">{selectedEvent.process}</span>
                            </div>
                            <div className="detail-row">
                                <span className="detail-label">Description:</span>
                                <span className="detail-value">{selectedEvent.description}</span>
                            </div>
                            {selectedEvent.details && (
                                <div className="detail-row">
                                    <span className="detail-label">Details:</span>
                                    <span className="detail-value">{selectedEvent.details}</span>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            )
            }
        </div >
    );
};
