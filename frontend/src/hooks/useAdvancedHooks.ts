import { useState, useEffect, useCallback, useRef } from 'react';

export function useDebounce<T>(value: T, delay: number, options?: {
    leading?: boolean;
    trailing?: boolean;
}): T {
    const [debouncedValue, setDebouncedValue] = useState<T>(value);
    const timerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const isFirstRender = useRef(true);
    const { leading = false, trailing = true } = options || {};

    useEffect(() => {
        if (leading && isFirstRender.current) {
            setDebouncedValue(value);
            isFirstRender.current = false;
            return;
        }

        if (trailing) {
            timerRef.current = setTimeout(() => {
                setDebouncedValue(value);
            }, delay);

            return () => {
                if (timerRef.current) {
                    clearTimeout(timerRef.current);
                }
            };
        }
    }, [value, delay, leading, trailing]);

    return debouncedValue;
}

export function useThrottle<T extends (...args: any[]) => any>(
    callback: T,
    delay: number
): T {
    const lastRan = useRef(Date.now());
    const timeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

    return useCallback(
        ((...args) => {
            const now = Date.now();

            if (now - lastRan.current >= delay) {
                callback(...args);
                lastRan.current = now;
            } else {
                if (timeoutRef.current) {
                    clearTimeout(timeoutRef.current);
                }

                timeoutRef.current = setTimeout(() => {
                    callback(...args);
                    lastRan.current = Date.now();
                }, delay - (now - lastRan.current));
            }
        }) as T,
        [callback, delay]
    );
}

export function usePrevious<T>(value: T): T | undefined {
    const ref = useRef<T | undefined>(undefined);

    useEffect(() => {
        ref.current = value;
    }, [value]);

    return ref.current;
}

export function useIntersectionObserver(
    ref: React.RefObject<Element>,
    options?: IntersectionObserverInit
): boolean {
    const [isIntersecting, setIsIntersecting] = useState(false);

    useEffect(() => {
        if (!ref.current) return;

        const observer = new IntersectionObserver(
            ([entry]) => {
                setIsIntersecting(entry.isIntersecting);
            },
            options
        );

        observer.observe(ref.current);

        return () => {
            observer.disconnect();
        };
    }, [ref, options]);

    return isIntersecting;
}

export function useLocalStorage<T>(key: string, initialValue: T): [T, (value: T) => void] {
    const [storedValue, setStoredValue] = useState<T>(() => {
        try {
            const item = window.localStorage.getItem(key);
            return item ? JSON.parse(item) : initialValue;
        } catch (error) {
            console.error(`Error loading ${key} from localStorage:`, error);
            return initialValue;
        }
    });

    const setValue = useCallback((value: T) => {
        try {
            setStoredValue(value);
            window.localStorage.setItem(key, JSON.stringify(value));
        } catch (error) {
            console.error(`Error saving ${key} to localStorage:`, error);
        }
    }, [key]);

    return [storedValue, setValue];
}

export function useAsync<T, E = Error>(
    asyncFunction: () => Promise<T>,
    dependencies: any[] = []
) {
    const [status, setStatus] = useState<'idle' | 'loading' | 'success' | 'error'>('idle');
    const [data, setData] = useState<T | null>(null);
    const [error, setError] = useState<E | null>(null);

    const execute = useCallback(async () => {
        setStatus('loading');
        setData(null);
        setError(null);

        try {
            const response = await asyncFunction();
            setData(response);
            setStatus('success');
            return response;
        } catch (err) {
            setError(err as E);
            setStatus('error');
            throw err;
        }
    }, dependencies);

    useEffect(() => {
        execute();
    }, [execute]);

    return { status, data, error, execute };
}
