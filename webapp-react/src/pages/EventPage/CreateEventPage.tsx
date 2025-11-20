import { getEventFreqNames } from "@/api/webapp/events.ts";
import { Page } from "@/components/Page.tsx";
import { SetlistSection } from "@/components/Setlist/SetlistSection.tsx";
import { setMainButton } from "@/helpers/mainButton.ts";
import {
  isFormChanged,
  isFormValid,
} from "@/pages/EventPage/util/formValidation.ts";
import { getLocalDateTimeString } from "@/pages/EventPage/util/helpers.ts";
import { EventForm } from "@/pages/EventPage/util/types.ts";
import { CalendarIcon } from "@heroicons/react/20/solid";
import { useSuspenseQuery } from "@tanstack/react-query";
import { IconButton, Input, List, Textarea } from "@telegram-apps/telegram-ui";
import { mainButton, miniApp, postEvent } from "@tma.js/sdk-react";
import { FC, useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router";

const CreateEventPage: FC = () => {
  const { t } = useTranslation();

  const [searchParams] = useSearchParams();
  const bandId = searchParams.get("bandId");
  const driveFolderId = searchParams.get("driveFolderId");
  const archiveFolderId = searchParams.get("archiveFolderId");

  if (!bandId || !driveFolderId) {
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
    },
  });

  // Form data state.
  const [initFormData] = useState<EventForm>({
    name: "",
    // date: getLocalDateTimeString(new Date()),
    date: getLocalDateTimeString(new Date()).split("T")[0],
    setlist: [],
    notes: "",
  });

  const [formData, setFormData] = useState<EventForm>(initFormData);

  useEffect(() => {
    postEvent("web_app_expand");
    postEvent("web_app_setup_swipe_behavior", { allow_vertical_swipe: false });
  }, []);

  useEffect(() => {
    const changed = isFormChanged(formData, initFormData);
    const valid = isFormValid(formData);

    setMainButton({
      visible: changed,
      text: t("save"),
      enabled: changed && valid,
      loader: false,
    });
  }, [formData, initFormData, t]);

  const handleMainButtonClick = useCallback(() => {
    setMainButton({ loader: true });
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    const utc = new Date(formData.date).toISOString();
    postEvent("web_app_data_send", {
      data: JSON.stringify({
        name: formData.name,
        time: utc,
        timezone: timezone,
        songIds: formData.setlist.map((song) => song.id),
        notes: formData.notes,
      }),
    });
    miniApp.close();
  }, [formData]);

  useEffect(() => {
    mainButton.onClick(handleMainButtonClick);

    return () => {
      // setMainButton({ enabled: true, loader: false });
      mainButton.offClick(handleMainButtonClick);
    };
  }, [handleMainButtonClick]);

  return (
    <Page back={false}>
      <div
        className="flex h-screen flex-col overflow-hidden"
        style={{}}
        // style={{
        //   // todo: check how to use safeAreaInset properly.
        //   marginTop: viewport.safeAreaInsetTop(),
        //   marginBottom: viewport.safeAreaInsetBottom(),
        // }}
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
                  name: value.target.value,
                }));
              }}
            ></Input>

            <datalist id="suggestions">
              {queryFreqNamesRes.data?.names.map(
                (name: string, index: number) => (
                  <option key={index} value={name}></option>
                ),
              ) || []}
            </datalist>

            <Input
              after={
                <IconButton mode="plain" size="s">
                  <CalendarIcon className="h-5 w-5 text-[var(--tg-theme-accent-text-color)]"></CalendarIcon>
                </IconButton>
              }
              // type="datetime-local"
              type="date"
              className="w-full rounded-xl bg-[var(--tg-theme-section-bg-color)] px-3 py-2 text-base font-medium"
              value={formData.date}
              // status={!isDateValid(formData.date) ? "error" : undefined}
              onChange={(val) => {
                setFormData((prev) => ({
                  ...prev,
                  date: val.target.value,
                }));
              }}
            />

            <SetlistSection
              driveFolderId={driveFolderId}
              archiveFolderId={archiveFolderId}
              songs={formData.setlist}
              onAddSong={(song) => {
                setFormData((prev) => ({
                  ...prev,
                  setlist: [...prev.setlist, song],
                }));
              }}
              onRemove={(songToRemove) => {
                setFormData((prev) => ({
                  ...prev,
                  setlist: prev.setlist.filter((s) => s.id !== songToRemove.id),
                }));
              }}
              onReorder={(newSetlist) => {
                setFormData((prev) => ({
                  ...prev,
                  setlist: newSetlist,
                }));
              }}
            />

            <Textarea
              header={t("notes")}
              placeholder={t("notesPlaceholder")}
              value={formData.notes}
              onChange={(val) => {
                setFormData((prev) => ({
                  ...prev,
                  notes: val.target.value,
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
