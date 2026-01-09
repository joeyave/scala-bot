export interface RespSongData {
  song: Song;
  bandTags: string[];
}

export interface RespSong {
  song: Song;
}

interface Song {
  id: string;
  driveFileId: string;
  bandId: string;
  band: Band;
  pdf: Pdf;
  tags: string[];
  isArchived: boolean;
}

interface Band {
  id: string;
  name: string;
  timezone: string;
  driveFolderId: string;
  archiveFolderId: string;
  roles?: Role[];
}

interface Pdf {
  modifiedTime: string;
  tgFileId: string;
  tgChannelMessageId: number;
  name: string;
  key: string;
  bpm: string;
  time: string;
  webViewLink: string;
}

export interface RespSongLyrics {
  lyricsHtml: string;
  sectionsNumber: number;
}

export interface RespTags {
  tags: string[];
}

export interface RespEventFreqNames {
  names: string[];
}

export interface RespEventData {
  event: Event;
}

// SetlistItem represents a song in the setlist with optional event-specific overrides
export interface SongOverride {
  songId: string;
  eventKey: string;
}

export interface Event {
  id: string;
  time: string;
  name: string;
  // memberships: any[] // todo
  bandId: string;
  band: Band;
  songIds: string[];
  songOverrides?: SongOverride[];
  songs: Song[];
  notes: string;
}

export interface Role {
  id: string;
  name: string;
  band_id: string;
  priority?: number;
}

export interface RespSearchDriveFiles {
  driveFiles: DriveFile[];
}

export interface DriveFile {
  id: string;
  name: string;
  modifiedTime: string;
  parents: string[];
  webViewLink: string;
}
