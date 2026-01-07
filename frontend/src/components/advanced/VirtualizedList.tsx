import { useEffect, useRef, useState, useCallback } from 'react';

interface VirtualizedListProps<T> {
    items: T[];
    itemHeight: number;
    containerHeight: number;
    renderItem: (item: T, index: number) => React.ReactNode;
    overscan?: number;
}

export function VirtualizedList<T>({
    items,
    itemHeight,
    containerHeight,
    renderItem,
    overscan = 3,
}: VirtualizedListProps<T>) {
    const [scrollTop, setScrollTop] = useState(0);
    const containerRef = useRef<HTMLDivElement>(null);

    const totalHeight = items.length * itemHeight;
    // const visibleCount = Math.ceil(containerHeight / itemHeight);

    // Вычисление видимого диапазона с overscan
    const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan);
    const endIndex = Math.min(
        items.length - 1,
        Math.floor((scrollTop + containerHeight) / itemHeight) + overscan
    );

    const visibleItems = items.slice(startIndex, endIndex + 1);

    // Обработка скролла с throttling через rAF
    const handleScroll = useCallback(() => {
        if (containerRef.current) {
            requestAnimationFrame(() => {
                setScrollTop(containerRef.current!.scrollTop);
            });
        }
    }, []);

    useEffect(() => {
        const container = containerRef.current;
        if (!container) return;

        container.addEventListener('scroll', handleScroll, { passive: true });
        return () => container.removeEventListener('scroll', handleScroll);
    }, [handleScroll]);

    return (
        <div
            ref={containerRef}
            style={{
                height: containerHeight,
                overflow: 'auto',
                position: 'relative',
            }}
        >
            <div style={{ height: totalHeight, position: 'relative' }}>
                {visibleItems.map((item, index) => {
                    const actualIndex = startIndex + index;
                    return (
                        <div
                            key={actualIndex}
                            style={{
                                position: 'absolute',
                                top: actualIndex * itemHeight,
                                height: itemHeight,
                                width: '100%',
                            }}
                        >
                            {renderItem(item, actualIndex)}
                        </div>
                    );
                })}
            </div>
        </div>
    );
}

export function useInfiniteScroll(
    callback: () => void,
    options: {
        threshold?: number;
        enabled?: boolean;
    } = {}
) {
    const { threshold = 0.8, enabled = true } = options;
    const observerRef = useRef<IntersectionObserver | null>(null);
    const targetRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (!enabled || !targetRef.current) return;

        observerRef.current = new IntersectionObserver(
            (entries) => {
                const [entry] = entries;
                if (entry.isIntersecting) {
                    callback();
                }
            },
            { threshold }
        );

        observerRef.current.observe(targetRef.current);

        return () => {
            if (observerRef.current) {
                observerRef.current.disconnect();
            }
        };
    }, [callback, threshold, enabled]);

    return targetRef;
}
