import { MultiselectOption } from "@telegram-apps/telegram-ui/dist/components/Form/Multiselect/types";

export interface SongFormValues {
  name: string;
  key: string;
  bpm: string;
  time: string;
  tags: MultiselectOption[];
}

export interface SongStateData {
  songId: string;
  userId: string;
  chatId: string;
  messageId: string;
  formData: SongFormValues;
  initialFormData: SongFormValues;
  sectionsNumber: number;
}
