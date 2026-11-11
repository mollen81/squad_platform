import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { TournamentsPage } from '@/pages/tournaments/TournamentsPage';
import { AuthPage } from '@/pages/auth/AuthPage';

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Главная страница открыта публично для всех гостей и игроков */}
        <Route path="/" element={<TournamentsPage />} />
        <Route path="/events" element={<TournamentsPage />} />

        {/* Страница авторизации и регистрации */}
        <Route path="/auth" element={<AuthPage />} />

        {/* Редирект на главную при неизвестных путях */}
        <Route path="*" element={<Navigate to="/events" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
