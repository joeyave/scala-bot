import {
  getEventData,
  getEventFreqNames,
  updateEvent,
} from "@/api/webapp/events.ts";
import { ReqBodyUpdateEvent } from "@/api/webapp/typesReq.ts";
import { EditableTitle } from "@/components/EditableTitle/EditableTitle.tsx";
import { Page } from "@/components/Page.tsx";
import { logger } from "@/helpers/logger.ts";
import { setMainButton } from "@/helpers/mainButton.ts";
import {
  isFormChanged,
  isFormValid,
} from "@/pages/EventPage/util/formValidation.ts";
import { EventForm, Song } from "@/pages/EventPage/util/types.ts";
import { CalendarIcon } from "@heroicons/react/20/solid";
import { useMutation, useSuspenseQuery } from "@tanstack/react-query";
import { IconButton, Input, List, Textarea } from "@telegram-apps/telegram-ui";
import { mainButton, miniApp, postEvent } from "@tma.js/sdk-react";
import { Notify } from "notiflix";
import { FC, useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useParams, useSearchParams } from "react-router";
import { SetlistSection } from "@/components/Setlist/SetlistSection.tsx";

interface EventMutationData {
  eventId: string;
  name: string;
  date: Date;
  songIds: string[];
  notes: string;
  messageId: string;
  chatId: string;
  userId: string;
}

export function useEventMutation() {
  const { mutateAsync: mutateEvent } = useMutation({
    mutationFn: async (d: EventMutationData) => {
      const queryParams = {
        messageId: d.messageId,
        chatId: d.chatId,
        userId: d.userId,
      };

      const body: ReqBodyUpdateEvent = {
        name: d.name,
        date: d.date.toISOString(),
        songIds: d.songIds,
        notes: d.notes,
      };

      return await updateEvent(d.eventId, queryParams, body);
    },
  });
  return mutateEvent;
}

const CreateEventPage: FC = () => {
  const { t } = useTranslation();

  const { eventId, messageId, chatId, userId } = useInitParams();

  if (!eventId || !messageId || !chatId || !userId) {
    throw new Error("Failed to get event page: invalid request params.");
  }

  const queryEventRes = useSuspenseQuery({
    queryKey: ["event", eventId],
    queryFn: async () => {
      const data = await getEventData(eventId);
      if (!data) {
        throw new Error("Failed to get event data");
      }
      return data;
    },
  });

  const queryFreqNamesRes = useSuspenseQuery({
    queryKey: ["freqNames", queryEventRes.data.event.bandId],
    queryFn: async () => {
      const data = await getEventFreqNames(queryEventRes.data.event.bandId);
      if (!data) {
        throw new Error("Failed to get event freq names");
      }
      return data;
    },
  });

  const mutateEvent = useEventMutation();

  const init: EventForm = {
    name: queryEventRes.data.event.name,
    date: queryEventRes.data.event.time,
    setlist: queryEventRes.data.event.songs.map((song) => {
      const s: Song = {
        id: song.id,
        name: `${song.pdf.name}`,
        key: song.pdf.key,
        bpm: song.pdf.bpm,
        time: song.pdf.time,
      };
      return s;
    }),
    notes: queryEventRes.data.event.notes,
  };

  // Form data state.
  const [initFormData] = useState<EventForm>(init);

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

  const handleMainButtonClick = useCallback(async () => {
    setMainButton({ loader: true });

    const data: EventMutationData = {
      eventId: eventId,
      name: formData.name,
      date: new Date(formData.date),
      songIds: formData.setlist.map((song) => song.id),
      notes: formData.notes,
      messageId: messageId,
      chatId: chatId,
      userId: userId,
    };

    await mutateEvent(data, {
      onSuccess: () => {
        miniApp.close();
      },
      onError: (err) => {
        logger.error("Failed to save event", { error: err });
        setMainButton({ visible: true, enabled: true, loader: false });
        Notify.failure(t("saveErrorEvent"));
      },
    });
  }, [
    chatId,
    eventId,
    formData.date,
    formData.name,
    formData.notes,
    formData.setlist,
    messageId,
    mutateEvent,
    t,
    userId,
  ]);

  useEffect(() => {
    const handleMainButtonClickSync = () => {
      void handleMainButtonClick();
    };

    mainButton.onClick(handleMainButtonClickSync);

    return () => {
      // setMainButton({ enabled: true, loader: false });
      mainButton.offClick(handleMainButtonClickSync);
    };
  }, [handleMainButtonClick]);

  return (
    <Page back={false}>
      <div
        className="flex h-screen flex-col"
        // style={{
        //   // todo: check how to use safeAreaInset properly.
        //   marginTop: viewport.safeAreaInsetTop(),
        //   marginBottom: viewport.safeAreaInsetBottom(),
        // }}
      >
        <div className="flex-none">
          <List>
            <EditableTitle
              // list="suggestions"
              // placeholder={t("namePlaceholder")}
              value={formData.name}
              // status={!isNameValid(formData.name) ? "error" : undefined}
              onChange={(value) => {
                setFormData((prev) => ({
                  ...prev,
                  name: value,
                }));
              }}
            ></EditableTitle>

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
              type="date"
              className="w-full rounded-xl bg-[var(--tg-theme-section-bg-color)] px-3 py-2 text-base font-medium"
              value={new Date(formData.date).toISOString().split("T")[0]}
              // status={!isDateValid(formData.date) ? "error" : undefined}
              onChange={(val) => {
                setFormData((prev) => ({
                  ...prev,
                  date: val.target.value,
                }));
              }}
            />

            <SetlistSection
              driveFolderId={queryEventRes.data.event.band.driveFolderId}
              archiveFolderId={queryEventRes.data.event.band.archiveFolderId}
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

export function useInitParams() {
  const { eventId } = useParams();
  const [searchParams] = useSearchParams();
  const messageId = searchParams.get("messageId");
  const chatId = searchParams.get("chatId");
  const userId = searchParams.get("userId");
  return { eventId, messageId, chatId, userId };
}

export default CreateEventPage;
