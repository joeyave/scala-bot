import { getSongData, getSongLyrics, updateSong } from "@/api/webapp/songs.ts";
import { ReqBodyUpdateSong } from "@/api/webapp/typesReq.ts";
import { BPMInput, formatBpm } from "@/components/BPMInput/BPMInput.tsx";
import {
  EditableTitle,
  formatTitle,
} from "@/components/EditableTitle/EditableTitle.tsx";
import { formatKey, KeyInput } from "@/components/KeyInput/KeyInput.tsx";
import { Page } from "@/components/Page.tsx";
import { TagsInput } from "@/components/TagsInput/TagsInput.tsx";
import {
  formatTimeSignature,
  TimeSignatureInput,
} from "@/components/TimeSignatureInput/TimeSignatureInput.tsx";
import { logger } from "@/helpers/logger";
import { setMainButton } from "@/helpers/mainButton.ts";
import {
  isBpmValid,
  isFormChanged,
  isFormValid,
  isNameValid,
  isTimeSignatureValid,
} from "@/pages/SongPage/util/formValidation.ts";
import { transposeAllText } from "@/pages/SongPage/util/transpose.ts";
import { SongForm, StateSongData } from "@/pages/SongPage/util/types.ts";
import { PageError } from "@/pages/UtilPages/PageError.tsx";
import { PageLoading } from "@/pages/UtilPages/PageLoading.tsx";
import { useMutation, useQuery } from "@tanstack/react-query";
import {
  mainButton,
  miniApp,
  postEvent,
  themeParams,
  viewport,
} from "@telegram-apps/sdk-react";
import { Button, List, Section, Text } from "@telegram-apps/telegram-ui";
import { MultiselectOption } from "@telegram-apps/telegram-ui/dist/components/Form/Multiselect/types";
import { Notify } from "notiflix";
import { FC, useCallback, useEffect, useState } from "react";
import { FileEarmarkTextFill } from "react-bootstrap-icons";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams, useSearchParams } from "react-router";

interface SongMutationData {
  songId: string;
  name: string;
  key: string;
  bpm: string;
  time: string;
  tags: string[];
  transposeSection?: string;
  messageId: string;
  chatId: string;
  userId: string;
}

export function useSongMutation() {
  const { mutateAsync: mutateSong } = useMutation({
    mutationFn: async (d: SongMutationData) => {
      const queryParams = {
        messageId: d.messageId,
        chatId: d.chatId,
        userId: d.userId,
      };

      const body: ReqBodyUpdateSong = {
        name: d.name,
        key: d.key || "?", // todo: refactor, remove question marks.
        bpm: d.bpm || "?",
        time: d.time || "?",
        tags: d.tags,
        transposeSection: d.transposeSection,
      };

      return await updateSong(d.songId, queryParams, body);
    },
  });
  return mutateSong;
}

export const SongPage: FC = () => {
  const { t } = useTranslation();

  const { songId, messageId, chatId, userId } = useInitParams();

  if (!songId || !messageId || !chatId || !userId) {
    throw new Error("Failed to get song page: invalid request params.");
  }

  console.log(themeParams.backgroundColor());

  const navigate = useNavigate();

  // State to track initial values.
  const [initialFormData, setInitialFormData] = useState<SongForm>({
    name: "",
    key: "",
    bpm: "",
    time: "",
    tags: [],
  });

  // Form data state.
  const [formData, setFormData] = useState<SongForm>({
    name: "",
    key: "",
    bpm: "",
    time: "",
    tags: [],
  });

  const [transposedLyricsHtml, setTransposedLyricsHtml] = useState<string>("");
  const [transpositionError, setTranspositionError] = useState<boolean>(false);

  useEffect(() => {
    postEvent("web_app_expand");
    postEvent("web_app_setup_swipe_behavior", { allow_vertical_swipe: false });
  }, []);

  const querySongDataRes = useQuery({
    queryKey: ["songData", songId, userId],
    queryFn: async () => {
      if (!songId || !userId) {
        throw new Error("Failed to get song data: invalid request params.");
      }
      const data = await getSongData(songId, userId);
      if (!data?.song) {
        throw new Error("Song data is empty.");
      }
      return data;
    },
  });

  const querySongLyricsRes = useQuery({
    queryKey: ["songLyrics", songId],
    queryFn: async () => {
      if (!songId) {
        throw new Error("Failed to get song data: invalid request params.");
      }
      const data = await getSongLyrics(songId);
      if (!data) {
        throw new Error("Song lyrics is empty.");
      }
      return data;
    },
  });

  const mutateSong = useSongMutation();

  // Set form data after fetches completed or updated.
  useEffect(() => {
    if (!querySongDataRes.isSuccess) {
      return;
    }

    const songData = querySongDataRes.data;

    // Initialize form data with formatted values from API.
    const initFormData: SongForm = {
      name: formatTitle(songData.song.pdf.name),
      key: formatKey(songData.song.pdf.key),
      bpm: formatBpm(songData.song.pdf.bpm),
      time: formatTimeSignature(songData.song.pdf.time),
      tags: songData.song.tags || [],
    };

    setInitialFormData(initFormData);
    setFormData(initFormData);

    if (querySongLyricsRes.isSuccess) {
      setTransposedLyricsHtml(querySongLyricsRes.data.lyricsHtml);
    }
  }, [
    querySongDataRes.data,
    querySongDataRes.isSuccess,
    querySongLyricsRes.data,
    querySongLyricsRes.isSuccess,
  ]);

  // Effect to handle button visibility based on form changes.
  useEffect(() => {
    const changed = isFormChanged(formData, initialFormData);
    const valid = isFormValid(formData, transpositionError);

    setMainButton({
      visible: changed,
      text: formData.key === initialFormData.key ? t("save") : t("transpose"),
      enabled: changed && valid,
      loader: false,
    });
  }, [formData, initialFormData, transpositionError, t]);

  const handleKeyChange = (newKey: string) => {
    setFormData((prev: SongForm) => ({ ...prev, key: newKey }));

    if (!querySongLyricsRes.isSuccess) {
      return;
    }

    const songLyrics = querySongLyricsRes.data;

    if (newKey != initialFormData.key) {
      try {
        const transposedHtml = transposeAllText(
          songLyrics.lyricsHtml,
          initialFormData.key,
          newKey,
        );

        setTransposedLyricsHtml(transposedHtml);
      } catch (err) {
        logger.error("Error transposing lyrics", { error: err });

        setTranspositionError(true);
        setTransposedLyricsHtml(songLyrics.lyricsHtml);
      }
    } else {
      setTranspositionError(false);
      setTransposedLyricsHtml(songLyrics.lyricsHtml);
    }
    return;
  };

  const handleMainButtonClick = useCallback(async () => {
    logger.debug("updating main button handler function");

    setMainButton({ loader: true });

    if (formData.key !== initialFormData.key) {
      const stateData: StateSongData = {
        formData,
        initialFormData,
        sectionsNumber: querySongLyricsRes.data?.sectionsNumber || 1,
      };

      await navigate(
        {
          pathname: `/songs/${songId}/edit/confirm`,
          search: `?${new URLSearchParams({ messageId, chatId, userId }).toString()}`,
        },
        { state: stateData },
      );
      return;
    }
    await mutateSong(
      {
        songId: songId,
        name: formData.name,
        key: formData.key,
        bpm: formData.bpm,
        time: formData.time,
        tags: formData.tags,
        transposeSection: undefined,
        messageId: messageId,
        chatId: chatId,
        userId: userId,
      },
      {
        onSuccess: () => {
          miniApp.close();
        },
        onError: (err) => {
          logger.error("Failed to save song", { error: err });
          setMainButton({ visible: true, enabled: true, loader: false });
          Notify.failure(t("saveError"));
        },
      },
    );
  }, [
    formData,
    initialFormData,
    mutateSong,
    navigate,
    querySongLyricsRes.data?.sectionsNumber,
    messageId,
    chatId,
    songId,
    userId,
    t,
  ]);

  useEffect(() => {
    if (!querySongDataRes.isSuccess) {
      return;
    }

    logger.debug("updating main button handler");

    if (!querySongLyricsRes.isSuccess && formData.key !== initialFormData.key) {
      setMainButton({ enabled: false });
    }

    const handleMainButtonClickSync = () => {
      void handleMainButtonClick();
    };

    mainButton.onClick(handleMainButtonClickSync);

    return () => {
      logger.debug("removing old main button handler");
      setMainButton({ enabled: true, loader: false });
      mainButton.offClick(handleMainButtonClickSync);
    };
  }, [
    initialFormData.key,
    formData.key,
    handleMainButtonClick,
    querySongDataRes.isSuccess,
    querySongLyricsRes.isSuccess,
  ]);

  // todo: test.
  if (querySongDataRes.isPending || !querySongDataRes.isFetchedAfterMount) {
    return <PageLoading></PageLoading>;
  }

  if (querySongDataRes.isError) {
    return <PageError error={querySongDataRes.error}></PageError>;
  }

  // Log successful rendering
  logger.info("Rendering song page", { songId, userId });
  logger.debug("Form data", { initialFormData, formData });

  return (
    <Page back={false}>
      <List
        style={{
          // todo: check how to use safeAreaInset properly.
          marginTop: viewport.safeAreaInsetTop(),
          marginBottom: viewport.safeAreaInsetBottom(),
        }}
      >
        <EditableTitle
          value={formData.name}
          status={!isNameValid(formData.name) ? "error" : undefined}
          onChange={(value) => {
            setFormData((prev) => ({
              ...prev,
              name: value,
            }));
          }}
        />

        <TagsInput
          options={
            querySongDataRes.data.bandTags.map((tag) => ({
              value: tag,
              label: tag,
            })) || []
          }
          value={formData.tags.map((tag) => {
            return { value: tag, label: tag } as MultiselectOption;
          })}
          onChange={(newOptions: MultiselectOption[]) => {
            // Update form data
            setFormData((prev) => ({
              ...prev,
              tags: newOptions.map((option) => option.value) as string[],
            }));
          }}
        />

        <div className="flex flex-row items-center gap-2">
          <div className="flex-1">
            <KeyInput
              value={formData.key || initialFormData.key} // Using an init key here to add custom value as an option.
              status={transpositionError ? "error" : undefined}
              onChange={handleKeyChange}
            />
          </div>
          <div className="flex-1">
            <BPMInput
              value={formData.bpm}
              status={
                !isBpmValid(formData.bpm) && formData.bpm.length > 0
                  ? "error"
                  : undefined
              }
              onChange={(val) => {
                setFormData((prev) => ({
                  ...prev,
                  bpm: val,
                }));
              }}
            />
          </div>
          <div className="flex-1">
            <TimeSignatureInput
              className=""
              value={formData.time}
              status={
                !isTimeSignatureValid(formData.time) && formData.time.length > 0
                  ? "error"
                  : undefined
              }
              onChange={(val) => {
                setFormData((prev) => ({
                  ...prev,
                  time: val,
                }));
              }}
            ></TimeSignatureInput>
          </div>
        </div>

        {/*<InlineButtons mode="gray">*/}
        {/*    <InlineButtonsItem*/}
        {/*        text="Google Doc"*/}
        {/*        onClick={() => {*/}
        {/*            window.open(apiResp.data.song.pdf.webViewLink, "_blank");*/}
        {/*        }}>*/}
        {/*        <FileEarmark size={24}/>*/}
        {/*    </InlineButtonsItem>*/}
        {/*    <InlineButtonsItem text="Google Doc">*/}
        {/*        <FileEarmark size={24}/>*/}
        {/*    </InlineButtonsItem>*/}
        {/*</InlineButtons>*/}

        <Button
          before={<FileEarmarkTextFill size={"1.2em"} />}
          stretched={true}
          Component="a"
          href={querySongDataRes.data.song.pdf.webViewLink}
          mode="bezeled"
          size="m"
          target="_blank"
        >
          {t("googleDoc")}
        </Button>

        <Section className={"sect"}>
          <Text>
            <div className="p-4 font-mono text-base/6 whitespace-pre-wrap">
              {
                // todo: make more readable.
                querySongLyricsRes.isLoading ||
                !querySongLyricsRes.isFetchedAfterMount ? ( // todo: test.
                  <>{t("loadingLyrics")}</>
                ) : querySongLyricsRes.isError ? (
                  <>{t("errorLoadingLyrics")}</>
                ) : transpositionError ? (
                  <>{t("errorTransposingLyrics")}</>
                ) : (
                  <div
                    // className={e("lyrics")}
                    dangerouslySetInnerHTML={{ __html: transposedLyricsHtml }}
                  />
                )
              }
            </div>
          </Text>
        </Section>
      </List>
    </Page>
  );
};

export function useInitParams() {
  const { songId } = useParams();
  const [searchParams] = useSearchParams();
  const messageId = searchParams.get("messageId");
  const chatId = searchParams.get("chatId");
  const userId = searchParams.get("userId");
  return { songId, messageId, chatId, userId };
}
