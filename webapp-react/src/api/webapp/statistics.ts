import { doReqWebappApi } from "@/api/webapp/doReq.ts";
import { RespStatistics } from "@/api/webapp/typesResp.ts";
import { isMockEnabled, mockStatisticsData } from "@/api/webapp/mockData.ts";

export async function getStatistics(
  bandId: string,
  fromDate?: string,
): Promise<RespStatistics | null> {
  if (isMockEnabled()) {
    console.info("📦 Using mock data for getStatistics");
    return Promise.resolve(mockStatisticsData);
  }

  const queryParams = fromDate ? { bandId, from: fromDate } : { bandId };

  const { data, err } = await doReqWebappApi<RespStatistics>(
    "/api/statistics",
    "GET",
    queryParams,
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data;
}
