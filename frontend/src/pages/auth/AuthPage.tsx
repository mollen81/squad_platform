import { useState, useMemo, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Mail,
  User,
  Lock,
  Eye,
  EyeClosed,
  CheckCircle,
  ShieldCheck,
  Terminal,
  Copy,
  Check,
  Refresh,
  Star,
  Group
} from 'iconoir-react';
import { Loading } from 'loading-dev';
import { HybridLoopLogo } from '@/graphics/logo/HybridLoopLogo';
import { SoldierCanvas } from '@/graphics/scenes/SoldierCanvas';
import { SteamIcon } from '@/shared/ui/SteamIcon';
import { useAuthStore } from '@/features/auth/useAuthStore';

export function AuthPage() {
  const navigate = useNavigate();
  const [mode, setMode] = useState<'register' | 'login'>('register');
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');
  const [copiedToken, setCopiedToken] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { token, currentUser, isAuthenticated, setSession, clearSession } = useAuthStore();

  const securityBits = useMemo(() => {
    if (!password) return 0;
    let bits = 0;
    if (password.length >= 8) bits++;
    if (/[A-ZА-Я]/.test(password) && /[a-zа-я]/.test(password)) bits++;
    if (/\d/.test(password)) bits++;
    if (/[^A-Za-zА-Яа-я0-9]/.test(password) || password.length >= 12) bits++;
    return bits;
  }, [password]);

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

    setIsSubmitting(true);
    setTimeout(() => {
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
      setIsSubmitting(false);
    }, 650);
  };

  const handleSteamAuth = () => {
    setErrorMsg('');
    setIsSubmitting(true);
    setTimeout(() => {
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
      setIsSubmitting(false);
    }, 650);
  };

  const handleCopyJwt = () => {
    if (!token) return;
    navigator.clipboard?.writeText(token);
    setCopiedToken(true);
    setTimeout(() => setCopiedToken(false), 1800);
  };

  const hasActiveSession = isAuthenticated && currentUser && token;

  return (
    <div className="h-screen w-screen bg-[#0e0e0e] text-[#ededed] overflow-hidden selection:bg-white selection:text-black select-none">
      <div className="w-full h-full bg-[#0e0e0e] grid grid-cols-1 lg:grid-cols-12 overflow-hidden">
        {/* ЛЕВАЯ КОЛОНКА: УВЕЛИЧЕННАЯ НА 25% ФОРМА АВТОРИЗАЦИИ */}
        <div className="lg:col-span-5 flex flex-col justify-between px-8 py-9 sm:px-14 sm:py-11 lg:px-16 lg:py-12 bg-[#0e0e0e] z-10 border-b lg:border-b-0 lg:border-r border-[#1f1f1f] overflow-y-auto">
          {/* Верхний блок с укрупненным логотипом и переключателем */}
          <div className="flex items-center justify-between gap-4">
            <div
              onClick={() => navigate('/events')}
              className="flex items-center gap-4 cursor-pointer"
            >
              <HybridLoopLogo width={74} height={52} />
              <div className="flex flex-col">
                <span className="text-[26px] font-bold tracking-tight text-white leading-none">
                  squad bit
                </span>
                <span className="text-[12.5px] text-[#737373] tracking-[0.18em] uppercase mt-2 font-semibold">
                  TACTICAL TOURNAMENTS
                </span>
              </div>
            </div>

            {!hasActiveSession && (
              <div className="flex items-center bg-[#141414] border border-[#282828] p-1.5 rounded-[4px]">
                <button
                  type="button"
                  onClick={() => {
                    setMode('register');
                    setErrorMsg('');
                  }}
                  className={`px-4 py-2 text-[14.5px] font-bold transition-colors rounded-[3px] cursor-pointer ${
                    mode === 'register'
                      ? 'bg-[#2a2a2a] text-white'
                      : 'text-[#7c7c7c] hover:text-[#d4d4d4]'
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
                  className={`px-4 py-2 text-[14.5px] font-bold transition-colors rounded-[3px] cursor-pointer ${
                    mode === 'login'
                      ? 'bg-[#2a2a2a] text-white'
                      : 'text-[#7c7c7c] hover:text-[#d4d4d4]'
                  }`}
                >
                  Вход
                </button>
              </div>
            )}
          </div>

          {/* Центральный блок: форма увеличена с 420px до 530px (+25%) */}
          <div className="my-auto py-8 w-full max-w-[530px] mx-auto">
            {isSubmitting ? (
              <div className="py-20 flex flex-col items-center justify-center gap-5 bg-[#121212] border border-[#262626] rounded-[4px]">
                <Loading />
                <span className="text-[15px] font-bold text-[#a8a8a8] uppercase tracking-wider">
                  СИНХРОНИЗАЦИЯ ПРОФИЛЯ ОПЕРАТОРА...
                </span>
              </div>
            ) : !hasActiveSession ? (
              <>
                {errorMsg && (
                  <div className="mb-6 px-5 py-3.5 bg-[#1c1111] border border-[#442222] text-[#f08080] text-[15px] font-semibold rounded-[4px]">
                    {errorMsg}
                  </div>
                )}

                <form onSubmit={handleFormSubmit} className="space-y-5">
                  <div>
                    <label className="block text-[16px] font-bold text-[#e5e5e5] mb-2.5">
                      Email
                    </label>
                    <div className="relative">
                      <Mail className="w-5 h-5 text-[#737373] absolute left-4 top-1/2 -translate-y-1/2 pointer-events-none" />
                      <input
                        type="email"
                        required
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        placeholder="hello@squadbit.gg"
                        className="w-full h-14 pl-12 pr-4 bg-[#161616] border border-[#2a2a2a] focus:border-[#737373] focus:bg-[#1a1a1a] focus:outline-none text-[16px] font-medium text-white placeholder:text-[#5e5e5e] rounded-[4px] transition-colors"
                      />
                    </div>
                  </div>

                  {mode === 'register' && (
                    <div>
                      <label className="block text-[16px] font-bold text-[#e5e5e5] mb-2.5">
                        Имя профиля
                      </label>
                      <div className="relative">
                        <User className="w-5 h-5 text-[#737373] absolute left-4 top-1/2 -translate-y-1/2 pointer-events-none" />
                        <input
                          type="text"
                          required
                          value={username}
                          onChange={(e) => setUsername(e.target.value)}
                          placeholder="Ваш игровой никнейм"
                          className="w-full h-14 pl-12 pr-4 bg-[#161616] border border-[#2a2a2a] focus:border-[#737373] focus:bg-[#1a1a1a] focus:outline-none text-[16px] font-medium text-white placeholder:text-[#5e5e5e] rounded-[4px] transition-colors"
                        />
                      </div>
                    </div>
                  )}

                  <div>
                    <label className="block text-[16px] font-bold text-[#e5e5e5] mb-2.5">
                      Пароль
                    </label>
                    <div className="relative">
                      <Lock className="w-5 h-5 text-[#737373] absolute left-4 top-1/2 -translate-y-1/2 pointer-events-none" />
                      <input
                        type={showPassword ? 'text' : 'password'}
                        required
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        placeholder="Установите пароль профиля"
                        className="w-full h-14 pl-12 pr-12 bg-[#161616] border border-[#2a2a2a] focus:border-[#737373] focus:bg-[#1a1a1a] focus:outline-none text-[16px] font-medium text-white placeholder:text-[#5e5e5e] rounded-[4px] transition-colors"
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="absolute right-4 top-1/2 -translate-y-1/2 text-[#737373] hover:text-[#dedede] transition-colors cursor-pointer"
                        tabIndex={-1}
                      >
                        {showPassword ? (
                          <EyeClosed className="w-5 h-5" />
                        ) : (
                          <Eye className="w-5 h-5" />
                        )}
                      </button>
                    </div>

                    {mode === 'register' && password.length > 0 && (
                      <div className="mt-3">
                        <div className="grid grid-cols-4 gap-2 h-[5px]">
                          {[1, 2, 3, 4].map((segment) => (
                            <div
                              key={segment}
                              className={`h-full rounded-[2px] transition-colors ${
                                segment <= securityBits ? 'bg-white' : 'bg-[#262626]'
                              }`}
                            />
                          ))}
                        </div>
                      </div>
                    )}
                  </div>

                  {mode === 'register' && (
                    <div>
                      <div className="flex items-center justify-between mb-2.5">
                        <label className="block text-[16px] font-bold text-[#e5e5e5]">
                          Подтверждение пароля
                        </label>
                        {confirmPassword.length > 0 && password === confirmPassword && (
                          <span className="text-[13px] font-semibold text-[#b8b8b8] flex items-center gap-1.5">
                            <CheckCircle className="w-4 h-4 text-white" /> Совпадает
                          </span>
                        )}
                      </div>
                      <div className="relative">
                        <Lock className="w-5 h-5 text-[#737373] absolute left-4 top-1/2 -translate-y-1/2 pointer-events-none" />
                        <input
                          type={showPassword ? 'text' : 'password'}
                          required
                          value={confirmPassword}
                          onChange={(e) => setConfirmPassword(e.target.value)}
                          placeholder="Повторите пароль"
                          className="w-full h-14 pl-12 pr-4 bg-[#161616] border border-[#2a2a2a] focus:border-[#737373] focus:bg-[#1a1a1a] focus:outline-none text-[16px] font-medium text-white placeholder:text-[#5e5e5e] rounded-[4px] transition-colors"
                        />
                      </div>
                    </div>
                  )}

                  <div className="pt-2.5">
                    <button
                      type="submit"
                      className="w-full h-14 bg-white hover:bg-[#e6e6e6] active:bg-[#d4d4d4] text-[#0a0a0a] text-[16.5px] font-bold rounded-[4px] transition-colors cursor-pointer"
                    >
                      {mode === 'register' ? 'Зарегистрироваться' : 'Войти'}
                    </button>
                  </div>
                </form>

                <div className="mt-5 text-center">
                  {mode === 'register' ? (
                    <p className="text-[15px] text-[#808080]">
                      Уже есть аккаунт?{' '}
                      <button
                        type="button"
                        onClick={() => {
                          setMode('login');
                          setErrorMsg('');
                        }}
                        className="text-[#cccccc] font-semibold hover:text-white underline underline-offset-4 transition-colors cursor-pointer"
                      >
                        Войти
                      </button>
                    </p>
                  ) : (
                    <p className="text-[15px] text-[#808080]">
                      Нет аккаунта?{' '}
                      <button
                        type="button"
                        onClick={() => {
                          setMode('register');
                          setErrorMsg('');
                        }}
                        className="text-[#cccccc] font-semibold hover:text-white underline underline-offset-4 transition-colors cursor-pointer"
                      >
                        Регистрация
                      </button>
                    </p>
                  )}
                </div>

                <div className="relative my-7 flex items-center justify-center">
                  <div className="w-full border-t border-[#222222]" />
                  <span className="px-4 bg-[#0e0e0e] text-[14px] font-semibold text-[#666666]">
                    или
                  </span>
                  <div className="w-full border-t border-[#222222]" />
                </div>

                <div>
                  <button
                    type="button"
                    onClick={handleSteamAuth}
                    className="w-full h-14 bg-[#121212] hover:bg-[#1b1b1b] border border-[#2c2c2c] hover:border-[#444444] text-white text-[16px] font-bold rounded-[4px] flex items-center justify-center gap-3.5 transition-colors cursor-pointer"
                  >
                    <SteamIcon className="w-5 h-5 text-white" />
                    <span>
                      {mode === 'register' ? 'Регистрация через Steam' : 'Войти через Steam'}
                    </span>
                  </button>

                  <p className="mt-3.5 text-[14px] text-[#7c7c7c] text-center leading-relaxed">
                    При регистрации через стим, будет выдан рейтинг в зависимости от опыта в игре
                  </p>
                </div>
              </>
            ) : (
              <div className="bg-[#121212] border border-[#282828] p-7 rounded-[4px] space-y-6">
                <div className="flex items-center justify-between border-b border-[#222222] pb-4">
                  <div className="flex items-center gap-3">
                    <ShieldCheck className="w-5 h-5 text-white" />
                    <span className="text-[16px] font-bold text-white">
                      UserPrincipal инициализирован
                    </span>
                  </div>
                  <span className="text-[13px] font-bold px-2.5 py-1 bg-[#1d1d1d] border border-[#303030] text-[#cccccc] rounded-[3px]">
                    200 OK
                  </span>
                </div>

                <div className="space-y-3 text-[14.5px]">
                  <div className="flex justify-between">
                    <span className="text-[#777777]">id (UUID):</span>
                    <span className="text-[#dedede] font-medium">
                      {currentUser.id.slice(0, 18)}...
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[#777777]">username:</span>
                    <span className="text-white font-bold">
                      {currentUser.username}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[#777777]">role:</span>
                    <span className="text-[#cccccc] font-semibold">{currentUser.role}</span>
                  </div>
                  {currentUser.steamId && (
                    <div className="flex justify-between">
                      <span className="text-[#777777]">steamId:</span>
                      <span className="text-[#dedede]">{currentUser.steamId}</span>
                    </div>
                  )}
                  <div className="flex justify-between pt-3 border-t border-[#1f1f1f]">
                    <span className="text-[#777777]">Начальный рейтинг:</span>
                    <span className="text-white font-bold flex items-center gap-2">
                      <Star className="w-4 h-4 text-[#cccccc]" />
                      {currentUser.rating} ELO
                      {currentUser.squadHours
                        ? ` (${currentUser.squadHours}ч в Squad)`
                        : ''}
                    </span>
                  </div>
                </div>

                <div>
                  <div className="flex items-center justify-between text-[13px] text-[#777777] mb-2">
                    <span className="flex items-center gap-1.5 font-semibold">
                      <Terminal className="w-4 h-4" /> Authorization: Bearer
                    </span>
                    <button
                      type="button"
                      onClick={handleCopyJwt}
                      className="text-[#aaaaaa] hover:text-white flex items-center gap-1.5 transition-colors cursor-pointer font-semibold"
                    >
                      {copiedToken ? (
                        <>
                          <Check className="w-4 h-4" /> Скопировано
                        </>
                      ) : (
                        <>
                          <Copy className="w-4 h-4" /> Копировать
                        </>
                      )}
                    </button>
                  </div>
                  <div className="p-3.5 bg-[#0b0b0b] border border-[#222222] text-[13px] text-[#999999] break-all leading-relaxed rounded-[3px]">
                    {token}
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3.5 pt-2">
                  <button
                    type="button"
                    onClick={() => navigate('/events')}
                    className="w-full h-12 bg-white hover:bg-[#e6e6e6] text-[#0a0a0a] text-[15px] font-bold rounded-[4px] transition-colors cursor-pointer"
                  >
                    На главную
                  </button>
                  <button
                    type="button"
                    onClick={clearSession}
                    className="w-full h-12 bg-[#1a1a1a] hover:bg-[#242424] border border-[#2e2e2e] text-[15px] text-white font-bold flex items-center justify-center gap-2 rounded-[4px] transition-colors cursor-pointer"
                  >
                    <Refresh className="w-4 h-4" />
                    <span>Выйти</span>
                  </button>
                </div>
              </div>
            )}
          </div>

          <div className="pt-5 border-t border-[#1a1a1a] flex items-center justify-between text-[13px] text-[#737373]">
            <div className="flex items-center gap-2.5">
              <span className="w-2 h-2 rounded-full bg-emerald-500" />
              <Group className="w-4 h-4 text-[#888888]" />
              <span>
                ОПЕРАТОРОВ В СЕТИ: <strong className="text-[#dedede] font-bold">1,420</strong>
              </span>
            </div>
            <span>
              ВСЕГО ПРОФИЛЕЙ: <strong className="text-[#cccccc] font-bold">18,940</strong>
            </span>
          </div>
        </div>

        {/* ПРАВАЯ КОЛОНКА: DOT-MATRIX БОЕЦ */}
        <div className="lg:col-span-7 relative min-h-[440px] lg:min-h-full bg-[#0e0e0e]">
          <SoldierCanvas />
        </div>
      </div>
    </div>
  );
}

export default AuthPage;
