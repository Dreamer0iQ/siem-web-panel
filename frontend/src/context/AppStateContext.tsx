import { createContext, useContext, useReducer, useCallback } from 'react';
import type { ReactNode } from 'react';

export interface Event {
    id: string;
    timestamp: string;
    type: string;
    severity: 'critical' | 'error' | 'warning' | 'info';
    host: string;
    user: string;
    description: string;
    details?: Record<string, any>;
}

// Типы статистики
export interface Stats {
    totalEvents: number;
    eventsByType: Record<string, number>;
    eventsBySeverity: Record<string, number>;
    activeHosts: string[];
    recentAlerts: number;
}

// Глобальное состояние
interface AppState {
    events: Event[];
    filteredEvents: Event[];
    selectedEvent: Event | null;
    stats: Stats | null;
    filters: {
        search: string;
        type: string;
        severity: string;
        host: string;
        dateRange: [string, string] | null;
    };
    ui: {
        loading: boolean;
        error: string | null;
        sidebarOpen: boolean;
        theme: 'light' | 'dark';
    };
    realtime: {
        connected: boolean;
        lastUpdate: string | null;
    };
}

// Action types
type AppAction =
    | { type: 'SET_EVENTS'; payload: Event[] }
    | { type: 'ADD_EVENT'; payload: Event }
    | { type: 'SELECT_EVENT'; payload: Event | null }
    | { type: 'SET_STATS'; payload: Stats }
    | { type: 'SET_FILTER'; payload: { key: keyof AppState['filters']; value: any } }
    | { type: 'CLEAR_FILTERS' }
    | { type: 'SET_LOADING'; payload: boolean }
    | { type: 'SET_ERROR'; payload: string | null }
    | { type: 'TOGGLE_SIDEBAR' }
    | { type: 'SET_THEME'; payload: 'light' | 'dark' }
    | { type: 'SET_REALTIME_STATUS'; payload: boolean }
    | { type: 'UPDATE_LAST_UPDATE'; payload: string };

// Initial state
const initialState: AppState = {
    events: [],
    filteredEvents: [],
    selectedEvent: null,
    stats: null,
    filters: {
        search: '',
        type: '',
        severity: '',
        host: '',
        dateRange: null,
    },
    ui: {
        loading: false,
        error: null,
        sidebarOpen: true,
        theme: 'light',
    },
    realtime: {
        connected: false,
        lastUpdate: null,
    },
};

function appReducer(state: AppState, action: AppAction): AppState {
    switch (action.type) {
        case 'SET_EVENTS':
            return {
                ...state,
                events: action.payload,
                filteredEvents: applyFilters(action.payload, state.filters),
            };

        case 'ADD_EVENT':
            const newEvents = [action.payload, ...state.events];
            return {
                ...state,
                events: newEvents,
                filteredEvents: applyFilters(newEvents, state.filters),
            };

        case 'SELECT_EVENT':
            return { ...state, selectedEvent: action.payload };

        case 'SET_STATS':
            return { ...state, stats: action.payload };

        case 'SET_FILTER':
            const newFilters = {
                ...state.filters,
                [action.payload.key]: action.payload.value,
            };
            return {
                ...state,
                filters: newFilters,
                filteredEvents: applyFilters(state.events, newFilters),
            };

        case 'CLEAR_FILTERS':
            return {
                ...state,
                filters: initialState.filters,
                filteredEvents: state.events,
            };

        case 'SET_LOADING':
            return { ...state, ui: { ...state.ui, loading: action.payload } };

        case 'SET_ERROR':
            return { ...state, ui: { ...state.ui, error: action.payload } };

        case 'TOGGLE_SIDEBAR':
            return { ...state, ui: { ...state.ui, sidebarOpen: !state.ui.sidebarOpen } };

        case 'SET_THEME':
            return { ...state, ui: { ...state.ui, theme: action.payload } };

        case 'SET_REALTIME_STATUS':
            return { ...state, realtime: { ...state.realtime, connected: action.payload } };

        case 'UPDATE_LAST_UPDATE':
            return { ...state, realtime: { ...state.realtime, lastUpdate: action.payload } };

        default:
            return state;
    }
}

// Функция фильтрации событий
function applyFilters(events: Event[], filters: AppState['filters']): Event[] {
    return events.filter(event => {
        if (filters.search && !JSON.stringify(event).toLowerCase().includes(filters.search.toLowerCase())) {
            return false;
        }
        if (filters.type && event.type !== filters.type) {
            return false;
        }
        if (filters.severity && event.severity !== filters.severity) {
            return false;
        }
        if (filters.host && event.host !== filters.host) {
            return false;
        }
        if (filters.dateRange) {
            const eventDate = new Date(event.timestamp);
            const [start, end] = filters.dateRange;
            if (eventDate < new Date(start) || eventDate > new Date(end)) {
                return false;
            }
        }
        return true;
    });
}

const AppStateContext = createContext<AppState | undefined>(undefined);
const AppDispatchContext = createContext<React.Dispatch<AppAction> | undefined>(undefined);

export function AppStateProvider({ children }: { children: ReactNode }) {
    const [state, dispatch] = useReducer(appReducer, initialState);

    return (
        <AppStateContext.Provider value={state}>
            <AppDispatchContext.Provider value={dispatch}>
                {children}
            </AppDispatchContext.Provider>
        </AppStateContext.Provider>
    );
}

// Custom hooks
export function useAppState() {
    const context = useContext(AppStateContext);
    if (context === undefined) {
        throw new Error('useAppState must be used within AppStateProvider');
    }
    return context;
}

export function useAppDispatch() {
    const context = useContext(AppDispatchContext);
    if (context === undefined) {
        throw new Error('useAppDispatch must be used within AppStateProvider');
    }
    return context;
}

export function useAppActions() {
    const dispatch = useAppDispatch();

    return {
        setEvents: useCallback((events: Event[]) => {
            dispatch({ type: 'SET_EVENTS', payload: events });
        }, [dispatch]),

        addEvent: useCallback((event: Event) => {
            dispatch({ type: 'ADD_EVENT', payload: event });
        }, [dispatch]),

        selectEvent: useCallback((event: Event | null) => {
            dispatch({ type: 'SELECT_EVENT', payload: event });
        }, [dispatch]),

        setFilter: useCallback((key: keyof AppState['filters'], value: any) => {
            dispatch({ type: 'SET_FILTER', payload: { key, value } });
        }, [dispatch]),

        clearFilters: useCallback(() => {
            dispatch({ type: 'CLEAR_FILTERS' });
        }, [dispatch]),

        setLoading: useCallback((loading: boolean) => {
            dispatch({ type: 'SET_LOADING', payload: loading });
        }, [dispatch]),

        setError: useCallback((error: string | null) => {
            dispatch({ type: 'SET_ERROR', payload: error });
        }, [dispatch]),

        toggleSidebar: useCallback(() => {
            dispatch({ type: 'TOGGLE_SIDEBAR' });
        }, [dispatch]),

        setTheme: useCallback((theme: 'light' | 'dark') => {
            dispatch({ type: 'SET_THEME', payload: theme });
        }, [dispatch]),

        setRealtimeStatus: useCallback((connected: boolean) => {
            dispatch({ type: 'SET_REALTIME_STATUS', payload: connected });
        }, [dispatch]),
    };
}

export function useFilteredEvents() {
    const state = useAppState();
    return state.filteredEvents;
}

export function useCriticalEvents() {
    const state = useAppState();
    return state.filteredEvents.filter(e => e.severity === 'critical');
}

export function useEventsByHost() {
    const state = useAppState();
    return state.filteredEvents.reduce((acc, event) => {
        acc[event.host] = (acc[event.host] || 0) + 1;
        return acc;
    }, {} as Record<string, number>);
}
