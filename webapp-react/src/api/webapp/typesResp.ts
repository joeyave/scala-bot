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
  version: number;
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
  metadataSyncWasUpdated: boolean;
  metadata: {
    name: string;
    key: string;
    bpm: string;
    time: string;
  };
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

export interface RespStatistics {
  bandName: string;
  currentDate: string;
  defaultFromDate: string;
  roles: StatisticsRole[];
  users: StatisticsUser[];
}

export interface StatisticsUser {
  id: number;
  name: string;
  events: StatisticsEvent[];
}

export interface StatisticsEvent {
  id: string;
  date: string;
  weekday: number;
  name: string;
  roles: StatisticsRole[];
}

export interface StatisticsRole {
  id: string;
  name: string;
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
  parents: string[];
  webViewLink: string;
}
