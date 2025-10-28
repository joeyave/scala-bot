import { getTags } from "@/api/webapp/tags.ts";
import { BPMInput } from "@/components/BPMInput/BPMInput.tsx";
import { KeyInput } from "@/components/KeyInput/KeyInput.tsx";
import { Page } from "@/components/Page.tsx";
import { TagsInput } from "@/components/TagsInput/TagsInput.tsx";
import { TimeSignatureInput } from "@/components/TimeSignatureInput/TimeSignatureInput.tsx";
import { logger } from "@/helpers/logger";
import { setMainButton } from "@/helpers/mainButton.ts";
import {
  isBpmValid,
  isNameValid,
  isTimeSignatureValid,
} from "@/pages/SongPage/util/formValidation.ts";
import { SongForm } from "@/pages/SongPage/util/types.ts";
import { useSuspenseQuery } from "@tanstack/react-query";
import { mainButton, miniApp, postEvent, viewport } from "@tma.js/sdk-react";
import { List, Textarea } from "@telegram-apps/telegram-ui";
import { MultiselectOption } from "@telegram-apps/telegram-ui/dist/components/Form/Multiselect/types";
import { FC, useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router";

const CreateSongPage: FC = () => {
  const { t } = useTranslation();

  const [searchParams] = useSearchParams();
  const bandId = searchParams.get("bandId");

  if (!bandId) {
    throw new Error("Failed to get song page: invalid request params.");
  }

  const queryTagsRes = useSuspenseQuery({
    queryKey: ["tags", bandId],
    queryFn: async () => {
      const data = await getTags(bandId);
      if (!data) {
        throw new Error("Failed to get tags");
      }
      return data;
    },
  });

  // Form data state.
  const [formData, setFormData] = useState<SongForm>({
    name: "",
    key: "",
    bpm: "",
    time: "",
    tags: [],
  });
  const lyricsRef = useRef<string>("");

  useEffect(() => {
    postEvent("web_app_expand");
    postEvent("web_app_setup_swipe_behavior", { allow_vertical_swipe: false });
  }, []);

  useEffect(() => {
    const changed = !!formData.name;
    const valid =
      isNameValid(formData.name) &&
      isBpmValid(formData.bpm) &&
      isTimeSignatureValid(formData.time);

    logger.debug("effect changed", { changed, valid });

    setMainButton({
      visible: changed,
      text: t("save"),
      enabled: changed && valid,
      loader: false,
    });
  }, [formData.name, formData.bpm, formData.time, t]);

  const handleMainButtonClick = useCallback(() => {
    logger.debug("updating main button handler function");

    setMainButton({ loader: true });

    postEvent("web_app_data_send", {
      data: JSON.stringify({
        name: formData.name,
        key: formData.key,
        bpm: formData.bpm,
        time: formData.time,
        tags: formData.tags,
        lyrics: lyricsRef.current,
      }),
    });
    miniApp.close();
  }, [formData]);

  useEffect(() => {
    logger.debug("updating main button handler");

    mainButton.onClick(handleMainButtonClick);

    return () => {
      logger.debug("removing old main button handler");

      setMainButton({ enabled: true, loader: false });
      mainButton.offClick(handleMainButtonClick);
    };
  }, [handleMainButtonClick]);

  // Log successful rendering
  logger.info("Rendering song creation page");
  logger.debug("Form data", { formData });

  return (
    <Page back={false}>
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
            <Textarea
              placeholder={t("namePlaceholder")}
              status={!isNameValid(formData.name) ? "error" : undefined}
              onChange={(value) => {
                setFormData((prev) => ({
                  ...prev,
                  name: value.target.value,
                }));
              }}
            ></Textarea>
            <TagsInput
              options={
                queryTagsRes.data?.tags.map((tag) => ({
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
                  value={t("keyPlaceholder")}
                  onChange={(val) => {
                    setFormData((prev) => ({
                      ...prev,
                      key: val,
                    }));
                  }}
                />
              </div>
              <div className="flex-1">
                <BPMInput
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
                  status={
                    !isTimeSignatureValid(formData.time) &&
                    formData.time.length > 0
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
          </List>
        </div>

        <div className="flex flex-1 flex-col">
          <List className="flex flex-1 flex-col pt-0 pb-0 **:h-full **:not-placeholder-shown:!font-mono **:not-placeholder-shown:!text-base/6">
            <Textarea
              placeholder={t("lyricsPlaceholder")}
              onChange={(e) => {
                lyricsRef.current = e.target.value;
              }}
            ></Textarea>
          </List>
        </div>
      </div>
    </Page>
  );
};

export default CreateSongPage;
