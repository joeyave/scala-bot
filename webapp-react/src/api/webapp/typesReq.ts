export interface ReqQueryParamsUpdateSong {
  messageId: string;
  chatId: string;
  userId: string;
}

export interface ReqBodyUpdateSong {
  name?: string;
  key?: string;
  bpm?: string;
  time?: string;
  tags?: string[];
  transposeSection?: string;
}
