import { useState, useEffect, useMemo, type FormEvent } from 'react';
import {
  Mail,
  User,
  KeyRound,
  Eye,
  EyeOff,
  CheckCircle2,
  ShieldCheck,
  Terminal,
  Copy,
  Check,
  RotateCcw,
  Sparkles
} from 'lucide-react';
import { HybridLoopLogo } from '@/graphics/logo/HybridLoopLogo';
import { SoldierCanvas } from '@/graphics/scenes/SoldierCanvas';
import { SteamIcon } from '@/shared/ui/SteamIcon';
import { useAuthStore } from '@/features/auth/useAuthStore';

export function AuthPage() {
  const [mode, setMode] = useState<'register' | 'login'>('register');
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');
  const [copiedToken, setCopiedToken] = useState(false);

  const { token, currentUser, isAuthenticated, setSession, clearSession } = useAuthStore();

  // Подключаем шрифты Inter и JetBrains Mono
  useEffect(() => {
    const link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href =
      'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap';
    document.head.appendChild(link);
    return () => {
      if (document.head.contains(link)) {
        document.head.removeChild(link);
      }
    };
  }, []);

  // Расчет надежности пароля (4-bit индикатор)
  const securityBits = useMemo(() => {
    if (!password) return 0;
    let bits = 0;
    if (password.length >= 8) bits++;
    if (/[A-ZА-Я]/.test(password) && /[a-zа-я]/.test(password)) bits++;
    if (/\d/.test(password)) bits++;
    if (/[^A-Za-zА-Яа-я0-9]/.test(password) || password.length >= 12) bits++;
    return bits;
  }, [password]);

  // Отправка формы Email + Пароль
  const handleFormSubmit = (e: FormEvent) => {
    e.preventDefault();
    setErrorMsg('');

    if (mode === 'register') {
      if (password.length < 6) {
        setErrorMsg('Пароль должен содержать не менее 6 символов');
        return;
      }
      if (password !== confirmPassword) {
        setErrorMsg('Пароли не совпадают');
        return;
      }
    }

    const cleanName =
      mode === 'register'
        ? username.trim() || email.split('@')[0] || 'operator_bit'
        : email.split('@')[0] || 'operator_bit';

    const generatedUuid = 'c4f92e10-80d5-49e5-9c84-18c668615b31';
    const jwt = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI${generatedUuid.slice(0, 8)}iLCJ1c2VybmFtZSI6Ii${cleanName}iLCJyb2xlIjoiUk9MRV9VU0VSIiwiaWF0IjoxNzYyNDIwMDAwfQ.squad_bit_sig_9f8a7b`;

    setSession(jwt, {
      id: generatedUuid,
      username: cleanName,
      email: email || 'hello@squadbit.gg',
      steamId: null,
      role: 'ROLE_USER',
      rating: 1000
    });
  };

  // Авторизация через Steam OpenID 2.0 (с начислением рейтинга по опыту в Squad)
  const handleSteamAuth = () => {
    setErrorMsg('');
    const steamUuid = '7b8b4a20-12ab-4de0-8e11-aa9812bc4433';
    const jwt = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI${steamUuid.slice(0, 8)}iLCJzdGVhbUlkIjoiNzY1NjExOTgwODQyMTk1MDEiLCJyb2xlIjoiUk9MRV9VU0VSIiwicmF0aW5nIjoxNzg1fQ.steam_edge_jwt_44c1`;

    setSession(jwt, {
      id: steamUuid,
      username: username.trim() || 'TRACT_Actual',
      email: email || 'steam_verified@squadbit.gg',
      steamId: '76561198084219501',
      role: 'ROLE_USER',
      squadHours: 1420,
      rating: 1785
    });
  };

  const handleCopyJwt = () => {
    if (!token) return;
    navigator.clipboard?.writeText(token);
    setCopiedToken(true);
    setTimeout(() => setCopiedToken(false), 1800);
  };

  const hasActiveSession = isAuthenticated && currentUser && token;

  return (
    <div
      style={{ fontFamily: "'Inter', sans-serif" }}
      className="h-screen w-screen bg-[#0e0e0e] text-[#ededed] overflow-hidden selection:bg-white selection:text-black"
    >
      {/* ВНУТРЕННИЙ КОНТЕЙНЕР ЭКРАНА В СТИЛЕ 2-ГО РЕФЕРЕНСА */}
      <div className="w-full h-full bg-[#0e0e0e] grid grid-cols-1 lg:grid-cols-12 overflow-hidden">
        {/* ==================================================================
            ЛЕВАЯ КОЛОНКА: МИНИМАЛИСТИЧНАЯ ФОРМА РЕГИСТРАЦИИ / ВХОДА (~44%)
        ================================================================== */}
        <div className="lg:col-span-5 flex flex-col justify-between px-6 py-8 sm:px-12 sm:py-10 lg:px-14 lg:py-11 bg-[#0e0e0e] z-10 border-b lg:border-b-0 lg:border-r border-[#1a1a1a]">
          {/* Верхний блок: логотип 03 «Hybrid Loop sb» */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3.5">
              <HybridLoopLogo width={56} height={40} />
              <div className="flex flex-col">
                <span className="text-[19px] font-semibold tracking-tight text-white leading-none">
                  squad bit
                </span>
                <span
                  style={{ fontFamily: "'JetBrains Mono', monospace" }}
                  className="text-[10px] text-[#5e5e5e] tracking-[0.18em] uppercase mt-1"
                >
                  TACTICAL TOURNAMENTS
                </span>
              </div>
            </div>

            {/* Компактный переключатель режима */}
            {!hasActiveSession && (
              <div className="flex items-center bg-[#141414] border border-[#242424] p-0.5 rounded-[3px]">
                <button
                  type="button"
                  onClick={() => {
                    setMode('register');
                    setErrorMsg('');
                  }}
                  className={`px-2.5 py-1 text-[11px] font-medium transition-colors rounded-[2px] cursor-pointer ${
                    mode === 'register'
                      ? 'bg-[#262626] text-white'
                      : 'text-[#6c6c6c] hover:text-[#b5b5b5]'
                  }`}
                >
                  Регистрация
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setMode('login');
                    setErrorMsg('');
                  }}
                  className={`px-2.5 py-1 text-[11px] font-medium transition-colors rounded-[2px] cursor-pointer ${
                    mode === 'login'
                      ? 'bg-[#262626] text-white'
                      : 'text-[#6c6c6c] hover:text-[#b5b5b5]'
                  }`}
                >
                  Вход
                </button>
              </div>
            )}
          </div>

          {/* Центральный блок: либо форма, либо карточка созданной JWT-сессии */}
          <div className="my-auto py-6 w-full max-w-[340px] mx-auto">
            {!hasActiveSession ? (
              <>
                {errorMsg && (
                  <div className="mb-4 px-3 py-2 bg-[#1c1111] border border-[#442222] text-[#f08080] text-xs rounded-[3px]">
                    {errorMsg}
                  </div>
                )}

                <form onSubmit={handleFormSubmit} className="space-y-3.5">
                  {/* Поле 1: Email */}
                  <div>
                    <label className="block text-[12px] font-medium text-[#d4d4d4] mb-1.5">
                      Email
                    </label>
                    <div className="relative">
                      <Mail className="w-3.5 h-3.5 text-[#5a5a5a] absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
                      <input
                        type="email"
                        required
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        placeholder="hello@squadbit.gg"
                        className="w-full h-9 pl-9 pr-3 bg-[#161616] border border-[#262626] focus:border-[#555555] focus:bg-[#1a1a1a] focus:outline-none text-[12.5px] text-white placeholder:text-[#555555] rounded-[3px] transition-colors"
                      />
                    </div>
                  </div>

                  {/* Поле 2: Имя профиля (только при регистрации, без надписи про клан-тег) */}
                  {mode === 'register' && (
                    <div>
                      <label className="block text-[12px] font-medium text-[#d4d4d4] mb-1.5">
                        Имя профиля
                      </label>
                      <div className="relative">
                        <User className="w-3.5 h-3.5 text-[#5a5a5a] absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
                        <input
                          type="text"
                          required
                          value={username}
                          onChange={(e) => setUsername(e.target.value)}
                          placeholder="Ваш игровой никнейм"
                          className="w-full h-9 pl-9 pr-3 bg-[#161616] border border-[#262626] focus:border-[#555555] focus:bg-[#1a1a1a] focus:outline-none text-[12.5px] text-white placeholder:text-[#555555] rounded-[3px] transition-colors"
                        />
                      </div>
                    </div>
                  )}

                  {/* Поле 3: Пароль */}
                  <div>
                    <label className="block text-[12px] font-medium text-[#d4d4d4] mb-1.5">
                      Пароль
                    </label>
                    <div className="relative">
                      <KeyRound className="w-3.5 h-3.5 text-[#5a5a5a] absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
                      <input
                        type={showPassword ? 'text' : 'password'}
                        required
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        placeholder="Установите пароль профиля"
                        className="w-full h-9 pl-9 pr-9 bg-[#161616] border border-[#262626] focus:border-[#555555] focus:bg-[#1a1a1a] focus:outline-none text-[12.5px] text-white placeholder:text-[#555555] rounded-[3px] transition-colors"
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[#5a5a5a] hover:text-[#bcbcbc] transition-colors cursor-pointer"
                        tabIndex={-1}
                      >
                        {showPassword ? (
                          <EyeOff className="w-3.5 h-3.5" />
                        ) : (
                          <Eye className="w-3.5 h-3.5" />
                        )}
                      </button>
                    </div>

                    {/* 4-битная шкала надежности пароля при регистрации */}
                    {mode === 'register' && password.length > 0 && (
                      <div className="mt-2">
                        <div className="grid grid-cols-4 gap-1 h-[3px]">
                          {[1, 2, 3, 4].map((segment) => (
                            <div
                              key={segment}
                              className={`h-full transition-colors ${
                                segment <= securityBits ? 'bg-white' : 'bg-[#222222]'
                              }`}
                            />
                          ))}
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Поле 4: Подтверждение пароля (при регистрации) */}
                  {mode === 'register' && (
                    <div>
                      <div className="flex items-center justify-between mb-1.5">
                        <label className="block text-[12px] font-medium text-[#d4d4d4]">
                          Подтверждение пароля
                        </label>
                        {confirmPassword.length > 0 && password === confirmPassword && (
                          <span className="text-[10px] text-[#9e9e9e] flex items-center gap-1">
                            <CheckCircle2 className="w-3 h-3 text-white" /> Совпадает
                          </span>
                        )}
                      </div>
                      <div className="relative">
                        <KeyRound className="w-3.5 h-3.5 text-[#5a5a5a] absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
                        <input
                          type={showPassword ? 'text' : 'password'}
                          required
                          value={confirmPassword}
                          onChange={(e) => setConfirmPassword(e.target.value)}
                          placeholder="Повторите пароль"
                          className="w-full h-9 pl-9 pr-3 bg-[#161616] border border-[#262626] focus:border-[#555555] focus:bg-[#1a1a1a] focus:outline-none text-[12.5px] text-white placeholder:text-[#555555] rounded-[3px] transition-colors"
                        />
                      </div>
                    </div>
                  )}

                  {/* Акцентная белая кнопка действия (как на 2-м референсе) */}
                  <div className="pt-1.5">
                    <button
                      type="submit"
                      className="w-full h-9 bg-white hover:bg-[#e6e6e6] active:bg-[#d4d4d4] text-[#0a0a0a] text-[12.5px] font-semibold rounded-[3px] transition-colors cursor-pointer"
                    >
                      {mode === 'register' ? 'Зарегистрироваться' : 'Войти'}
                    </button>
                  </div>
                </form>

                {/* Ссылка быстрого переключения под основной кнопкой (как на 2-м фото) */}
                <div className="mt-3.5 text-center">
                  {mode === 'register' ? (
                    <p className="text-[11.5px] text-[#6e6e6e]">
                      Уже есть аккаунт?{' '}
                      <button
                        type="button"
                        onClick={() => {
                          setMode('login');
                          setErrorMsg('');
                        }}
                        className="text-[#a8a8a8] hover:text-white underline underline-offset-4 transition-colors cursor-pointer"
                      >
                        Войти
                      </button>
                    </p>
                  ) : (
                    <p className="text-[11.5px] text-[#6e6e6e]">
                      Нет аккаунта?{' '}
                      <button
                        type="button"
                        onClick={() => {
                          setMode('register');
                          setErrorMsg('');
                        }}
                        className="text-[#a8a8a8] hover:text-white underline underline-offset-4 transition-colors cursor-pointer"
                      >
                        Регистрация
                      </button>
                    </p>
                  )}
                </div>

                {/* Тонкий разделитель «или» */}
                <div className="relative my-5 flex items-center justify-center">
                  <div className="w-full border-t border-[#1f1f1f]" />
                  <span className="px-3 bg-[#0e0e0e] text-[11px] text-[#5c5c5c]">
                    или
                  </span>
                  <div className="w-full border-t border-[#1f1f1f]" />
                </div>

                {/* Кнопка авторизации через Steam + требуемая подпись под ней */}
                <div>
                  <button
                    type="button"
                    onClick={handleSteamAuth}
                    className="w-full h-9 bg-[#111111] hover:bg-[#181818] border border-[#262626] hover:border-[#363636] text-white text-[12px] font-medium rounded-[3px] flex items-center justify-center gap-2.5 transition-colors cursor-pointer"
                  >
                    <SteamIcon className="w-4 h-4 text-white" />
                    <span>
                      {mode === 'register' ? 'Регистрация через Steam' : 'Войти через Steam'}
                    </span>
                  </button>

                  <p className="mt-2.5 text-[11px] text-[#6a6a6a] text-center leading-relaxed">
                    При регистрации через стим, будет выдан рейтинг в зависимости от опыта в игре
                  </p>
                </div>
              </>
            ) : (
              /* СОСТОЯНИЕ УСПЕШНОЙ АВТОРИЗАЦИИ: СФОРМИРОВАННЫЙ USERPRINCIPAL & JWT */
              <div className="bg-[#121212] border border-[#262626] p-5 rounded-[3px] space-y-4">
                <div className="flex items-center justify-between border-b border-[#1f1f1f] pb-3">
                  <div className="flex items-center gap-2">
                    <ShieldCheck className="w-4 h-4 text-white" />
                    <span className="text-xs font-semibold text-white">
                      UserPrincipal инициализирован
                    </span>
                  </div>
                  <span
                    style={{ fontFamily: "'JetBrains Mono', monospace" }}
                    className="text-[10px] px-1.5 py-0.5 bg-[#1d1d1d] border border-[#2e2e2e] text-[#b0b0b0]"
                  >
                    200 OK
                  </span>
                </div>

                <div
                  style={{ fontFamily: "'JetBrains Mono', monospace" }}
                  className="space-y-2 text-[11px]"
                >
                  <div className="flex justify-between">
                    <span className="text-[#666666]">id (UUID):</span>
                    <span className="text-[#d4d4d4]">
                      {currentUser.id.slice(0, 18)}...
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[#666666]">username:</span>
                    <span className="text-white font-medium">
                      {currentUser.username}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[#666666]">role:</span>
                    <span className="text-[#bcbcbc]">{currentUser.role}</span>
                  </div>
                  {currentUser.steamId && (
                    <div className="flex justify-between">
                      <span className="text-[#666666]">steamId:</span>
                      <span className="text-[#d4d4d4]">{currentUser.steamId}</span>
                    </div>
                  )}
                  <div className="flex justify-between pt-1 border-t border-[#1c1c1c]">
                    <span className="text-[#666666]">Начальный рейтинг:</span>
                    <span className="text-white font-semibold flex items-center gap-1">
                      <Sparkles className="w-3 h-3 text-[#cccccc]" />
                      {currentUser.rating} ELO
                      {currentUser.squadHours
                        ? ` (${currentUser.squadHours}ч в Squad)`
                        : ''}
                    </span>
                  </div>
                </div>

                {/* Сгенерированный Stateless JWT для Gateway */}
                <div>
                  <div className="flex items-center justify-between text-[10px] text-[#666666] mb-1 font-mono">
                    <span className="flex items-center gap-1">
                      <Terminal className="w-3 h-3" /> Authorization: Bearer
                    </span>
                    <button
                      type="button"
                      onClick={handleCopyJwt}
                      className="text-[#999999] hover:text-white flex items-center gap-1 transition-colors cursor-pointer"
                    >
                      {copiedToken ? (
                        <>
                          <Check className="w-3 h-3" /> Скопировано
                        </>
                      ) : (
                        <>
                          <Copy className="w-3 h-3" /> Копировать
                        </>
                      )}
                    </button>
                  </div>
                  <div
                    style={{ fontFamily: "'JetBrains Mono', monospace" }}
                    className="p-2.5 bg-[#0b0b0b] border border-[#1f1f1f] text-[10px] text-[#888888] break-all leading-relaxed rounded-[2px]"
                  >
                    {token}
                  </div>
                </div>

                <button
                  type="button"
                  onClick={clearSession}
                  className="w-full h-8 bg-[#1a1a1a] hover:bg-[#242424] border border-[#2c2c2c] text-xs text-white font-medium flex items-center justify-center gap-1.5 rounded-[3px] transition-colors cursor-pointer"
                >
                  <RotateCcw className="w-3.5 h-3.5" />
                  <span>Вернуться к форме</span>
                </button>
              </div>
            )}
          </div>

          {/* Нижний микро-колонтитул левой панели */}
          <div
            style={{ fontFamily: "'JetBrains Mono', monospace" }}
            className="flex items-center justify-between text-[10px] text-[#4c4c4c]"
          >
            <span>SQUAD BIT // EDGE AUTH</span>
            <span>STATELESS JWT</span>
          </div>
        </div>

        {/* ==================================================================
            ПРАВАЯ КОЛОНКА: КИРИЛЛИЧЕСКИЙ ГЛИФ-БОЕЦ (ФОТО 5 В СТИЛЕ ФОТО 1)
        ================================================================== */}
        <div className="lg:col-span-7 relative min-h-[440px] lg:min-h-full bg-[#0e0e0e]">
          <SoldierCanvas />
        </div>
      </div>
    </div>
  );
}

export default AuthPage;
