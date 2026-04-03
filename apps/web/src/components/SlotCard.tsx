import { Link } from 'react-router-dom';
import { format, parseISO } from 'date-fns';
import { ru } from 'date-fns/locale';
import { Slot } from '../api/types';

interface SlotCardProps {
  slot: Slot;
}

export function SlotCard({ slot }: SlotCardProps) {
  const formattedDate = slot.starts_at 
    ? format(parseISO(slot.starts_at), "d MMMM yyyy, HH:mm", { locale: ru })
    : "Дата не указана";

  return (
    <div className="bg-white border-4 border-black p-6 flex flex-col h-full shadow-[6px_6px_0px_0px_rgba(0,0,0,1)] hover:-translate-y-1 hover:shadow-[10px_10px_0px_0px_rgba(0,0,0,1)] transition-all relative overflow-hidden group selection:bg-lime-300">
      <div className="flex justify-between items-start mb-4 relative z-10">
        <span className="inline-block bg-lime-400 border-2 border-black text-black text-xs px-3 py-1 font-black uppercase tracking-widest shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
          {slot.sport}
        </span>
        <span className="text-black text-sm font-bold bg-white border-b-2 border-black inline-block px-1">
          {slot.district}
        </span>
      </div>

      <h3 className="text-2xl font-black text-black mb-2 uppercase tracking-tight leading-none leading-tight">{slot.venue_name}</h3>
      <p className="text-gray-800 text-sm mb-6 line-clamp-2 font-mono" title={slot.address}>
        {slot.address}
      </p>

      <div className="space-y-4 mb-8 flex-grow font-mono text-sm font-bold">
        <div className="flex items-center text-black border-l-4 border-lime-400 pl-3">
          <svg className="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="square" strokeLinejoin="miter" strokeWidth="2.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          {formattedDate}
        </div>
        <div className="flex items-center text-black border-l-4 border-lime-400 pl-3">
          <svg className="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="square" strokeLinejoin="miter" strokeWidth="2.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {slot.duration_minutes} МИН
        </div>
        <div className="flex items-center justify-between mt-4 pt-4 border-t-2 border-dashed border-black">
          <span>ОПЛАТА:</span>
          <span className="font-black text-lg bg-lime-400 px-2 py-0.5 border-2 border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">{slot.expected_price} ₽</span>
        </div>
        <div className="flex items-center justify-between">
          <span>СВОБОДНЫЕ:</span>
          <span className="font-black text-black text-lg">{slot.free_spots ?? '-'} / {slot.capacity}</span>
        </div>
      </div>

      <Link 
        to={`/slots/${slot.id}`} 
        className="block w-full text-center bg-transparent hover:bg-black text-black hover:text-white font-black py-3 px-4 border-4 border-black transition-colors uppercase tracking-widest text-lg font-mono relative z-10"
      >
        ПОДРОБНЕЕ
      </Link>
      
      <div className="absolute -bottom-6 -right-6 text-9xl text-gray-100 font-black opacity-30 pointer-events-none group-hover:scale-110 transition-transform duration-500 z-0">
        #
      </div>
    </div>
  );
}
