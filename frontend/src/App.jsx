import React, { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, Link, useLocation } from 'react-router-dom';
import { fetchGraphQL, QUERIES, MUTATIONS } from './api/client';

const GlobalStyles = () => (
    <style>{`
    @import url('https://fonts.googleapis.com/css2?family=Share+Tech+Mono&family=Inter:wght@400;500;600;700;800&display=swap');
    
    :root {
      --bg: 9, 9, 11; /* Глубокий черный */
      --surface: 24, 24, 27; /* Графит */
      --primary: 255, 77, 0; /* #FF4D00 - Vermilion */
      --blood-red: 220, 38, 38; /* #DC2626 */
      --text: 250, 250, 250;
      --muted: 161, 161, 170;
      --border: 39, 39, 42;
      
      --font-tech: 'Share Tech Mono', monospace;
      --font-ui: 'Inter', sans-serif;
    }
    
    body {
      background-color: rgb(var(--bg));
      color: rgb(var(--text));
      font-family: var(--font-ui);
      margin: 0;
      overflow-x: hidden;
    }

    .tech-text { font-family: var(--font-tech); }

    /* Эффект матового стекла для карточек (Glassmorphism) */
    .glass-card {
      background: rgba(var(--surface), 0.75);
      backdrop-filter: blur(12px);
      -webkit-backdrop-filter: blur(12px);
      border: 1px solid rgba(255, 255, 255, 0.08);
    }
  `}</style>
);

// Живой процессуальный фон с переливающимися горизонталями карты
const TopographicBackground = () => {
    const canvasRef = React.useRef(null);

    React.useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        let animationFrameId;
        let time = 0;

        const handleResize = () => {
            canvas.width = window.innerWidth;
            canvas.height = window.innerHeight;
        };
        handleResize();
        window.addEventListener('resize', handleResize);

        const render = () => {
            time += 0.003;
            const { width, height } = canvas;
            ctx.clearRect(0, 0, width, height);

            // Sharp White с мягкой прозрачностью
            ctx.strokeStyle = 'rgba(255, 255, 255, 0.12)';
            ctx.lineWidth = 1.2;

            const numLines = 32;
            const stepY = (height + 300) / numLines;

            for (let i = -5; i < numLines + 5; i++) {
                const baseY = i * stepY;
                ctx.beginPath();

                const stepX = 16;
                for (let x = -50; x <= width + 50; x += stepX) {
                    // Взаимосвязанные синусоидальные гармоники: соседние линии сдвигаются со смещением фазы (i)
                    const wave1 = Math.sin(x * 0.003 + time + i * 0.22) * 48;
                    const wave2 = Math.cos(x * 0.006 - time * 0.7 + i * 0.12) * 28;
                    const wave3 = Math.sin((x + baseY) * 0.002 + time * 1.1) * 32;

                    const y = baseY + wave1 + wave2 + wave3;

                    if (x === -50) {
                        ctx.moveTo(x, y);
                    } else {
                        ctx.lineTo(x, y);
                    }
                }
                ctx.stroke();
            }

            animationFrameId = requestAnimationFrame(render);
        };

        render();

        return () => {
            window.removeEventListener('resize', handleResize);
            cancelAnimationFrame(animationFrameId);
        };
    }, []);

    return (
        <canvas
            ref={canvasRef}
            className="fixed inset-0 pointer-events-none z-0 w-full h-full"
        />
    );
};

// Компонент логотипа без сторонних оверлеев и анимаций
const KolibriLogo = ({ className = "w-32 h-32", src = "/kolibri.png" }) => (
    <div className={`relative inline-flex items-center justify-center ${className}`}>
        <img
            src={src}
            alt="Kolibri Squad Logo"
            className="w-full h-full object-contain filter drop-shadow-lg"
            onError={(e) => {
                // Заглушка, если файл еще не положен в public/kolibri.png
                e.target.onerror = null;
                e.target.style.display = 'none';
                e.target.nextElementSibling.style.display = 'block';
            }}
        />

        {/* Резервный силуэт на случай отсутствия файла в public/ */}
        <svg
            viewBox="0 0 100 100"
            className="w-full h-full drop-shadow-lg hidden"
            xmlns="http://www.w3.org/2000/svg"
        >
            <path
                d="M50 20 C 35 20, 25 35, 25 50 C 25 65, 30 75, 45 80 L 45 85 L 50 80 C 65 85, 80 75, 80 50 C 80 30, 65 20, 50 20 Z M 25 45 Q 10 50, 5 70 Q 15 60, 25 55 Z M 55 80 L 55 90 L 60 90 L 60 80 Z"
                fill="#FFFFFF"
            />
        </svg>
    </div>
);

const Icons = {
    Steam: () => <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor"><path d="M11.979 0C5.364 0 0 5.362 0 11.979c0 4.673 2.68 8.718 6.592 10.742l3.435-4.945c-.244-.229-.404-.546-.404-.897 0-.693.565-1.259 1.259-1.259.693 0 1.258.566 1.258 1.259 0 .193-.047.375-.126.536l3.66 2.605c.801-.131 1.62-.258 2.378-.475 4.398-1.266 7.636-5.32 7.636-10.02C25.688 5.362 20.323 0 13.708 0h-1.729Zm5.358 15.65c-1.39 0-2.52.883-2.906 2.093l-3.327-2.37c.026-.145.045-.292.045-.445 0-1.785-1.448-3.233-3.233-3.233-1.785 0-3.234 1.448-3.234 3.233 0 1.545 1.084 2.844 2.532 3.167l-3.332 4.793A11.933 11.933 0 0 1 0 11.979c0-6.615 5.364-11.979 11.979-11.979 6.616 0 11.98 5.364 11.98 11.979 0 5.176-3.272 9.57-7.851 11.238-.135-2.073-1.854-3.72-3.951-3.72a3.95 3.95 0 0 0-3.834 2.996l3.15 2.242A11.897 11.897 0 0 0 11.98 23.96c6.616 0 11.98-5.365 11.98-11.981 0-6.616-5.364-11.98-11.98-11.98Z"/></svg>,
    Dashboard: () => <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="3" y="3" width="7" height="9"/><rect x="14" y="3" width="7" height="5"/><rect x="14" y="12" width="7" height="9"/><rect x="3" y="16" width="7" height="5"/></svg>,
    Clans: () => <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>,
    Events: () => <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>,
    LogOut: () => <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
};

const LoginPage = () => {
    const handleSteamRedirect = () => {
        const returnUrl = window.location.origin;
        const params = new URLSearchParams({
            'openid.ns': 'http://specs.openid.net/auth/2.0',
            'openid.mode': 'checkid_setup',
            'openid.return_to': returnUrl,
            'openid.realm': returnUrl,
            'openid.identity': 'http://specs.openid.net/auth/2.0/identifier_select',
            'openid.claimed_id': 'http://specs.openid.net/auth/2.0/identifier_select'
        });
        window.location.href = `https://steamcommunity.com/openid/login?${params.toString()}`;
    };

    return (
        <div className="min-h-screen flex flex-col items-center justify-center relative overflow-hidden bg-[rgb(var(--bg))]">
            <TopographicBackground />
            <div className="z-10 text-center max-w-2xl px-6 flex flex-col items-center">

                <div className="mb-8 flex flex-col items-center glass-card p-10 rounded-2xl shadow-2xl relative overflow-hidden group">
                    {/* Свечение на фоне карточки */}
                    <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-48 h-48 bg-[rgb(var(--primary))] opacity-[0.1] rounded-full blur-3xl group-hover:opacity-[0.2] transition-opacity duration-500"></div>

                    <h1 className="text-5xl md:text-7xl font-black tech-text tracking-widest text-white mb-2 uppercase">
                        KOLIBRI SQUAD
                    </h1>
                    <p className="text-[rgb(var(--primary))] tech-text tracking-[0.3em] uppercase text-sm font-bold">
                        Tactical Network Authorization
                    </p>
                </div>

                <button
                    onClick={handleSteamRedirect}
                    className="inline-flex items-center gap-4 px-10 py-5 bg-[rgb(var(--primary))] text-white font-bold tracking-widest uppercase transition-all duration-300 rounded-sm shadow-[0_0_20px_rgba(255,77,0,0.3)] hover:shadow-[0_0_30px_rgba(255,77,0,0.6)] hover:bg-white hover:text-[rgb(var(--primary))] hover:scale-105"
                >
                    <Icons.Steam />
                    <span className="tech-text">Authenticate via Steam</span>
                </button>
            </div>
        </div>
    );
};

const DashboardPage = ({ userId }) => {
    const [stats, setStats] = useState(null);
    const [error, setError] = useState(null);

    useEffect(() => {
        const loadStats = async () => {
            try {
                const data = await fetchGraphQL(QUERIES.GET_STATS, { userId });
                setStats(data.getPlayerStats);
            } catch (err) {
                setError(err.message);
            }
        };
        if (userId) loadStats();
    }, [userId]);

    if (error) return <div className="p-10 tech-text text-[rgb(var(--blood-red))]">ERROR: {error}</div>;
    if (!stats) return <div className="p-10 tech-text text-[rgb(var(--primary))] animate-pulse">DOWNLOADING COMBAT METRICS...</div>;

    const kdRatio = stats.deaths > 0 ? (stats.kills / stats.deaths).toFixed(2) : stats.kills;

    return (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 animate-in fade-in duration-500 relative z-10">

            {/* ELO Card */}
            <div className="col-span-1 md:col-span-2 glass-card rounded-xl p-8 relative overflow-hidden group">
                <div className="absolute -left-10 -bottom-10 w-48 h-48 bg-[rgb(var(--primary))] opacity-[0.05] rounded-full blur-3xl group-hover:opacity-[0.1] transition-opacity duration-500"></div>

                <h3 className="tech-text text-[rgb(var(--muted))] mb-8 tracking-widest text-sm uppercase flex items-center gap-3">
                    <span className="w-2 h-2 bg-[rgb(var(--primary))] rounded-full block animate-pulse"></span>
                    COMBAT ELO RATING
                </h3>

                <div className="flex items-baseline gap-4 mb-4 relative z-10">
                    <span className="text-7xl font-black text-white">{stats.eloRating}</span>
                    <span className="text-xl text-[rgb(var(--primary))] tech-text font-bold">PTS</span>
                </div>

                <div className="w-full bg-black/50 h-2 mt-8 relative rounded-full overflow-hidden border border-white/5">
                    <div className="absolute top-0 left-0 h-full bg-[rgb(var(--primary))] shadow-[0_0_10px_rgba(255,77,0,1)] rounded-full transition-all duration-1000" style={{width: `${(stats.eloRating % 100)}%`}}></div>
                </div>
                <div className="flex justify-between mt-2 tech-text text-[rgb(var(--muted))] text-xs">
                    <span>CURRENT TIER: CLASSIFIED</span>
                    <span>NEXT RANK PROGRESS</span>
                </div>
            </div>

            {/* K/D Card */}
            <div className="glass-card rounded-xl p-8 flex flex-col justify-between">
                <div>
                    <h3 className="tech-text text-[rgb(var(--muted))] mb-6 tracking-widest text-sm uppercase">LETHALITY INDEX</h3>
                    <div className="space-y-4 tech-text text-lg">
                        <div className="flex justify-between border-b border-white/10 pb-2">
                            <span className="text-[rgb(var(--muted))]">KILLS</span>
                            <span className="text-white font-bold">{stats.kills}</span>
                        </div>
                        <div className="flex justify-between border-b border-white/10 pb-2">
                            <span className="text-[rgb(var(--muted))]">DEATHS</span>
                            <span className="text-white font-bold">{stats.deaths}</span>
                        </div>
                    </div>
                </div>
                <div className="mt-8 pt-4 border-t border-[rgb(var(--primary))]/30 flex justify-between items-baseline">
                    <span className="tech-text text-[rgb(var(--primary))] font-bold">K/D RATIO</span>
                    <span className="text-4xl font-black text-white">{kdRatio}</span>
                </div>
            </div>

            {/* Stats Footer */}
            <div className="col-span-1 md:col-span-3 glass-card rounded-xl p-8 grid grid-cols-1 md:grid-cols-3 gap-8">
                <div className="border-l-2 border-[rgb(var(--primary))] pl-6">
                    <p className="tech-text text-[rgb(var(--muted))] text-xs mb-2 tracking-widest">DEPLOYMENT ROLE</p>
                    <p className="text-3xl font-black tech-text uppercase text-white">{stats.favoriteRole || "UNKNOWN"}</p>
                </div>
                <div className="border-l-2 border-[rgb(var(--primary))] pl-6">
                    <p className="tech-text text-[rgb(var(--muted))] text-xs mb-2 tracking-widest">COMBAT HOURS</p>
                    <p className="text-3xl font-black tech-text text-white">{stats.totalPlaytimeHours} <span className="text-lg text-[rgb(var(--muted))] font-normal">HRS</span></p>
                </div>
                <div className="border-l-2 border-[rgb(var(--primary))] pl-6">
                    <p className="tech-text text-[rgb(var(--muted))] text-xs mb-2 tracking-widest">ALLIES REVIVED</p>
                    <p className="text-3xl font-black tech-text text-[rgb(var(--primary))]">{stats.revives}</p>
                </div>
            </div>
        </div>
    );
};

const ClansPage = () => (
    <div className="glass-card rounded-xl p-12 text-center py-24 relative z-10">
        <Icons.Clans />
        <h2 className="tech-text text-3xl font-bold text-white mt-4 mb-2 uppercase">Clan Network</h2>
        <p className="text-[rgb(var(--primary))] tech-text tracking-widest">AWAITING BACKEND SYNCHRONIZATION...</p>
    </div>
);

const EventsPage = () => (
    <div className="glass-card rounded-xl p-12 text-center py-24 relative z-10">
        <Icons.Events />
        <h2 className="tech-text text-3xl font-bold text-white mt-4 mb-2 uppercase">Combat Operations</h2>
        <p className="text-[rgb(var(--primary))] tech-text tracking-widest">MATCHMAKING PROTOCOLS OFFLINE</p>
    </div>
);

const MainLayout = ({ children, steamId, onLogout }) => {
    const location = useLocation();

    const navItems = [
        { path: '/', label: 'DASHBOARD', icon: Icons.Dashboard },
        { path: '/clans', label: 'CLANS', icon: Icons.Clans },
        { path: '/events', label: 'OPERATIONS', icon: Icons.Events },
    ];

    return (
        <div className="min-h-screen flex flex-col bg-[rgb(var(--bg))] relative">
            <TopographicBackground />

            {/* Верхняя панель управления (Top Navigation) */}
            <header className="border-b border-white/10 glass-card z-20 sticky top-0 backdrop-blur-md">
                <div className="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between gap-4">

                    {/* Левая часть: Логотип и заголовок */}
                    <div className="flex items-center gap-4">
                        <KolibriLogo className="w-12 h-12 flex-shrink-0" />
                        <div>
                            <h1 className="text-xl font-black tech-text tracking-widest text-white uppercase leading-none">KOLIBRI</h1>
                            <p className="text-[rgb(var(--primary))] text-[10px] tech-text font-bold tracking-wider mt-0.5">SQUAD NETWORK</p>
                        </div>
                    </div>

                    {/* Навигация по центру */}
                    <nav className="hidden md:flex items-center gap-2">
                        {navItems.map(item => {
                            const isActive = location.pathname === item.path;
                            return (
                                <Link
                                    key={item.path}
                                    to={item.path}
                                    className={`flex items-center gap-2.5 px-4 py-2.5 rounded-lg transition-all duration-300 tech-text text-xs tracking-widest uppercase font-bold
                    ${isActive
                                        ? 'bg-[rgb(var(--primary))] text-white shadow-[0_0_15px_rgba(255,77,0,0.4)]'
                                        : 'text-[rgb(var(--muted))] hover:text-white hover:bg-white/5'
                                    }`}
                                >
                                    <item.icon /> {item.label}
                                </Link>
                            );
                        })}
                    </nav>

                    {/* Правая часть: Данные оператора и кнопка Disconnect */}
                    <div className="flex items-center gap-6">
                        <div className="text-right hidden sm:block">
                            <div className="tech-text text-[rgb(var(--muted))] text-[10px] tracking-widest">OPERATOR ID</div>
                            <div className="font-bold text-white tech-text text-xs">{steamId}</div>
                        </div>

                        <button
                            onClick={onLogout}
                            className="flex items-center gap-2 px-3.5 py-2 rounded-lg border border-[rgb(var(--blood-red))]/40 text-[rgb(var(--blood-red))] hover:bg-[rgb(var(--blood-red))] hover:text-white transition-all tech-text text-xs font-bold tracking-widest uppercase"
                            title="DISCONNECT"
                        >
                            <Icons.LogOut />
                            <span className="hidden lg:inline">DISCONNECT</span>
                        </button>
                    </div>

                </div>

                {/* Мобильное меню навигации */}
                <nav className="flex md:hidden border-t border-white/5 px-4 py-2 justify-around bg-black/40">
                    {navItems.map(item => {
                        const isActive = location.pathname === item.path;
                        return (
                            <Link
                                key={item.path}
                                to={item.path}
                                className={`flex items-center gap-2 px-3 py-1.5 rounded-md transition-all tech-text text-xs tracking-widest font-bold
                  ${isActive
                                    ? 'text-[rgb(var(--primary))]'
                                    : 'text-[rgb(var(--muted))]'
                                }`}
                            >
                                <item.icon /> {item.label}
                            </Link>
                        );
                    })}
                </nav>
            </header>

            {/* Основной контент */}
            <main className="flex-1 max-w-7xl w-full mx-auto p-6 md:p-10 relative overflow-y-auto z-10">
                {children}
            </main>
        </div>
    );
};

const ProtectedRoute = ({ children, isAuthenticated }) => {
    if (!isAuthenticated) return <Navigate to="/login" replace />;
    return children;
};

export default function App() {
    const [authState, setAuthState] = useState({ userId: null, steamId: null, isLoading: true });

    useEffect(() => {
        const urlParams = new URLSearchParams(window.location.search);
        const mode = urlParams.get('openid.mode');

        if (mode === 'id_res') {
            const processSteamRedirect = async () => {
                try {
                    const paramsObject = Object.fromEntries(urlParams.entries());
                    const data = await fetchGraphQL(MUTATIONS.LOGIN_STEAM, { paramsJson: JSON.stringify(paramsObject) });

                    const { userId, steamId, token } = data.loginWithSteam;
                    localStorage.setItem('squad_jwt', token);
                    localStorage.setItem('squad_user_id', userId);
                    localStorage.setItem('squad_steam_id', steamId);

                    window.history.replaceState({}, document.title, window.location.pathname);
                    setAuthState({ userId, steamId, isLoading: false });
                } catch (error) {
                    console.error("Auth Error:", error);
                    setAuthState({ userId: null, steamId: null, isLoading: false });
                }
            };
            processSteamRedirect();
        } else {
            const savedUserId = localStorage.getItem('squad_user_id');
            const savedSteamId = localStorage.getItem('squad_steam_id');
            setAuthState({ userId: savedUserId, steamId: savedSteamId, isLoading: false });
        }
    }, []);

    const handleLogout = () => {
        localStorage.clear();
        setAuthState({ userId: null, steamId: null, isLoading: false });
    };

    if (authState.isLoading) {
        return <div className="min-h-screen bg-[rgb(var(--bg))] flex items-center justify-center"><div className="tech-text text-[rgb(var(--primary))] font-bold text-xl animate-pulse tracking-widest">CONNECTING TO KOLIBRI NETWORK...</div></div>;
    }

    return (
        <BrowserRouter>
            <GlobalStyles />
            <Routes>
                <Route path="/login" element={authState.userId ? <Navigate to="/" replace /> : <LoginPage />} />

                <Route path="/" element={<ProtectedRoute isAuthenticated={!!authState.userId}><MainLayout steamId={authState.steamId} onLogout={handleLogout}><DashboardPage userId={authState.userId} /></MainLayout></ProtectedRoute>} />
                <Route path="/clans" element={<ProtectedRoute isAuthenticated={!!authState.userId}><MainLayout steamId={authState.steamId} onLogout={handleLogout}><ClansPage /></MainLayout></ProtectedRoute>} />
                <Route path="/events" element={<ProtectedRoute isAuthenticated={!!authState.userId}><MainLayout steamId={authState.steamId} onLogout={handleLogout}><EventsPage /></MainLayout></ProtectedRoute>} />

                <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
        </BrowserRouter>
    );
}