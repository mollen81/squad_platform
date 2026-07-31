import { Routes, Route, Link } from 'react-router-dom';
import { ClanCard } from './features/clans/ClanCard';

const Dashboard = () => <div className="text-white p-6">Главный Дашборд и ELO Лидерборд</div>;
const Profile = () => <div className="text-white p-6">Профиль игрока</div>;
const EventManagement = () => <div className="text-white p-6">Матчмейкинг и Ивенты</div>;

// Страница с просмотром тестового клана
const ClanHub = () => (
    <div className="p-6">
        <ClanCard clanId="test-clan-id-123" currentUserId="76561198000000000" />
    </div>
);

function App() {
    return (
        <div className="min-h-screen bg-slate-900 text-slate-100 font-sans">
            <nav className="flex gap-6 p-4 bg-slate-800 border-b border-slate-700">
                <Link to="/" className="hover:text-blue-400 font-semibold">Дашборд</Link>
                <Link to="/clans" className="hover:text-blue-400 font-semibold">Кланы</Link>
                <Link to="/events" className="hover:text-blue-400 font-semibold">Ивенты</Link>
                <Link to="/profile/76561198000000000" className="hover:text-blue-400 font-semibold ml-auto">Профиль</Link>
            </nav>

            <main className="container mx-auto mt-4">
                <Routes>
                    <Route path="/" element={<Dashboard />} />
                    <Route path="/profile/:steamId" element={<Profile />} />
                    <Route path="/clans" element={<ClanHub />} />
                    <Route path="/events" element={<EventManagement />} />
                </Routes>
            </main>
        </div>
    );
}

export default App;