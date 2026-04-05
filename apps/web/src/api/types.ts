export interface ApiError {
  status: number;
  code: string;
  message: string;
}

export type SlotStatus = 'OPEN' | 'CANCELLED' | 'COMPLETED';

export interface Slot {
  id: string;
  sport: string;
  district: string;
  venue_name: string;
  address: string;
  starts_at: string;
  deadline_at: string;
  duration_minutes: number;
  capacity: number;
  min_players: number;
  expected_price: number;
  max_price: number;
  rules_text?: string;
  status: SlotStatus;
  created_at: string;
  updated_at: string;
  
  // Дополнительные поля для UI
  current_participants?: number;
  free_spots?: number;
}

export type ParticipationStatus = 'RESERVED' | 'PAID';

export interface Participation {
  id: string;
  slot_id: string;
  user_id: string;
  status: ParticipationStatus;
  reserved_at: string;
  paid_at?: string;
}

// Комбинированные данные для страницы "Мои участия"
export interface MyParticipation extends Participation {
  sport: string;
  district: string;
  venue_name: string;
  address: string;
  starts_at: string;
  deadline_at: string;
  duration_minutes: number;
  expected_price: number;
  max_price: number;
  slot_status: string;
}

export type PaymentStatus = 'PENDING' | 'PAID' | 'FAILED' | 'REFUNDED';

export interface Payment {
  id: string;
  participant_id: string;
  idempotency_key: string;
  amount: number;
  currency: string;
  provider: string;
  status: PaymentStatus;
  provider_payment_id?: string;
  provider_metadata?: string;
  created_at: string;
  updated_at: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}
