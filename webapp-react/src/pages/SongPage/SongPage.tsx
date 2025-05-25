import { getSong, updateSong } from "@/api/webapp/songs.ts";
import {
  ReqBodyUpdateSong,
  ReqQueryParamsUpdateSong,
} from "@/api/webapp/typesReq.ts";
import { RespDataGetSong } from "@/api/webapp/typesResp.ts";
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
import { bem } from "@/css/bem.ts";
import { logger } from "@/helpers/logger";
import { typedErr } from "@/helpers/util.ts";
import {
  isBpmValid,
  isFormChanged,
  isFormValid,
  isNameValid,
  isTimeSignatureValid,
} from "@/pages/SongPage/util/formValidation.ts";
import { transposeAllText } from "@/pages/SongPage/util/transpose.ts";
import { SongForm, SongStateData } from "@/pages/SongPage/util/types.ts";
import { PageError } from "@/pages/UtilPages/PageError.tsx";
import { PageLoading } from "@/pages/UtilPages/PageLoading.tsx";
import {
  mainButton,
  miniApp,
  postEvent,
  viewport,
} from "@telegram-apps/sdk-react";
import { Button, List, Section, Text } from "@telegram-apps/telegram-ui";
import { MultiselectOption } from "@telegram-apps/telegram-ui/dist/components/Form/Multiselect/types";
import { FC, useEffect, useState } from "react";
import { FileEarmarkTextFill } from "react-bootstrap-icons";
import { useNavigate, useParams, useSearchParams } from "react-router";
import "./SongPage.css";

const [, e] = bem("song-page");

export const SongPage: FC = () => {
  const params = useParams();
  const songId = params.id;
  const [searchParams] = useSearchParams();
  const messageId = searchParams.get("messageId");
  const chatId = searchParams.get("chatId");
  const userId = searchParams.get("userId");

  const navigate = useNavigate();

  const [songData, setSongData] = useState<RespDataGetSong | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  // Form data state using the new interface
  const [formData, setFormData] = useState<SongForm>({
    name: "",
    key: "",
    bpm: "",
    time: "",
    tags: [],
  });

  // State to track initial values
  const [initialFormData, setInitialFormData] = useState<SongForm>({
    name: "",
    key: "",
    bpm: "",
    time: "",
    tags: [],
  });

  // Add a new state for transposed lyrics
  const [transposedLyricsHtml, setTransposedLyricsHtml] = useState<string>("");

  // Add a state to track transposition errors
  const [transpositionError, setTranspositionError] = useState<boolean>(false);

  useEffect(() => {
    postEvent("web_app_expand");
    postEvent("web_app_setup_swipe_behavior", { allow_vertical_swipe: false });
  }, []);

  useEffect(() => {
    const fetchSong = async () => {
      if (!songId || !userId) {
        setError(new Error("Failed to get song data: invalid request params."));
        return;
      }

      const { data, err } = await getSong(songId, userId);

      if (err || !data?.song?.lyricsHtml) {
        setError(err || new Error("Song data is empty."));
      } else if (data) {
        try {
          setSongData(data);

          // Initialize form data with formatted values from API.
          const initFormData: SongForm = {
            name: formatTitle(data.song.pdf.name),
            key: formatKey(data.song.pdf.key),
            bpm: formatBpm(data.song.pdf.bpm),
            time: formatTimeSignature(data.song.pdf.time),
            tags:
              data.song.tags?.map((tag) => ({
                value: tag,
                label: tag,
              })) || [],
          };

          setInitialFormData(initFormData);
          setFormData(initFormData);
          setTransposedLyricsHtml(data.song.lyricsHtml);
        } catch (err) {
          setError(typedErr(err));
        }
      }

      setLoading(false);
    };

    void fetchSong();
  }, [songId, userId]);

  // todo: improve, make more generic.
  // Effect to handle button visibility based on form changes
  useEffect(() => {
    const changed = isFormChanged(formData, initialFormData);
    const valid = isFormValid(formData, transpositionError);

    // Update button visibility
    mainButton.setParams({
      isVisible: changed,
      isEnabled: changed && valid,
      text: "Save",
    });

    // return () => {
    //     mainButton.setParams({
    //         isVisible: false,
    //         isEnabled: true,
    //     });
    // };
  }, [formData, initialFormData, transpositionError]);

  const handleKeyChange = (newKey: string) => {
    setFormData((prev: SongForm) => ({ ...prev, key: newKey }));

    if (!songData?.song?.lyricsHtml) {
      return;
    }

    if (newKey != initialFormData.key) {
      try {
        const transposedHtml = transposeAllText(
          songData.song.lyricsHtml,
          initialFormData.key,
          newKey,
        );

        setTransposedLyricsHtml(transposedHtml);
      } catch (err) {
        logger.error("Error transposing lyrics", { error: err });

        setTranspositionError(true);
        setTransposedLyricsHtml(songData.song.lyricsHtml);
      }
    } else {
      setTranspositionError(false);
      setTransposedLyricsHtml(songData.song.lyricsHtml);
    }
    return;
  };

  useEffect(() => {
    if (!songId || !userId || !chatId || !messageId) {
      setError(new Error("Failed to update song: invalid request params."));
      return;
    }

    const handleMainButtonClick = async () => {
      mainButton.setParams({ isLoaderVisible: true });

      if (formData.key !== initialFormData.key) {
        const stateData: SongStateData = {
          songId,
          userId,
          chatId,
          messageId,
          formData,
          initialFormData,
          sectionsNumber: songData?.song?.sectionsNumber || 1,
        };

        await navigate(`/songs/${songId}/edit/confirm`, {
          state: stateData,
        });
      } else {
        const queryParams: ReqQueryParamsUpdateSong = {
          messageId: messageId,
          chatId: chatId,
          userId: userId,
        };
        // Prepare the form data to send
        const body: ReqBodyUpdateSong = {
          name: formData.name,
          key: formData.key || "?", // todo: refactor, remove question marks.
          bpm: formData.bpm || "?",
          time: formData.time || "?",
          tags: formData.tags.map((tag) => String(tag.value)),
        };

        const err = await updateSong(songId, queryParams, body);
        if (err) {
          logger.error("Failed to save song", { error: err });
          setError(err);
          mainButton.setParams({ isVisible: false });
        } else {
          miniApp.close();
        }
      }
    };

    const handleMainButtonClickSync = () => {
      void handleMainButtonClick();
    };

    mainButton.onClick(handleMainButtonClickSync);

    return () => {
      mainButton.setParams({ isLoaderVisible: false });
      mainButton.offClick(handleMainButtonClickSync);
    };
  }, [
    songId,
    userId,
    messageId,
    chatId,
    formData,
    navigate,
    initialFormData,
    songData?.song?.sectionsNumber,
  ]);

  // Show loading state
  if (loading) {
    return <PageLoading></PageLoading>;
  }

  // Show error state
  if (error || !songData) {
    return (
      <PageError error={error || new Error("Empty api response")}></PageError>
    );
  }

  // Log successful rendering
  logger.info("Rendering song page", { songId, userId });

  return (
    <Page back={false}>
      <List
        className={e("body")}
        // todo: check how to use safeAreaInset properly.
        style={{
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
            songData.bandTags.map((tag) => ({
              value: tag,
              label: tag,
            })) || []
          }
          value={formData.tags}
          onChange={(newOptions: MultiselectOption[]) => {
            // Update form data
            setFormData((prev) => ({
              ...prev,
              tags: newOptions,
            }));
          }}
        />

        <div className={e("inputs-row")}>
          <KeyInput
            value={initialFormData.key} // Using init key here to add custom value as option.
            status={transpositionError ? "error" : undefined}
            onChange={handleKeyChange}
          />
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
          <TimeSignatureInput
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
          href={songData.song.pdf.webViewLink}
          mode="bezeled"
          size="m"
          target="_blank"
        >
          Google Doc
        </Button>

        <Section>
          <Text>
            <div
              className={e("lyrics")}
              dangerouslySetInnerHTML={{
                __html: transposedLyricsHtml || songData.song.lyricsHtml,
              }}
            />
          </Text>
        </Section>
      </List>
    </Page>
  );
};
