// Backend modelleriyle birebir ortusen TypeScript tipleri

export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
}

export interface TourImage {
  id: string;
  tour_id: string;
  image_url: string;
  sort_order: number;
}

export interface Departure {
  id: string;
  tour_id: string;
  start_date: string;
  end_date: string;
  capacity: number;
  filled: number;
  price: number;
  available?: number;
}

export interface TourItem {
  id: string;
  tour_id: string;
  description: string;
}

export interface ItineraryDay {
  id: string;
  tour_id: string;
  day_no: number;
  title: string;
  description: string;
}

export interface Tour {
  id: string;
  title: string;
  slug: string;
  description: string;
  category_id: string;
  category_name?: string;
  cover_image: string;
  duration_days: number;
  duration_nights: number;
  location: string;
  base_price: number;
  currency: string;
  active: boolean;
  featured: boolean;
  publish_order?: number;
  images?: TourImage[];
  departures?: Departure[];
  included?: TourItem[];
  excluded?: TourItem[];
  bring_items?: TourItem[];
  itinerary?: ItineraryDay[];
}

export interface Traveler {
  id: string;
  booking_id: string;
  full_name: string;
  birth_date: string;
  id_last4?: string;
  is_child: boolean;
}

export interface Booking {
  id: string;
  tour_id: string;
  tour_title?: string;
  departure_id: string;
  user_id: string;
  adult_count: number;
  child_count: number;
  total_price: number;
  status: "pending" | "paid" | "confirmed" | "cancelled" | "failed";
  created_at: string;
  travelers?: Traveler[];
}

export interface User {
  id: string;
  full_name: string;
  email: string;
  phone?: string;
  role: "customer" | "admin";
  created_at?: string;
}

export interface Payment {
  id: string;
  booking_id: string;
  iyzico_payment_id: string;
  amount: number;
  status: string;
  installment: number;
  created_at: string;
}

export interface TourListResponse {
  tours: Tour[];
  total: number;
  page: number;
  limit: number;
}

export interface DashboardStats {
  total_bookings: number;
  pending_payments: number;
  total_revenue: number;
  total_customers: number;
}

export interface SalesPoint {
  date: string;
  total: number;
}

export interface OccupancyRow {
  tour: string;
  start_date: string;
  capacity: number;
  filled: number;
}

export interface Coupon {
  id: string;
  code: string;
  discount_type: "percent" | "fixed";
  discount_value: number;
  valid_from?: string;
  valid_until?: string;
  active: boolean;
  created_at: string;
}
