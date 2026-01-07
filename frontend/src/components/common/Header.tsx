import { useNavigate, useLocation } from 'react-router-dom';
import { authService } from '../../services/auth';
import { LogoutIcon } from './Icons';
import './Header.scss';

export const Header = () => {
    const navigate = useNavigate();
    const location = useLocation();

    const handleLogout = () => {
        authService.logout();
        navigate('/login');
    };

    const isActive = (path: string) => location.pathname === path;

    return (
        <header className="pt-nav">
            <nav className="nav-links">
                <button
                    className={`nav-link ${isActive('/dashboard') ? 'active' : ''}`}
                    onClick={() => navigate('/dashboard')}
                >
                    Dashboards
                </button>
                <button
                    className={`nav-link ${isActive('/events') ? 'active' : ''}`}
                    onClick={() => navigate('/events')}
                >
                    Events
                </button>
            </nav>

            <div className="nav-right" style={{ marginLeft: 'auto' }}>
                <button className="nav-link logout" onClick={handleLogout}>
                    <LogoutIcon size={18} />
                </button>
            </div>
        </header>
    );
};
