import { useEffect, useState } from 'react';
import { Header } from '../components/common/Header';
import { Card } from '../components/common/Card';
import {
    ActivityIcon,
    AlertIcon,
    ChartIcon,
    ClockIcon,
    ServerIcon,
    UserIcon,
    ShieldIcon,
} from '../components/common/Icons';
import { dashboardAPI, type ActiveAgent, type RecentLogin, type ActiveHost, type EventTypeCount, type SeverityCount, type UserActivity, type ProcessActivity, type TimelinePoint } from '../services/api';
import { Chart as ChartJS, ArcElement, Tooltip, Legend, CategoryScale, LinearScale, BarElement, LineElement, PointElement, Title } from 'chart.js';
import { Pie, Line } from 'react-chartjs-2';
import './Dashboard.scss';

ChartJS.register(ArcElement, Tooltip, Legend, CategoryScale, LinearScale, BarElement, LineElement, PointElement, Title);

export const Dashboard = () => {
    const [agents, setAgents] = useState<ActiveAgent[]>([]);
    const [logins, setLogins] = useState<RecentLogin[]>([]);
    const [hosts, setHosts] = useState<ActiveHost[]>([]);
    const [eventsByType, setEventsByType] = useState<EventTypeCount[]>([]);
    const [eventsBySeverity, setEventsBySeverity] = useState<SeverityCount[]>([]);
    const [topUsers, setTopUsers] = useState<UserActivity[]>([]);
    const [topProcesses, setTopProcesses] = useState<ProcessActivity[]>([]);
    const [timeline, setTimeline] = useState<TimelinePoint[]>([]);

    const fetchData = async () => {
        try {
            const [agentsRes, loginsRes, hostsRes, typeRes, severityRes, usersRes, processesRes, timelineRes] = await Promise.all([
                dashboardAPI.getActiveAgents(),
                dashboardAPI.getRecentLogins(),
                dashboardAPI.getActiveHosts(),
                dashboardAPI.getEventsByType(),
                dashboardAPI.getEventsBySeverity(),
                dashboardAPI.getTopUsers(),
                dashboardAPI.getTopProcesses(),
                dashboardAPI.getTimeline(),
            ]);

            setAgents(agentsRes?.data || []);
            setLogins(loginsRes?.data || []);
            setHosts(hostsRes?.data || []);
            setEventsByType(typeRes?.data || []);
            setEventsBySeverity(severityRes?.data || []);
            setTopUsers(usersRes?.data || []);
            setTopProcesses(processesRes?.data || []);
            setTimeline(timelineRes?.data || []);
        } catch (error) {
            console.error('Failed to fetch dashboard data:', error);
        }
    };

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 30000);
        return () => clearInterval(interval);
    }, []);

    const typeChartData = {
        labels: eventsByType?.map(e => e.type) || [],
        datasets: [{
            data: eventsByType?.map(e => e.count) || [],
            backgroundColor: ['#2C3E50', '#95A5A6', '#BDC3C7', '#E74C3C', '#F39C12'],
            borderWidth: 1,
            borderColor: '#FFFFFF'
        }],
    };

    const timelineChartData = {
        labels: timeline?.map(t => t.hour) || [],
        datasets: [{
            label: 'Event Volume',
            data: timeline?.map(t => t.count) || [],
            backgroundColor: 'rgba(44, 62, 80, 0.1)', // Light blue-gray fill
            borderColor: '#2C3E50', // Dark blue-gray line
            borderWidth: 2,
            fill: true,
            tension: 0.4,
            pointBackgroundColor: '#FFFFFF',
            pointBorderColor: '#2C3E50',
            pointBorderWidth: 2
        }],
    };

    const commonOptions = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: {
                position: 'right' as const,
                labels: {
                    boxWidth: 12,
                    color: '#2C3E50', // Dark text for legends
                    font: { family: '"Helvetica Neue", Helvetica, Arial, sans-serif' }
                }
            }
        }
    };

    const lineOptions = {
        ...commonOptions,
        plugins: { legend: { display: false } },
        scales: {
            x: {
                grid: { display: false, color: '#E2E8F0' },
                ticks: { color: '#7F8C8D' }
            },
            y: {
                border: { dash: [4, 4], color: '#E2E8F0' },
                grid: { color: 'rgba(0, 0, 0, 0.05)' },
                ticks: { color: '#7F8C8D' }
            }
        }
    };

    return (
        <div className="pt-dashboard">
            <Header />
            <div className="dashboard-content">
                <div className="dashboard-grid">

                    <Card title="Active Agents" icon={<ServerIcon />} className="col-span-3 widget-h-md">
                        <div className="agent-list">
                            <table className="pt-table">
                                <thead>
                                    <tr>
                                        <th>Hostname</th>
                                        <th>Status</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {agents?.map((agent, i) => (
                                        <tr key={i}>
                                            <td>
                                                <div style={{ fontWeight: 500 }}>{agent.hostname}</div>
                                                <div style={{ fontSize: '11px', color: '#7F8C8D' }}>{agent.ip_address}</div>
                                            </td>
                                            <td>
                                                <span className={`status-dot ${agent.status === 'active' ? 'active' : 'inactive'}`}></span>
                                                <span style={{ fontSize: '12px', marginLeft: '4px' }}>{agent.status}</span>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </Card>

                    <Card title="Recent Logins" icon={<ClockIcon />} className="col-span-5 widget-h-md">
                        <table className="pt-table">
                            <thead>
                                <tr>
                                    <th>Time</th>
                                    <th>User</th>
                                    <th>Status</th>
                                    <th>IP</th>
                                </tr>
                            </thead>
                            <tbody>
                                {logins?.slice(0, 6).map((l, i) => (
                                    <tr key={i}>
                                        <td>{new Date(l.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</td>
                                        <td>{l.user}</td>
                                        <td>
                                            <span className={l.success ? 'text-success' : 'text-danger'} style={{ fontWeight: 600, fontSize: '11px' }}>
                                                {l.success ? 'SUCCESS' : 'FAILURE'}
                                            </span>
                                        </td>
                                        <td>{l.ip_address}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </Card>

                    <Card title="Events by Type" icon={<ChartIcon />} className="col-span-4 widget-h-md">
                        <div className="chart-wrapper">
                            <Pie data={typeChartData} options={commonOptions} />
                        </div>
                    </Card>

                    <Card title="Distribution by Severity" icon={<AlertIcon />} className="col-span-3 widget-h-md">
                        <div className="chart-wrapper">
                            <Pie data={{
                                labels: eventsBySeverity?.map(e => e.severity) || [],
                                datasets: [{
                                    data: eventsBySeverity?.map(e => e.count) || [],
                                    backgroundColor: ['#E74C3C', '#F39C12', '#F1C40F', '#27ae60'],
                                    borderWidth: 1,
                                    borderColor: '#FFFFFF'
                                }]
                            }} options={commonOptions} />
                        </div>
                    </Card>

                    <Card title="Active Hosts" icon={<ShieldIcon />} className="col-span-3 widget-h-md">
                        <table className="pt-table">
                            <thead>
                                <tr>
                                    <th>Host</th>
                                    <th>Events</th>
                                </tr>
                            </thead>
                            <tbody>
                                {hosts?.slice(0, 6).map((h, i) => (
                                    <tr key={i}>
                                        <td>{h.hostname}</td>
                                        <td style={{ fontWeight: 'bold' }}>{h.event_count}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </Card>

                    <Card title="Top Processes" icon={<ActivityIcon />} className="col-span-3 widget-h-md">
                        <table className="pt-table">
                            <thead>
                                <tr>
                                    <th>Process</th>
                                    <th>Count</th>
                                </tr>
                            </thead>
                            <tbody>
                                {topProcesses?.slice(0, 6).map((p, i) => (
                                    <tr key={i}>
                                        <td>{p.process_name}</td>
                                        <td>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                                <div style={{
                                                    width: `${Math.min(100, (p.event_count / (topProcesses[0]?.event_count || 1)) * 100)}px`,
                                                    height: '6px',
                                                    background: '#2C3E50', // Use primary dark color
                                                    borderRadius: '0px'
                                                }} />
                                                <span style={{ fontSize: '11px' }}>{p.event_count}</span>
                                            </div>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </Card>

                    <Card title="Top Users" icon={<UserIcon />} className="col-span-3 widget-h-md">
                        <table className="pt-table">
                            <thead>
                                <tr>
                                    <th>User</th>
                                    <th>Activity</th>
                                </tr>
                            </thead>
                            <tbody>
                                {topUsers?.slice(0, 6).map((u, i) => (
                                    <tr key={i}>
                                        <td>{u.username}</td>
                                        <td style={{ fontWeight: 'bold' }}>{u.event_count}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </Card>

                    <Card title="Traffic Volume (24h)" icon={<ActivityIcon />} className="col-span-12 widget-h-lg">
                        <div className="chart-wrapper large">
                            <Line data={timelineChartData} options={lineOptions} />
                        </div>
                    </Card>

                </div>
            </div>
        </div>
    );
};
