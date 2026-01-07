import { useEffect, useRef } from 'react';

export interface HeatmapData {
    hour: number;
    day: number;
    value: number;
}

interface Props {
    data: HeatmapData[];
    width?: number;
    height?: number;
}

export const ActivityHeatmap: React.FC<Props> = ({ 
    data, 
    width = 800, 
    height = 300 
}) => {
    const canvasRef = useRef<HTMLCanvasElement>(null);

    useEffect(() => {
        if (!canvasRef.current || data.length === 0) return;

        const canvas = canvasRef.current;
        const ctx = canvas.getContext('2d');
        if (!ctx) return;

        // Настройка размеров
        canvas.width = width;
        canvas.height = height;

        // Очистка
        ctx.clearRect(0, 0, width, height);

        // Параметры сетки
        const cellWidth = width / 24; // 24 часа
        const cellHeight = height / 7; // 7 дней

        // Находим максимальное значение для нормализации
        const maxValue = Math.max(...data.map(d => d.value));

        // Цветовая схема (от холодного к горячему)
        const getColor = (value: number): string => {
            const intensity = value / maxValue;
            
            if (intensity < 0.2) return `rgba(52, 152, 219, ${intensity * 2})`;
            if (intensity < 0.4) return `rgba(46, 204, 113, ${intensity})`;
            if (intensity < 0.6) return `rgba(241, 196, 15, ${intensity})`;
            if (intensity < 0.8) return `rgba(230, 126, 34, ${intensity})`;
            return `rgba(231, 76, 60, ${intensity})`;
        };

        // Отрисовка ячеек
        data.forEach(({ hour, day, value }) => {
            const x = hour * cellWidth;
            const y = day * cellHeight;

            ctx.fillStyle = getColor(value);
            ctx.fillRect(x, y, cellWidth - 1, cellHeight - 1);

            // Добавление текста для высоких значений
            if (value / maxValue > 0.5) {
                ctx.fillStyle = '#fff';
                ctx.font = '10px Arial';
                ctx.textAlign = 'center';
                ctx.textBaseline = 'middle';
                ctx.fillText(
                    value.toString(),
                    x + cellWidth / 2,
                    y + cellHeight / 2
                );
            }
        });

        // Отрисовка меток дней недели
        const days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
        ctx.fillStyle = '#333';
        ctx.font = '12px Arial';
        ctx.textAlign = 'right';
        days.forEach((day, i) => {
            ctx.fillText(day, width - 5, i * cellHeight + cellHeight / 2);
        });

        // Отрисовка меток часов
        ctx.textAlign = 'center';
        for (let hour = 0; hour < 24; hour += 3) {
            ctx.fillText(
                `${hour}:00`,
                hour * cellWidth + cellWidth / 2,
                height - 5
            );
        }

        // Добавление легенды
        const legendWidth = 200;
        const legendHeight = 20;
        const legendX = (width - legendWidth) / 2;
        const legendY = 10;

        const gradient = ctx.createLinearGradient(legendX, legendY, legendX + legendWidth, legendY);
        gradient.addColorStop(0, 'rgba(52, 152, 219, 0.3)');
        gradient.addColorStop(0.25, 'rgba(46, 204, 113, 0.5)');
        gradient.addColorStop(0.5, 'rgba(241, 196, 15, 0.7)');
        gradient.addColorStop(0.75, 'rgba(230, 126, 34, 0.9)');
        gradient.addColorStop(1, 'rgba(231, 76, 60, 1)');

        ctx.fillStyle = gradient;
        ctx.fillRect(legendX, legendY, legendWidth, legendHeight);

        ctx.strokeStyle = '#333';
        ctx.strokeRect(legendX, legendY, legendWidth, legendHeight);

        ctx.fillStyle = '#333';
        ctx.font = '10px Arial';
        ctx.fillText('Low', legendX - 20, legendY + legendHeight / 2);
        ctx.fillText('High', legendX + legendWidth + 20, legendY + legendHeight / 2);

    }, [data, width, height]);

    return (
        <div className="heatmap-container">
            <canvas ref={canvasRef}></canvas>
        </div>
    );
};
