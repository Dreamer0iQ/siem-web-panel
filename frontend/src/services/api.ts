import axios from 'axios';
import { authService } from './auth';

const API_BASE_URL = 'http://localhost:8080/api';

const apiClient = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

apiClient.interceptors.request.use(
    (config) => {
        // Проверяем валидность сессии перед каждым запросом
        if (!authService.isAuthenticated()) {
            authService.logout();
            window.location.href = '/login';
            return Promise.reject(new Error('Session expired'));
        }

        const authHeader = authService.getAuthHeader();
        if (authHeader) {
            config.headers.Authorization = authHeader;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

apiClient.interceptors.response.use(
    (response) => {
        authService.refreshSession();
        return response;
    },
    (error) => {
        if (error.response?.status === 401) {
            authService.logout();
            window.location.href = '/login';
        }
        return Promise.reject(error);
    }
);

export interface ActiveAgent {
    hostname: string;
    ip_address: string;
    last_activity: string;
    status: string;
}

export interface RecentLogin {
    timestamp: string;
    user: string;
    host: string;
    success: boolean;
    ip_address: string;
}

export interface ActiveHost {
    hostname: string;
    event_count: number;
    last_event: string;
    ip_address: string;
}

export interface EventTypeCount {
    type: string;
    count: number;
}

export interface SeverityCount {
    severity: string;
    count: number;
}

export interface UserActivity {
    username: string;
    event_count: number;
}

export interface ProcessActivity {
    process_name: string;
    event_count: number;
}

export interface TimelinePoint {
    hour: string;
    count: number;
}

export interface Event {
    id: string;
    timestamp: string;
    type: string;
    severity: string;
    host: string;
    user: string;
    process: string;
    description: string;
    ip_address: string;
    success?: boolean;
    details?: string;
}

export interface EventsResponse {
    events: Event[];
    total: number;
    page: number;
    limit: number;
    pages: number;
}

export interface EventFilters {
    page?: number;
    limit?: number;
    type?: string;
    severity?: string;
    user?: string;
    process?: string;
}

export const dashboardAPI = {
    getActiveAgents: () => apiClient.get<ActiveAgent[]>('/dashboard/agents'),
    getRecentLogins: () => apiClient.get<RecentLogin[]>('/dashboard/logins'),
    getActiveHosts: () => apiClient.get<ActiveHost[]>('/dashboard/hosts'),
    getEventsByType: () => apiClient.get<EventTypeCount[]>('/dashboard/events-by-type'),
    getEventsBySeverity: () => apiClient.get<SeverityCount[]>('/dashboard/events-by-severity'),
    getTopUsers: () => apiClient.get<UserActivity[]>('/dashboard/top-users'),
    getTopProcesses: () => apiClient.get<ProcessActivity[]>('/dashboard/top-processes'),
    getTimeline: () => apiClient.get<TimelinePoint[]>('/dashboard/timeline'),
};

export const eventsAPI = {
    getEvents: (filters: EventFilters = {}) => {
        const params = new URLSearchParams(
            Object.entries(filters)
                .filter(([_, value]) => value !== undefined && value !== '')
                .map(([key, value]) => [key, String(value)])
        );

        return apiClient.get<EventsResponse>(`/events?${params.toString()}`);
    },

    getEventDetail: (id: string) => apiClient.get<Event>(`/events/${id}`),

    exportEvents: (format: 'json' | 'csv', filters: EventFilters = {}) => {
        const params = new URLSearchParams(
            Object.entries({ format, ...filters })
                .filter(([_, value]) => value !== undefined && value !== '')
                .map(([key, value]) => [key, String(value)])
        );

        return apiClient.get(`/events/export?${params.toString()}`, {
            responseType: 'blob',
        });
    },
};

export default apiClient;
