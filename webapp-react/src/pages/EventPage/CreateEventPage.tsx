import { Page } from "@/components/Page.tsx";
import { logger } from "@/helpers/logger";
import { setMainButton } from "@/helpers/mainButton.ts";
import { isFormChanged, isFormValid } from "@/pages/EventPage/util/formValidation.ts";
import { EventForm } from "@/pages/EventPage/util/types.ts";
import { Input, List } from "@telegram-apps/telegram-ui";
import { mainButton, miniApp, postEvent, viewport } from "@tma.js/sdk-react";
import { FC, useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getEventFreqNames } from "@/api/webapp/events.ts";

const CreateEventPage: FC = () => {
  const { t } = useTranslation();

  const [searchParams] = useSearchParams();
  const bandId = searchParams.get("bandId");

  if (!bandId) {
    throw new Error("Failed to get event page: invalid request params.");
  }

  const queryFreqNamesRes = useSuspenseQuery({
    queryKey: ["freqNames", bandId],
    queryFn: async () => {
      const data = await getEventFreqNames(bandId);
      if (!data) {
        throw new Error("Failed to get event freq names");
      }
      return data;
    }
  });

  // Form data state.
  const [initFormData] = useState<EventForm>({
    name: "",
    date: new Date().toISOString().split("T")[0]
  });

  const [formData, setFormData] = useState<EventForm>(initFormData);

  useEffect(() => {
    postEvent("web_app_expand");
    postEvent("web_app_setup_swipe_behavior", { allow_vertical_swipe: false });
  }, []);

  useEffect(() => {
    const changed = isFormChanged(formData, initFormData);
    const valid = isFormValid(formData);

    logger.debug("effect changed", { changed, valid });

    setMainButton({
      visible: changed,
      text: t("save"),
      enabled: changed && valid,
      loader: false
    });
  }, [formData, initFormData, t]);

  const handleMainButtonClick = useCallback(() => {
    logger.debug("updating main button handler function");

    setMainButton({ loader: true });

    postEvent("web_app_data_send", {
      data: JSON.stringify({
        name: formData.name,
        time: new Date(formData.date)
      })
    });
    miniApp.close();
  }, [formData]);

  useEffect(() => {
    logger.debug("updating main button handler");

    mainButton.onClick(handleMainButtonClick);

    return () => {
      logger.debug("removing old main button handler");

      // setMainButton({ enabled: true, loader: false });
      mainButton.offClick(handleMainButtonClick);
    };
  }, [handleMainButtonClick]);

  // Log successful rendering
  logger.info("Rendering event creation page");
  logger.debug("Form data", { formData });

  return (
    <Page back={false}>
      <div
        className="flex h-screen flex-col overflow-hidden"
        style={{
          // todo: check how to use safeAreaInset properly.
          marginTop: viewport.safeAreaInsetTop(),
          marginBottom: viewport.safeAreaInsetBottom()
        }}
      >
        <div className="flex-none">
          <List>
            <Input
              list="suggestions"
              placeholder={t("namePlaceholder")}
              // status={!isNameValid(formData.name) ? "error" : undefined}
              onChange={(value) => {
                setFormData((prev) => ({
                  ...prev,
                  name: value.target.value
                }));
              }}
            ></Input>

            <datalist id="suggestions">
              {
                queryFreqNamesRes.data?.names.map((name: string, index: number) => (
                    <option key={index} value={name}></option>
                  )
                ) || []
              }
            </datalist>

            <Input
              type="date"
              value={formData.date}
              // status={!isDateValid(formData.date) ? "error" : undefined}
              onChange={(val) => {
                setFormData((prev) => ({
                  ...prev,
                  date: val.target.value
                }));
              }}
            />
          </List>
        </div>
      </div>
    </Page>
  );
};

export default CreateEventPage;