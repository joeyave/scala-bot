import { downloadSongPdf } from "@/api/webapp/songs.ts";
import { Page } from "@/components/Page.tsx";
import { logger } from "@/helpers/logger";
import { useInitParams, useSongMutation } from "@/pages/SongPage/SongPage.tsx";
import { StateSongData } from "@/pages/SongPage/util/types.ts";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { backButton, mainButton, miniApp } from "@telegram-apps/sdk-react";
import { Headline, List, Select, Text } from "@telegram-apps/telegram-ui";
import { Notify } from "notiflix";
import { FC, useCallback, useEffect, useMemo, useState } from "react";
import { useLocation } from "react-router";

export const SongConfirmationPage: FC = () => {
  const { songId, messageId, chatId, userId } = useInitParams();

  if (!songId || !messageId || !chatId || !userId) {
    throw new Error("Failed to get song page: invalid request params.");
  }

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  const { state } = useLocation();
  const st = state as StateSongData;

  const [sectionNumber, setSectionNumber] = useState<string>("-1");

  const queryClient = useQueryClient();
  const mutateSong = useSongMutation();

  const queryPdf = useQuery({
    queryKey: ["songPdf", songId],
    queryFn: async () => {
      if (!songId) {
        throw new Error("Failed to get song data: invalid request params.");
      }
      const data = await downloadSongPdf(songId);
      if (!data) {
        throw new Error("Song pdf is empty.");
      }
      return data;
    },
  });

  // Inside your component:
  const pdfObjectUrl = useMemo(() => {
    if (queryPdf.data) {
      return URL.createObjectURL(queryPdf.data);
    }
    return null;
  }, [queryPdf.data]);

  useEffect(() => {
    return () => {
      if (pdfObjectUrl) {
        URL.revokeObjectURL(pdfObjectUrl);
      }
    };
  }, [pdfObjectUrl]);

  const handleMainButtonClick = useCallback(async () => {
    backButton.hide();
    mainButton.setParams({ isLoaderVisible: true });

    await mutateSong(
      {
        songId: songId,
        name: st.formData.name,
        key: st.formData.key,
        bpm: st.formData.bpm,
        time: st.formData.time,
        tags: st.formData.tags,
        transposeSection: sectionNumber,
        messageId: messageId,
        chatId: chatId,
        userId: userId,
      },
      {
        onSuccess: () => {
          miniApp.close();
        },
        onError: (err) => {
          void queryClient.invalidateQueries({
            queryKey: ["songData", songId, userId],
          });
          void queryClient.invalidateQueries({
            queryKey: ["songLyrics", songId],
          });

          logger.error("Failed to save song", { error: err });
          Notify.failure("Failed to save song. Please try again later.");
          mainButton.setParams({ isLoaderVisible: false });
          backButton.show();
        },
      },
    );
  }, [
    mutateSong,
    songId,
    st.formData.name,
    st.formData.key,
    st.formData.bpm,
    st.formData.time,
    st.formData.tags,
    sectionNumber,
    messageId,
    chatId,
    userId,
    queryClient,
  ]);

  useEffect(() => {
    mainButton.setParams({
      isVisible: true,
      isEnabled: true,
      text: "Save Changes",
    });

    const handleMainButtonClickSync = () => {
      void handleMainButtonClick();
    };

    mainButton.onClick(handleMainButtonClickSync);

    return () => {
      mainButton.offClick(handleMainButtonClickSync);
    };
  }, [handleMainButtonClick]);

  return (
    <Page back={true}>
      <div className="flex h-screen flex-col">
        <div className="flex-none">
          <List>
            <Headline>
              You are changing the key from {st.initialFormData.key} to{" "}
              {st.formData.key}. Where should page with new key be placed?
            </Headline>
            <Select
              onChange={(e) => {
                setSectionNumber(e.target.value);
              }}
            >
              <option key={"-1"} value={"-1"}>
                At the end of the doc
              </option>
              {Array(st.sectionsNumber)
                .fill(null)
                .map((_, i) => (
                  <option key={String(i)} value={String(i)}>
                    Instead of {i + 1} page
                  </option>
                ))}
              <option key={""} value={""}>
                Only first header
              </option>
            </Select>
          </List>
        </div>
        <div className="flex-1">
          <List className="h-full pt-0 pb-0">
            <div className="h-full p-2">
              {
                // todo: make more readable.
                queryPdf.isLoading || !queryPdf.isFetchedAfterMount ? ( // todo: test.
                  <Text>Loading PDF...</Text>
                ) : queryPdf.isSuccess ? (
                  <iframe
                    className="h-full w-full border-0"
                    src={pdfObjectUrl || undefined}
                  ></iframe>
                ) : (
                  <Text>Error loading PDF :(</Text>
                )
              }
            </div>
          </List>
        </div>
      </div>
    </Page>
  );
};
