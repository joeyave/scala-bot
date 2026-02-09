import { doReqWebappApi } from "@/api/webapp/doReq.ts";
import {
  RespSong,
  RespSongData,
  RespSongLyrics,
} from "@/api/webapp/typesResp.ts";
import { ReqBodyUpdateSong, ReqQueryParamsUpdateSong } from "./typesReq.ts";

export async function getSongData(
  songId: string,
  userId: string,
): Promise<RespSongData | null> {
  const { data, err } = await doReqWebappApi<RespSongData>(
    `/api/songs/${songId}`,
    "GET",
    { userId },
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data;
}

export async function getSongByDriveFileId(
  driveFileId: string,
): Promise<RespSong | null> {
  const { data, err } = await doReqWebappApi<RespSong>(
    `/api/v2/songs/find-by-drive-file-id`,
    "GET",
    { driveFileId: driveFileId },
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data;
}

export async function getSongLyrics(
  songId: string,
): Promise<RespSongLyrics | null> {
  const { data, err } = await doReqWebappApi<RespSongLyrics>(
    `/api/songs/${songId}/lyrics`,
    "GET",
    undefined,
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data;
}

export async function updateSong(
  songId: string,
  queryParams: ReqQueryParamsUpdateSong,
  body: ReqBodyUpdateSong,
) {
  const { err } = await doReqWebappApi(
    `/api/songs/${songId}/edit`,
    "POST",
    queryParams,
    { Accept: "application/json" },
    body,
  );

  if (err) {
    throw err;
  }

  return;
}

export async function formatSong(
  songId: string,
  queryParams: ReqQueryParamsUpdateSong,
) {
  const { err } = await doReqWebappApi(
    `/api/songs/${songId}/format`,
    "POST",
    queryParams,
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return;
}

export async function downloadSongPdf(songId: string): Promise<Blob | null> {
  const { data, err } = await doReqWebappApi<Blob>(
    `/api/songs/${songId}/download`,
    "GET",
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data;
}
