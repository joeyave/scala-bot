import { doReqWebappApi } from "@/api/webapp/doReq.ts";
import { ReqBodyUpdateSong, ReqQueryParamsUpdateSong } from "./typesReq.ts";
import { RespDataGetSong } from "@/api/webapp/typesResp.ts";

type Resp<T> = {
  data: T | null;
  err: Error | null;
};

export async function getSong(
  songId: string,
  userId: string,
): Promise<Resp<RespDataGetSong>> {
  const { data, err } = await doReqWebappApi<RespDataGetSong>(
    `/api/songs/${songId}`,
    "GET",
    { userId },
    { Accept: "application/json" },
  );

  return { data, err };
}

export async function updateSong(
  songId: string,
  queryParams: ReqQueryParamsUpdateSong,
  body: ReqBodyUpdateSong,
) {
  const { err } = await doReqWebappApi(
    `/web-app/songs/${songId}/edit/confirm`,
    "POST",
    queryParams,
    { Accept: "application/json" },
    body,
  );

  return err;
}
