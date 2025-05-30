import { Page } from "@/components/Page.tsx";
import { logger } from "@/helpers/logger";
import { useInitParams, useSongMutation } from "@/pages/SongPage/SongPage.tsx";
import { StateSongData } from "@/pages/SongPage/util/types.ts";
import { useQueryClient } from "@tanstack/react-query";
import {
  backButton,
  mainButton,
  miniApp,
  themeParams,
  viewport,
  viewportWidth,
} from "@telegram-apps/sdk-react";
import {
  Headline,
  List,
  Placeholder,
  Select,
  Spinner,
} from "@telegram-apps/telegram-ui";
import { Notify } from "notiflix";
import { FC, useCallback, useEffect, useRef, useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { Document as DocumentPDF, Page as PagePDF, pdfjs } from "react-pdf";
import { useLocation } from "react-router";

pdfjs.GlobalWorkerOptions.workerSrc = new URL(
  "pdfjs-dist/build/pdf.worker.min.mjs",
  import.meta.url,
).toString();

const SongConfirmationPage: FC = () => {
  const { songId, messageId, chatId, userId } = useInitParams();

  if (!songId || !messageId || !chatId || !userId) {
    throw new Error("Failed to get song page: invalid request params.");
  }

  const { t } = useTranslation();

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  const { state } = useLocation();
  const st = state as StateSongData;

  const [sectionNumber, setSectionNumber] = useState<string>("-1");

  const queryClient = useQueryClient();
  const mutateSong = useSongMutation();

  // const queryPdf = useQuery({
  //   queryKey: ["songPdf", songId],
  //   queryFn: async () => {
  //     if (!songId) {
  //       throw new Error("Failed to get song data: invalid request params.");
  //     }
  //     const data = await downloadSongPdf(songId);
  //     if (!data) {
  //       throw new Error("Song pdf is empty.");
  //     }
  //     return data;
  //   },
  // });

  // const pdfObjectUrl = useMemo(() => {
  //   if (queryPdf.data) {
  //     return URL.createObjectURL(queryPdf.data);
  //   }
  //   return null;
  // }, [queryPdf.data]);

  // useEffect(() => {
  //   return () => {
  //     if (pdfObjectUrl) {
  //       URL.revokeObjectURL(pdfObjectUrl);
  //     }
  //   };
  // }, [pdfObjectUrl]);

  const [numPages, setNumPages] = useState<number>();

  function onDocumentLoadSuccess({ numPages }: { numPages: number }): void {
    setNumPages(numPages);
  }

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
          Notify.failure(t("saveError"));
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
    t,
  ]);

  useEffect(() => {
    mainButton.setParams({
      isVisible: true,
      isEnabled: true,
      text: t("save"),
    });

    const handleMainButtonClickSync = () => {
      void handleMainButtonClick();
    };

    mainButton.onClick(handleMainButtonClickSync);

    return () => {
      mainButton.offClick(handleMainButtonClickSync);
    };
  }, [handleMainButtonClick, t]);

  function calcPdfWidth(width: number) {
    return width - 83 + 15 - 32;
  }

  const [pdfWidth, setPdfWidth] = useState(calcPdfWidth(viewportWidth()));
  const initWidth = useRef<number>(pdfWidth);
  useEffect(() => {
    const set = function (newWidth: number) {
      if (newWidth < initWidth.current) {
        setPdfWidth(calcPdfWidth(newWidth));
      }
    };

    viewport.width.sub(set);

    return () => {
      viewport.width.unsub(set);
    };
  }, []);

  return (
    <Page back={true}>
      <div
        className="flex h-screen flex-col overflow-hidden"
        style={{
          // todo: check how to use safeAreaInset properly.
          marginTop: viewport.safeAreaInsetTop(),
          marginBottom: viewport.safeAreaInsetBottom(),
        }}
      >
        <div className="flex-none">
          <List>
            <Headline>
              <Trans
                i18nKey="changeKeyTitle"
                values={{
                  from: st.initialFormData.key,
                  to: st.formData.key,
                }}
                components={{
                  strong: <strong />,
                }}
              />
            </Headline>{" "}
            <Select
              onChange={(e) => {
                setSectionNumber(e.target.value);
              }}
            >
              <option key={"-1"} value={"-1"}>
                {t("atEnd")}
              </option>
              {Array(st.sectionsNumber)
                .fill(null)
                .map((_, i) => (
                  <option key={String(i)} value={String(i)}>
                    {t("insteadOfPage", { number: i + 1 })}
                  </option>
                ))}
              <option key={""} value={""}>
                {t("onlyFirstHeader")}
              </option>
            </Select>
          </List>
        </div>

        <div // flex to the bottom of the screen.
          className="flex min-h-0 flex-1 flex-col"
        >
          <List // padding to the sides.
            className="flex min-h-0 flex-1 flex-col items-center pt-0 pb-0"
          >
            <div // rounded corners.
              className="min-h-0 w-fit flex-1 rounded-lg"
              style={{
                backgroundColor: themeParams.sectionBackgroundColor(),
              }}
            >
              <div // inner padding.
                className="h-full p-4"
              >
                <div // pdf.
                  className="h-full overflow-y-auto"
                  style={{ scrollbarWidth: "thin" }}
                >
                  <DocumentPDF
                    className={
                      "h-full" +
                      (numPages ? "" : " flex flex-col justify-center")
                    }
                    loading={
                      <Placeholder
                        style={{
                          width: pdfWidth,
                        }}
                        header={t("loadingPdf")}
                        description={t("waitMsg")}
                      >
                        <Spinner size="l" />
                      </Placeholder>
                    }
                    error={
                      <Placeholder
                        style={{
                          width: pdfWidth,
                        }}
                        header={t("errorLoadingPdf")}
                        description=""
                      />
                    }
                    file={`${window.location.origin}/api/songs/${songId}/download`}
                    onLoadSuccess={onDocumentLoadSuccess}
                  >
                    {numPages &&
                      Array.from(
                        { length: numPages },
                        (_, index) => index + 1,
                      ).map((pageNumber, index) => (
                        <PagePDF
                          key={pageNumber}
                          className={index < numPages - 1 ? "pb-2" : undefined}
                          width={pdfWidth}
                          canvasBackground={themeParams.secondaryBackgroundColor()}
                          renderTextLayer={false}
                          renderAnnotationLayer={false}
                          pageNumber={pageNumber}
                        />
                      ))}
                  </DocumentPDF>
                </div>
              </div>
            </div>
          </List>
        </div>
      </div>
    </Page>
  );
};

export default SongConfirmationPage;