import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import { SlotsPage } from './pages/SlotsPage';
import { SlotDetailsPage } from './pages/SlotDetailsPage';
import { MyParticipationsPage } from './pages/MyParticipationsPage';

function Home() {
  return (
    <div className="min-h-screen bg-[#F4F4F5] flex flex-col justify-center py-12 px-4 sm:px-6 lg:px-8 selection:bg-lime-300">
      <div className="max-w-md w-full space-y-12 mx-auto text-center">
        <div>
          <h2 className="mt-6 text-center text-8xl font-black text-black tracking-tighter uppercase leading-none">
            ZAL
          </h2>
          <p className="mt-4 text-center font-bold text-lg text-black uppercase tracking-widest bg-lime-400 inline-block px-2 border-2 border-black -rotate-2">
            Твой спорт. Твои правила.
          </p>
        </div>
        <div className="flex flex-col gap-6 mt-16 font-mono">
          <Link 
            to="/slots" 
            className="w-full flex justify-center py-5 px-6 border-4 border-black text-xl font-black text-black bg-lime-400 hover:bg-lime-500 uppercase tracking-widest shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] transition-all active:translate-x-[4px] active:translate-y-[4px] active:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]"
          >
            Найти Игру
          </Link>
          <Link 
            to="/me/participations" 
            className="w-full flex justify-center py-5 px-6 border-4 border-black text-xl font-black text-black bg-white hover:bg-gray-100 uppercase tracking-widest shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] transition-all active:translate-x-[4px] active:translate-y-[4px] active:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]"
          >
            Мои участия
          </Link>
        </div>
      </div>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/slots" element={<SlotsPage />} />
        <Route path="/slots/:id" element={<SlotDetailsPage />} />
        <Route path="/me/participations" element={<MyParticipationsPage />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
