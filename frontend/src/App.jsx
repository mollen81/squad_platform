import React, { useState, useEffect } from 'react';
import { fetchGraphQL, QUERIES, MUTATIONS } from './api/client';

const GlobalStyles = () => (
    <style>{`
    @import url('https://fonts.googleapis.com/css2?family=Share+Tech+Mono&family=Inter:wght@400;500;600;700&display=swap');
    
    :root {
      /* Tactical Terminal Palette */
      --bg: 9, 9, 11; /* Очень темный фон */
      --surface: 24, 24, 27; /* Темно-серые карточки */
      --primary: 16, 185, 129; /* Emerald Green (Терминальный зеленый) */
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
      background-image: 
        linear-gradient(rgba(16, 185, 129, 0.03) 1px, transparent 1px),
        linear-gradient(90deg, rgba(16, 185, 129, 0.03) 1px, transparent 1px);
      background-size: 40px 40px;
    }

    .tech-text { font-family: var(--font-tech); }
    
    .glitch-hover:hover {
      animation: glitch-skew 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94) both infinite;
      color: rgb(var(--primary));
      border-color: rgb(var(--primary));
    }

    @keyframes glitch-skew {
      0% { transform: skew(0deg); }
      20% { transform: skew(-1deg); }
      40% { transform: skew(1deg); }
      60% { transform: skew(-0.5deg); }
      80% { transform: skew(0.5deg); }
      100% { transform: skew(0deg); }
    }
    
    /* Эффект радара (сканирующей линии) */
    .radar-scan {
      position: absolute;
      top: -100%; left: 0; right: 0; height: 10px;
      background: linear-gradient(to bottom, transparent, rgba(16, 185, 129, 0.5));
      animation: scan 4s linear infinite;
      pointer-events: none;
      z-index: 50;
    }
    @keyframes scan {
      0% { top: -10%; opacity: 0; }
      10% { opacity: 1; }
      90% { opacity: 1; }
      100% { top: 110%; opacity: 0; }
    }
  `}</style>
);

const Icons = {
    Steam: () => <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor"><path d="M11.979 0C5.364 0 0 5.362 0 11.979c0 4.673 2.68 8.718 6.592 10.742l3.435-4.945c-.244-.229-.404-.546-.404-.897 0-.693.565-1.259 1.259-1.259.693 0 1.258.566 1.258 1.259 0 .193-.047.375-.126.536l3.66 2.605c.801-.131 1.62-.258 2.378-.475 4.398-1.266 7.636-5.32 7.636-10.02C25.688 5.362 20.323 0 13.708 0h-1.729Zm5.358 15.65c-1.39 0-2.52.883-2.906 2.093l-3.327-2.37c.026-.145.045-.292.045-.445 0-1.785-1.448-3.233-3.233-3.233-1.785 0-3.234 1.448-3.234 3.233 0 1.545 1.084 2.844 2.532 3.167l-3.332 4.793A11.933 11.933 0 0 1 0 11.979c0-6.615 5.364-11.979 11.979-11.979 6.616 0 11.98 5.364 11.98 11.979 0 5.176-3.272 9.57-7.851 11.238-.135-2.073-1.854-3.72-3.951-3.72a3.95 3.95 0 0 0-3.834 2.996l3.15 2.242A11.897 11.897 0 0 0 11.98 23.96c6.616 0 11.98-5.365 11.98-11.981 0-6.616-5.364-11.98-11.98-11.98Z"/></svg>,
    Shield: () => <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>,
    Target: () => <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>
};

// Экран Авторизации
const TerminalLogin = ({ onAuthenticating }) => {
    const handleSteamRedirect = () => {
        // Формируем правильный URL для Steam OpenID
        const returnUrl = window.location.origin; // Тот адрес, где запущен фронтенд (например, http://localhost:3000)

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
        <div className="min-h-screen flex flex-col items-center justify-center relative overflow-hidden">
            <div className="radar-scan"></div>
            <div className="z-10 text-center max-w-2xl px-6">
                <div className="mb-8 border border-[rgb(var(--primary))]/30 bg-[rgb(var(--primary))]/5 p-6 rounded-sm shadow-[0_0_15px_rgba(16,185,129,0.1)]">
                    <p className="text-[rgb(var(--primary))] tech-text mb-2 text-sm tracking-[0.3em] uppercase">Secure connection required</p>
                    <h1 className="text-5xl md:text-7xl font-bold tech-text tracking-widest text-white mb-2">
                        SQUAD_HUB
                    </h1>
                    <p className="text-[rgb(var(--muted))] text-sm uppercase tracking-wider">Tactical Evaluation Network</p>
                </div>

                <button
                    onClick={handleSteamRedirect}
                    className="glitch-hover inline-flex items-center gap-4 px-8 py-4 bg-[rgb(var(--surface))] border-2 border-[rgb(var(--border))] text-white font-bold tracking-widest uppercase transition-all duration-300 rounded-sm"
                >
                    <Icons.Steam />
                    <span className="tech-text">Authenticate via Steam</span>
                </button>
            </div>
        </div>
    );
};

// Экран Дашборда
const TacticalDashboard = ({ userId, steamId }) => {
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
        loadStats();
    }, [userId]);

    if (error) return <div className="p-10 tech-text text-red-500">ERROR: {error}</div>;
    if (!stats) return <div className="p-10 tech-text text-[rgb(var(--primary))] animate-pulse">ESTABLISHING UPLINK... DOWNLOADING SERVICE RECORD...</div>;

    const kdRatio = stats.deaths > 0 ? (stats.kills / stats.deaths).toFixed(2) : stats.kills;

    return (
        <div className="min-h-screen p-6 md:p-12 relative">
            <div className="radar-scan opacity-30"></div>

            {/* Шапка */}
            <header className="max-w-6xl mx-auto flex flex-col md:flex-row justify-between items-start md:items-end border-b-2 border-[rgb(var(--primary))]/30 pb-6 mb-10">
                <div>
                    <p className="text-[rgb(var(--primary))] tech-text tracking-widest text-sm mb-2">:: CLASSIFIED :: SERVICE RECORD ::</p>
                    <h1 className="text-4xl font-bold uppercase tech-text">OP_ID: {steamId || userId.split('-')[0]}</h1>
                </div>
                <div className="mt-4 md:mt-0 text-right tech-text text-[rgb(var(--muted))] text-sm flex items-center gap-4">
                    <span>STATUS: <span className="text-[rgb(var(--primary))] animate-pulse">ONLINE</span></span>
                    <button className="border border-[rgb(var(--border))] px-4 py-2 hover:bg-[rgb(var(--surface))] transition-colors" onClick={() => window.location.href='/'}>
                        DISCONNECT
                    </button>
                </div>
            </header>

            {/* Сетка виджетов */}
            <div className="max-w-6xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-6">

                {/* ELO Widget */}
                <div className="col-span-1 md:col-span-2 bg-[rgb(var(--surface))]/80 border border-[rgb(var(--border))] p-8 relative overflow-hidden backdrop-blur-sm">
                    <div className="absolute top-0 right-0 p-4 opacity-10">
                        <Icons.Shield />
                    </div>
                    <h3 className="tech-text text-[rgb(var(--muted))] mb-8 tracking-widest text-sm uppercase flex items-center gap-2">
                        <span className="w-2 h-2 bg-[rgb(var(--primary))] block"></span> Combat Rating (ELO)
                    </h3>
                    <div className="flex items-baseline gap-4 mb-4">
                        <span className="text-7xl font-bold tech-text text-[rgb(var(--primary))]">{stats.eloRating}</span>
                        <span className="text-xl text-[rgb(var(--muted))] tech-text">PTS</span>
                    </div>
                    <div className="w-full bg-[rgb(var(--bg))] h-1 mt-8 relative">
                        <div className="absolute top-0 left-0 h-full bg-[rgb(var(--primary))] shadow-[0_0_10px_rgba(16,185,129,0.8)]" style={{width: `${(stats.eloRating % 100)}%`}}></div>
                    </div>
                    <p className="tech-text text-[rgb(var(--muted))] text-xs mt-2 text-right">PROGRESS TO NEXT TIER</p>
                </div>

                {/* K/D Widget */}
                <div className="bg-[rgb(var(--surface))]/80 border border-[rgb(var(--border))] p-8 backdrop-blur-sm flex flex-col justify-between">
                    <div>
                        <h3 className="tech-text text-[rgb(var(--muted))] mb-6 tracking-widest text-sm uppercase flex items-center gap-2">
                            <span className="w-2 h-2 bg-[rgb(var(--primary))] block"></span> Lethality Index
                        </h3>
                        <div className="space-y-4 tech-text text-lg">
                            <div className="flex justify-between border-b border-[rgb(var(--border))] pb-2">
                                <span className="text-[rgb(var(--muted))]">CONFIRMED KILLS</span>
                                <span className="text-white">{stats.kills}</span>
                            </div>
                            <div className="flex justify-between border-b border-[rgb(var(--border))] pb-2">
                                <span className="text-[rgb(var(--muted))]">KIA</span>
                                <span className="text-white">{stats.deaths}</span>
                            </div>
                        </div>
                    </div>
                    <div className="mt-8 pt-4 border-t-2 border-[rgb(var(--primary))]/30 flex justify-between items-baseline">
                        <span className="tech-text text-[rgb(var(--primary))]">K/D RATIO</span>
                        <span className="text-4xl font-bold tech-text text-white">{kdRatio}</span>
                    </div>
                </div>

                {/* Tactical Info Widget */}
                <div className="col-span-1 md:col-span-3 bg-[rgb(var(--surface))]/80 border border-[rgb(var(--border))] p-8 backdrop-blur-sm grid grid-cols-1 md:grid-cols-3 gap-8">
                    <div className="border-l-2 border-[rgb(var(--primary))] pl-6">
                        <p className="tech-text text-[rgb(var(--muted))] text-xs mb-2 tracking-widest">PRIMARY DEPLOYMENT ROLE</p>
                        <p className="text-3xl font-bold tech-text uppercase text-white">{stats.favoriteRole || "UNKNOWN"}</p>
                    </div>
                    <div className="border-l-2 border-[rgb(var(--primary))] pl-6">
                        <p className="tech-text text-[rgb(var(--muted))] text-xs mb-2 tracking-widest">HOURS IN COMBAT</p>
                        <p className="text-3xl font-bold tech-text text-white">{stats.totalPlaytimeHours} <span className="text-lg text-[rgb(var(--muted))]">HRS</span></p>
                    </div>
                    <div className="border-l-2 border-[rgb(var(--primary))] pl-6">
                        <p className="tech-text text-[rgb(var(--muted))] text-xs mb-2 tracking-widest">ALLIES REVIVED</p>
                        <p className="text-3xl font-bold tech-text text-[rgb(var(--primary))]">{stats.revives}</p>
                    </div>
                </div>

            </div>
        </div>
    );
};

export default function App() {
    const [authState, setAuthState] = useState({ userId: null, steamId: null, isLoading: true });

    useEffect(() => {
        // 1. Проверяем URL на наличие параметров возврата от Steam
        const urlParams = new URLSearchParams(window.location.search);
        const mode = urlParams.get('openid.mode');

        // Если мы вернулись из Steam с успешным логином
        if (mode === 'id_res') {
            const processSteamRedirect = async () => {
                try {
                    // Превращаем параметры URL в плоский объект
                    const paramsObject = Object.fromEntries(urlParams.entries());
                    const paramsJsonString = JSON.stringify(paramsObject);

                    // Отправляем JSON-строку на наш Gateway
                    const data = await fetchGraphQL(MUTATIONS.LOGIN_STEAM, { paramsJson: paramsJsonString });

                    const { userId, steamId, token } = data.loginWithSteam;

                    // Сохраняем токен и ID для сессии
                    localStorage.setItem('squad_jwt', token);
                    localStorage.setItem('squad_user_id', userId);
                    localStorage.setItem('squad_steam_id', steamId);

                    // Очищаем грязный URL в браузере (без перезагрузки страницы)
                    window.history.replaceState({}, document.title, window.location.pathname);

                    setAuthState({ userId, steamId, isLoading: false });
                } catch (error) {
                    console.error("Steam Auth Error:", error);
                    setAuthState({ userId: null, steamId: null, isLoading: false });
                }
            };

            processSteamRedirect();
        } else {
            // 2. Если параметров в URL нет, проверяем, есть ли уже сохраненная сессия
            const savedUserId = localStorage.getItem('squad_user_id');
            const savedSteamId = localStorage.getItem('squad_steam_id');
            if (savedUserId) {
                setAuthState({ userId: savedUserId, steamId: savedSteamId, isLoading: false });
            } else {
                setAuthState({ userId: null, steamId: null, isLoading: false });
            }
        }
    }, []);

    if (authState.isLoading) {
        return (
            <div className="min-h-screen bg-[#09090b] flex items-center justify-center">
                <div className="tech-text text-[#10b989] text-xl animate-pulse">DECRYPTING TRANSMISSION...</div>
            </div>
        );
    }

    return (
        <>
            <GlobalStyles />
            {!authState.userId ? (
                <TerminalLogin />
            ) : (
                <TacticalDashboard userId={authState.userId} steamId={authState.steamId} />
            )}
        </>
    );
}