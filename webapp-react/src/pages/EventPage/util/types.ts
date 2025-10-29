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
}
