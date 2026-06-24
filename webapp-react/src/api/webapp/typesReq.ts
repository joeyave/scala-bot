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

export interface ReqQueryParamsUpdateEvent {
  messageId: string;
  chatId: string;
  userId: string;
}

export interface SongOverride {
  songId: string;
  eventKey: string;
}

export interface ReqBodyUpdateEvent {
  name?: string;
  date?: string;
  timezone?: string;
  songIds?: string[];
  songOverrides?: SongOverride[];
  notes?: string;
}

export interface ReqBodySettingsBand {
  name: string;
  driveFolderId?: string;
  driveFolderUrl?: string;
  timezone: string;
}

export interface ReqBodySettingsBandPatch {
  name?: string;
  driveFolderId?: string;
  driveFolderUrl?: string;
  timezone?: string;
}

export interface ReqBodySettingsMemberRole {
  isAdmin: boolean;
}
