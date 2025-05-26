export interface SongForm {
  name: string;
  key: string;
  bpm: string;
  time: string;
  tags: string[];
}

export interface StateSongData {
  formData: SongForm;
  initialFormData: SongForm;
  sectionsNumber: number;
}
