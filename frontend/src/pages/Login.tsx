import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { authService } from '../services/auth';
import { TreeIcon } from '../components/common/Icons';
import './Login.scss';

export const Login = () => {
    const [isLogin, setIsLogin] = useState(true);
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [firstName, setFirstName] = useState('');
    const [lastName, setLastName] = useState('');
    const [email, setEmail] = useState('');

    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        if (!isLogin) {
            setError("Registration is currently disabled by administrator.");
            setLoading(false);
            return;
        }

        try {
            const success = await authService.login({ username, password });

            if (success) {
                navigate('/dashboard');
            } else {
                setError('Invalid username or password');
            }
        } catch (err) {
            console.error('Login error:', err);
            setError('Connection failed. Please check if the server is running.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="login-page">
            <div className="login-background-tree">
                <TreeIcon size={600} className="big-tree" />
            </div>

            <div className="login-container">
                <div className="auth-tabs">
                    <button
                        className={`tab-btn ${isLogin ? 'active' : ''}`}
                        onClick={() => setIsLogin(true)}
                    >
                        Login
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="minimal-form">
                    {error && <div className="error-message">{error}</div>}

                    {isLogin ? (
                        <>
                            <div className="form-group underline">
                                <label htmlFor="username">Username</label>
                                <input
                                    id="username"
                                    type="text"
                                    value={username}
                                    onChange={(e) => setUsername(e.target.value)}
                                    required
                                    autoFocus
                                />
                            </div>
                            <div className="form-group underline">
                                <label htmlFor="password">Password</label>
                                <input
                                    id="password"
                                    type="password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    required
                                />
                            </div>
                        </>
                    ) : (
                        <>
                            <div className="name-row">
                                <div className="form-group underline">
                                    <label>First Name*</label>
                                    <input type="text" value={firstName} onChange={e => setFirstName(e.target.value)} />
                                </div>
                                <div className="form-group underline">
                                    <label>Last Name*</label>
                                    <input type="text" value={lastName} onChange={e => setLastName(e.target.value)} />
                                </div>
                            </div>
                            <div className="form-group underline">
                                <label>Email address*</label>
                                <input type="email" value={email} onChange={e => setEmail(e.target.value)} />
                            </div>
                            <div className="form-group underline">
                                <label>Password*</label>
                                <input type="password" />
                            </div>
                        </>
                    )}

                    <div className="form-actions">
                        <button type="submit" className="submit-btn" disabled={loading}>
                            {loading ? 'Processing...' : (isLogin ? 'Sign In' : 'Register')}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};
