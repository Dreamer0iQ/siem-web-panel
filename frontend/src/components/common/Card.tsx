import type { ReactNode } from 'react';
import './Card.scss';

interface CardProps {
    title?: string;
    icon?: ReactNode;
    children: ReactNode;
    className?: string;
    loading?: boolean;
}

export const Card = ({ title, icon, children, className = '', loading = false }: CardProps) => {
    return (
        <div className={`card ${className}`}>
            {title && (
                <div className="card-header">
                    {icon && <div className="card-icon">{icon}</div>}
                    <h3 className="card-title">{title}</h3>
                </div>
            )}
            <div className="card-content">
                {loading ? (
                    <div className="card-loading">
                        <div className="skeleton" style={{ height: '40px', marginBottom: '12px' }} />
                        <div className="skeleton" style={{ height: '40px', marginBottom: '12px' }} />
                        <div className="skeleton" style={{ height: '40px' }} />
                    </div>
                ) : (
                    children
                )}
            </div>
        </div>
    );
};
