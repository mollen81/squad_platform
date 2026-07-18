import React, { useState, useEffect } from 'react';

const GlobalStyles = () => (
    <style>{`
    @import url('https://fonts.cdnfonts.com/css/nohemi');
    
    :root {
      /* Палитра Onyx & Candy Blue */
      --bg: 2 2 2;
      --card: 15 15 15;
      --primary: 178 213 229;
      --text: 255 255 255;
      --muted: 161 161 170;
      --font-heading: 'Nohemi', sans-serif;
      --font-body: 'Inter', system-ui, sans-serif;
    }
    
    body {
      background-color: rgb(var(--bg));
      color: rgb(var(--text));
      font-family: var(--font-body);
      margin: 0;
      padding: 0;
      -webkit-font-smoothing: antialiased;
    }

    h1, h2, h3, .font-heading {
      font-family: var(--font-heading);
    }
    
    ::-webkit-scrollbar { width: 8px; }
    ::-webkit-scrollbar-track { background: rgb(var(--bg)); }
    ::-webkit-scrollbar-thumb { background: rgb(var(--card)); border-radius: 4px; }
    ::-webkit-scrollbar-thumb:hover { background: rgb(var(--muted)); }

    @keyframes fadeInUp {
      from { opacity: 0; transform: translateY(20px); }
      to { opacity: 1; transform: translateY(0); }
    }
    .animate-fade-in-up {
      animation: fadeInUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards;
      opacity: 0;
    }
  `}</style>
);

const Icons = {
    Activity: () => <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>,
    Crosshair: () => <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><line x1="22" x2="18" y1="12" y2="12"/><line x1="6" x2="2" y1="12" y2="12"/><line x1="12" x2="12" y1="6" y2="2"/><line x1="12" x2="12" y1="22" y2="18"/></svg>,
    Steam: () => <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor"><path d="M11.979 0C5.364 0 0 5.362 0 11.979c0 4.673 2.68 8.718 6.592 10.742l3.435-4.945c-.244-.229-.404-.546-.404-.897 0-.693.565-1.259 1.259-1.259.693 0 1.258.566 1.258 1.259 0 .193-.047.375-.126.536l3.66 2.605c.801-.131 1.62-.258 2.378-.475 4.398-1.266 7.636-5.32 7.636-10.02C25.688 5.362 20.323 0 13.708 0h-1.729Zm5.358 15.65c-1.39 0-2.52.883-2.906 2.093l-3.327-2.37c.026-.145.045-.292.045-.445 0-1.785-1.448-3.233-3.233-3.233-1.785 0-3.234 1.448-3.234 3.233 0 1.545 1.084 2.844 2.532 3.167l-3.332 4.793A11.933 11.933 0 0 1 0 11.979c0-6.615 5.364-11.979 11.979-11.979 6.616 0 11.98 5.364 11.98 11.979 0 5.176-3.272 9.57-7.851 11.238-.135-2.073-1.854-3.72-3.951-3.72a3.95 3.95 0 0 0-3.834 2.996l3.15 2.242A11.897 11.897 0 0 0 11.98 23.96c6.616 0 11.98-5.365 11.98-11.981 0-6.616-5.364-11.98-11.98-11.98Z"/></svg>,
    Refresh: () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>,
    LogOut: () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/></svg>
};

const fetchGraphQL = async (query, variables = {}) => {
    try {
        const response = await fetch('http://localhost:8080/graphql', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ query, variables }),
        });
        const json = await response.json();
        if (json.errors) throw new Error(json.errors[0].message);
        return json.data;
    } catch (error) {
        console.error("GraphQL Request Failed:", error);
        throw error;
    }
};

const AuthScreen = ({ onLogin }) => {
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState(null);
    const [steamId, setSteamId] = useState("76561198888277695"); // Ваш SteamID для удобства

    const handleSteamLogin = async (e) => {
        e.preventDefault();
        if (!steamId.trim()) return;

        setIsLoading(true);
        setError(null);
        try {
            const mutation = `
        mutation($params: String!) {
          loginWithSteam(openidParamsJson: $params) {
            userId
            steamId
            isNewUser
          }
        }
      `;

            const paramsJson = JSON.stringify({ test_steam_id: steamId });
            const data = await fetchGraphQL(mutation, { params: paramsJson });

            // Имитируем ожидание Kafka для новых пользователей (пока она собирает стату)
            if (data.loginWithSteam.isNewUser) {
                await new Promise(resolve => setTimeout(resolve, 3000));
            }

            onLogin(data.loginWithSteam.userId);
        } catch (err) {
            setError("Ошибка связи с сервером. Убедитесь, что Gateway запущен на порту 8080.");
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex flex-col items-center justify-center relative overflow-hidden bg-[rgb(var(--bg))] p-6">
            <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-[rgb(var(--primary))] opacity-[0.03] rounded-full blur-[100px] pointer-events-none"></div>

            <div className="z-10 w-full max-w-md text-center space-y-8 animate-fade-in-up">
                <div>
                    <h2 className="text-[rgb(var(--muted))] tracking-[0.2em] text-sm uppercase mb-4 font-semibold">Киберспортивная Платформа</h2>
                    <h1 className="text-5xl md:text-7xl font-bold font-heading text-transparent bg-clip-text bg-gradient-to-b from-white to-[rgb(var(--muted))] mb-4">
                        SQUAD HUB
                    </h1>
                </div>

                <form onSubmit={handleSteamLogin} className="bg-[rgb(var(--card))] border border-[rgb(var(--primary))]/10 p-8 rounded-3xl shadow-2xl space-y-6 relative overflow-hidden">
                    <div className="space-y-2 text-left relative z-10">
                        <label className="text-sm font-medium text-[rgb(var(--muted))] ml-1">SteamID (Тестовый вход)</label>
                        <input
                            type="text"
                            value={steamId}
                            onChange={(e) => setSteamId(e.target.value)}
                            className="w-full bg-[rgb(var(--bg))] border border-[rgb(var(--muted))]/20 text-white rounded-xl px-4 py-3 outline-none focus:border-[rgb(var(--primary))]/50 transition-colors font-mono"
                        />
                    </div>

                    {error && (
                        <div className="bg-red-500/10 border border-red-500/20 text-red-400 p-3 rounded-xl text-sm relative z-10">
                            {error}
                        </div>
                    )}

                    <button
                        type="submit"
                        disabled={isLoading}
                        className="w-full relative z-10 inline-flex items-center justify-center gap-3 px-8 py-4 bg-[rgb(var(--primary))] text-[rgb(var(--bg))] rounded-xl font-heading font-bold text-lg transition-all hover:scale-[1.02] hover:shadow-[0_0_30px_-5px_rgb(var(--primary))] disabled:opacity-50 disabled:pointer-events-none"
                    >
                        {isLoading ? (
                            <span className="animate-pulse flex items-center gap-2">
                <Icons.Refresh /> Идет анализ профиля...
              </span>
                        ) : (
                            <>
                                <Icons.Steam /> <span>Войти через Steam</span>
                            </>
                        )}
                    </button>
                </form>
            </div>
        </div>
    );
};

const DashboardScreen = ({ userId, onLogout }) => {
    const [stats, setStats] = useState(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState(null);

    const loadStats = async () => {
        setIsLoading(true);
        setError(null);
        try {
            const query = `
        query($userId: String!) {
          getPlayerStats(userId: $userId) {
            eloRating
            kills
            deaths
            revives
            favouriteRole
            totalPlaytimeHours
          }
        }
      `;
            const data = await fetchGraphQL(query, { userId });
            setStats(data.getPlayerStats);
        } catch (err) {
            setError(err.message || "Не удалось загрузить данные");
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        loadStats();
    }, [userId]);

    if (isLoading) {
        return (
            <div className="min-h-screen flex flex-col items-center justify-center bg-[rgb(var(--bg))]">
                <div className="w-12 h-12 border-4 border-[rgb(var(--card))] border-t-[rgb(var(--primary))] rounded-full animate-spin mb-4"></div>
                <p className="text-[rgb(var(--primary))] font-heading tracking-widest animate-pulse">РАСЧЕТ ELO...</p>
            </div>
        );
    }

    if (error || !stats) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-[rgb(var(--bg))] text-white p-6">
                <div className="bg-[rgb(var(--card))] p-8 rounded-3xl text-center space-y-4 border border-red-500/20 max-w-sm w-full">
                    <p className="text-red-400 font-medium">{error || "Статистика не найдена"}</p>
                    <button onClick={onLogout} className="w-full px-6 py-3 bg-[rgb(var(--bg))] hover:bg-[rgb(var(--bg))]/80 transition-colors rounded-xl font-heading">Вернуться</button>
                </div>
            </div>
        );
    }

    const kdRatio = stats.deaths > 0 ? (stats.kills / stats.deaths).toFixed(2) : stats.kills;
    const rankProgress = ((stats.eloRating % 100) / 100) * 100;

    return (
        <div className="min-h-screen flex bg-[rgb(var(--bg))] selection:bg-[rgb(var(--primary))] selection:text-[rgb(var(--bg))]">

            {/* Sidebar Area */}
            <aside className="w-20 lg:w-64 border-r border-[rgb(var(--card))] bg-[rgb(var(--bg))] flex flex-col p-6 z-20 hidden md:flex">
                <div className="flex items-center gap-4 mb-12 text-[rgb(var(--primary))]">
                    <div className="w-10 h-10 rounded-xl bg-[rgb(var(--primary))] text-[rgb(var(--bg))] flex items-center justify-center font-heading font-bold text-xl">
                        SQ
                    </div>
                    <span className="font-bold text-2xl font-heading tracking-wide hidden lg:block">SQUAD</span>
                </div>

                <nav className="space-y-3 flex-1">
                    <button className="w-full flex items-center gap-4 px-4 py-3 bg-[rgb(var(--card))] text-[rgb(var(--primary))] rounded-xl font-medium shadow-sm border border-[rgb(var(--primary))]/10">
                        <Icons.Activity /> <span className="hidden lg:block font-heading">Сводка</span>
                    </button>
                </nav>

                <button onClick={onLogout} className="mt-auto flex items-center gap-4 px-4 py-3 text-[rgb(var(--muted))] hover:text-red-400 transition-colors">
                    <Icons.LogOut /> <span className="hidden lg:block">Выйти</span>
                </button>
            </aside>

            {/* Main Content */}
            <main className="flex-1 p-6 lg:p-12 overflow-y-auto">
                <div className="max-w-5xl mx-auto w-full space-y-8">

                    <header className="flex justify-between items-end animate-fade-in-up">
                        <div>
                            <p className="text-[rgb(var(--muted))] mb-1 tracking-wider uppercase text-xs font-semibold">Ваш профиль</p>
                            <h1 className="text-4xl lg:text-5xl font-bold font-heading uppercase text-white">Боец #{userId.split('-')[0]}</h1>
                        </div>
                        <button onClick={loadStats} className="hidden sm:flex items-center gap-2 px-5 py-2.5 bg-[rgb(var(--card))] hover:border-[rgb(var(--primary))]/50 border border-[rgb(var(--card))] text-white rounded-xl transition-all font-heading text-sm">
                            <Icons.Refresh /> Обновить
                        </button>
                    </header>

                    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">

                        {/* ELO Rating Card */}
                        <div className="col-span-1 lg:col-span-2 relative overflow-hidden rounded-3xl border border-[rgb(var(--card))] bg-[rgb(var(--card))] p-8 shadow-2xl animate-fade-in-up" style={{animationDelay: '100ms'}}>
                            <div className="absolute -right-20 -top-20 w-64 h-64 bg-[rgb(var(--primary))] opacity-10 rounded-full blur-[60px]"></div>

                            <h3 className="text-[rgb(var(--muted))] font-medium mb-6 font-heading tracking-wide uppercase text-sm">Рейтинг Скилла</h3>

                            <div className="flex items-baseline gap-4 mb-2">
                                <span className="text-7xl lg:text-8xl font-bold font-heading text-[rgb(var(--primary))]">{stats.eloRating}</span>
                                <span className="text-2xl text-white font-heading font-medium">ELO</span>
                            </div>

                            <div className="mt-12 space-y-3 relative z-10">
                                <div className="flex justify-between text-sm font-medium text-[rgb(var(--muted))] font-heading uppercase">
                                    <span>Текущий прогресс</span>
                                    <span>До повышения</span>
                                </div>
                                <div className="h-3 w-full bg-[rgb(var(--bg))] rounded-full overflow-hidden border border-[rgb(var(--muted))]/10">
                                    <div className="h-full bg-[rgb(var(--primary))] rounded-full transition-all duration-1000" style={{ width: `${rankProgress}%` }}></div>
                                </div>
                            </div>
                        </div>

                        {/* K/D Efficiency Card */}
                        <div className="rounded-3xl border border-[rgb(var(--card))] bg-[rgb(var(--card))] p-8 shadow-xl flex flex-col justify-between animate-fade-in-up" style={{animationDelay: '200ms'}}>
                            <div>
                                <div className="flex items-center gap-3 mb-8 text-[rgb(var(--muted))]">
                                    <Icons.Crosshair />
                                    <h3 className="font-medium font-heading tracking-wide uppercase text-sm">Эффективность</h3>
                                </div>
                                <div className="space-y-5">
                                    <div className="flex justify-between items-end border-b border-[rgb(var(--bg))] pb-4">
                                        <span className="text-[rgb(var(--muted))] font-medium">Kills</span>
                                        <span className="text-2xl font-bold font-heading text-white">{stats.kills}</span>
                                    </div>
                                    <div className="flex justify-between items-end border-b border-[rgb(var(--bg))] pb-4">
                                        <span className="text-[rgb(var(--muted))] font-medium">Deaths</span>
                                        <span className="text-2xl font-bold font-heading text-white">{stats.deaths}</span>
                                    </div>
                                </div>
                            </div>
                            <div className="pt-6 mt-6 border-t border-[rgb(var(--bg))] flex justify-between items-center">
                                <span className="text-[rgb(var(--primary))] font-bold font-heading uppercase">K/D Ratio</span>
                                <span className="text-4xl font-bold font-heading text-[rgb(var(--primary))]">{kdRatio}</span>
                            </div>
                        </div>

                        {/* Playstyle & Roles Funnel */}
                        <div className="col-span-1 lg:col-span-3 grid grid-cols-1 md:grid-cols-2 gap-6 animate-fade-in-up" style={{animationDelay: '300ms'}}>

                            <div className="rounded-3xl border border-[rgb(var(--card))] bg-[rgb(var(--card))] p-8 shadow-xl">
                                <h3 className="text-[rgb(var(--muted))] font-medium mb-6 font-heading tracking-wide uppercase text-sm">Основная Роль</h3>
                                <div className="w-full flex flex-col items-center py-4 space-y-3">
                                    <div className="w-full h-14 bg-[rgb(var(--primary))] rounded-xl flex items-center justify-center shadow-lg">
                                        <span className="text-[rgb(var(--bg))] font-bold font-heading text-lg tracking-widest uppercase">{stats.favouriteRole || "Rifleman"}</span>
                                    </div>
                                    <div className="w-[85%] h-12 bg-[rgb(var(--primary))] rounded-xl flex items-center justify-center opacity-70">
                                        <span className="text-[rgb(var(--bg))] font-bold font-heading text-sm uppercase tracking-wider">Отыграно: {stats.totalPlaytimeHours || 0} часов</span>
                                    </div>
                                    <div className="w-[65%] h-10 bg-[rgb(var(--primary))] rounded-xl flex items-center justify-center opacity-40">
                                        <span className="text-[rgb(var(--bg))] font-bold font-heading text-xs uppercase tracking-wider">Уникальный стиль</span>
                                    </div>
                                </div>
                            </div>

                            <div className="rounded-3xl border border-[rgb(var(--card))] bg-[rgb(var(--card))] p-8 shadow-xl flex flex-col justify-center items-center text-center">
                                <h3 className="text-[rgb(var(--muted))] font-medium mb-4 font-heading tracking-wide uppercase text-sm w-full text-left">Командная игра</h3>
                                <div className="w-24 h-24 rounded-full border-4 border-[rgb(var(--primary))]/30 flex flex-col items-center justify-center mb-4">
                                    <span className="text-3xl font-bold font-heading text-[rgb(var(--primary))]">{stats.revives}</span>
                                </div>
                                <p className="text-white font-heading font-bold text-lg uppercase tracking-wide">Поднято союзников</p>
                                <p className="text-[rgb(var(--muted))] text-sm mt-2">Медики гордятся тобой</p>
                            </div>

                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
};

export default function App() {
    const [userId, setUserId] = useState(null);

    return (
        <>
            <GlobalStyles />
            {!userId ? (
                <AuthScreen onLogin={setUserId} />
            ) : (
                <DashboardScreen userId={userId} onLogout={() => setUserId(null)} />
            )}
        </>
    );
}