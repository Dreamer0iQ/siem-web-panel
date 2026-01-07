interface IconProps {
    size?: number;
    color?: string;
    className?: string;
}

export const UserIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2M12 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8z"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export const ServerIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <rect x="2" y="2" width="20" height="8" rx="2" stroke={color} strokeWidth="2" />
        <rect x="2" y="14" width="20" height="8" rx="2" stroke={color} strokeWidth="2" />
        <path d="M6 6h.01M6 18h.01" stroke={color} strokeWidth="2" strokeLinecap="round" />
    </svg>
);

export const ActivityIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M22 12h-4l-3 9L9 3l-3 9H2"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export const ShieldIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export const AlertIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0zM12 9v4M12 17h.01"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export const SearchIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <circle cx="11" cy="11" r="8" stroke={color} strokeWidth="2" />
        <path d="m21 21-4.35-4.35" stroke={color} strokeWidth="2" strokeLinecap="round" />
    </svg>
);

export const FilterIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M22 3H2l8 9.46V19l4 2v-8.54L22 3z"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export const DownloadIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export const LogoutIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export const ChartIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M3 3v18h18M7 16l4-4 4 4 6-6"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export const ClockIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <circle cx="12" cy="12" r="10" stroke={color} strokeWidth="2" />
        <path d="M12 6v6l4 2" stroke={color} strokeWidth="2" strokeLinecap="round" />
    </svg>
);

export const CheckIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M20 6L9 17l-5-5"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export const XIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M18 6L6 18M6 6l12 12"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export const MenuIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M3 12h18M3 6h18M3 18h18"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export const SnowflakeIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M12 2v20M17 7l-5 5-5-5M17 17l-5-5-5 5M2 12h20M7 7l5 5 5-5M7 17l5-5 5 5"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

export const TreeIcon = ({ size = 24, color = 'currentColor', className = '' }: IconProps) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className={className}>
        <path
            d="M12 3l-8 14h16l-8-14zM12 17v4"
            stroke={color}
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
        <path
            d="M8 12l4-5 4 5"
            stroke={color}
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);
