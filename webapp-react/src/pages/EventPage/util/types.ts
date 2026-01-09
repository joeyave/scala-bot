export interface EventForm {
  name: string;
  date: string;
  setlist: Song[];
  notes: string;
}

export interface Song {
  id: string;
  name: string;
  key: string;
  bpm: string;
  time: string;
  eventKey?: string; // Key override for this event (if different from original)
}
