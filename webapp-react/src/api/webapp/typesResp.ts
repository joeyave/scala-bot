export interface RespSongData {
  song: Song;
  bandTags: string[];
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
  driveFolderId: string;
  archiveFolderId: string;
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
